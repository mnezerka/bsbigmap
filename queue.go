package main

import (
    "encoding/json"
    "io/ioutil"
    "os"
    "image"
    "image/draw"
    "image/png"
    "net/http"
    "path"
    "path/filepath"
    "time"
    "github.com/op/go-logging"
)

const QUEUE_REQUEST_STATE_NEW = "new"
const QUEUE_REQUEST_STATE_DONE = "done"
const QUEUE_REQUEST_STATE_ERROR = "error"

type QueueRequest struct {
    Id string
    Params InputParams
    State string
    Created int64
}

type Queue struct {
    log *logging.Logger
    dir string
    validity time.Duration
    interval time.Duration
}

// constructor
func NewQueue(log *logging.Logger, dir string, validity, interval time.Duration) (*Queue, error) {

    log.Debugf("Interval is %v", interval)

    if _, err := os.Stat(dir); os.IsNotExist(err) {
        log.Infof("Creating queue directory: %s", dir)
        err = os.MkdirAll(dir, os.ModePerm)
        if err != nil {
            return nil, err
        }
    }

    q := &Queue{log: log, dir: dir, validity: validity, interval: interval}

    go q.monitor()

    return q, nil
}

func (q *Queue) GetRequests() ([]*QueueRequest, error) {

    var requests []*QueueRequest

    // read all files in queue directory
    files, err := ioutil.ReadDir(q.dir)
    if err != nil {
        q.log.Errorf("Cannot read content of directory %s: %s", q.dir, err)
    }

    for _, file := range files {

        // skip directories and no-json files
        if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
            continue
        }

        // read content of json file
        filePath := filepath.Join(q.dir, file.Name())
        fileIn, err := ioutil.ReadFile(filePath)
        if err != nil {
            q.log.Warningf("Cannot read content of file %s: %s", filePath, err)
            continue
        }
        request := QueueRequest{}
        err = json.Unmarshal([]byte(fileIn), &request)
        if err != nil {
            q.log.Warningf("Cannot parse json file %s: %s", filePath, err)
        }

        requests = append(requests, &request)
    }

    return requests, nil
}

func (q *Queue) monitor() {

    // infinite loop that is exited on closing of this routine
    // by program exiting - no sync (channel) with main routine
    for {
        q.log.Debugf("Queue monitor iteration")

        // read all files in queue directory
        files, err := ioutil.ReadDir(q.dir)
        if err != nil {
            q.log.Errorf("Cannot read content of directory %s: %s", q.dir, err)
        }

        for _, file := range files {

            // skip directories and no-json files
            if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
                continue
            }

            q.log.Debugf("- file %s", file.Name())

            // read content of json file
            filePath := filepath.Join(q.dir, file.Name())
            fileIn, err := ioutil.ReadFile(filePath)
            if err != nil {
                q.log.Warningf("Cannot read content of file %s: %s", filePath, err)
                continue
            }
            request := QueueRequest{}
            err = json.Unmarshal([]byte(fileIn), &request)
            if err != nil {
                q.log.Warningf("Cannot parse json file %s: %s", filePath, err)
            }

            // check request validity and delete expired requests
            created := time.Unix(request.Created, 0)
            if time.Since(created) > q.validity {
                q.removeRequest(&request)
                continue
            }

            if request.State == QUEUE_REQUEST_STATE_NEW {
                q.processRequest(&request)
            }
        }

        time.Sleep(q.interval)
    }
}

func GetJsonFileName(id string) string {
    return id + ".json"
}

func GetImageFileName(id string) string {
    return id + ".png"
}

func (q *Queue) processRequest(request *QueueRequest) {
    q.log.Debugf("Processing request %s", request.Id)

    err := q.generateRequestImage(request)

    if err != nil {
        q.log.Errorf("%s", err)
        request.State = QUEUE_REQUEST_STATE_ERROR
    } else {
        request.State = QUEUE_REQUEST_STATE_DONE
    }

    // update json file of new status
    if err := q.writeRequestJson(request); err != nil {
        q.log.Debugf("Cannot write to json file: %s", err)
    }
}


func (q *Queue) generateRequestImage(request *QueueRequest) error {
    q.log.Debugf("Generating image for request %s", request.Id)

    ip := request.Params

    q.log.Debugf("Input params: %v", ip);
    finalRect := image.Rectangle{image.Point{0, 0}, image.Point{(ip.XMax - ip.XMin + 1) * ip.Scale, (ip.YMax - ip.YMin + 1) * ip.Scale}}
    q.log.Debugf("Final image size: %v", finalRect);
    final := image.NewRGBA(finalRect)

    // get tiles for current set of parameters
    tiles := ip.Provider.getTiles(ip.XMin, ip.YMin, ip.XMax, ip.YMax, ip.Zoom, ip.Scale)

    // loop through all tiles
    for i := 0; i < len(*tiles); i++ {

        t := (*tiles)[i]

        q.log.Debugf("Fetching tile %s (%d:%d)", t.Url, t.Left, t.Top);
        res, err := http.Get(t.Url)
        if err != nil || res.StatusCode != 200 {
            // silently skip this tile
            //WriteErrorResponse(w, http., fmt.Errorf("Only GET method is allowed"))
            q.log.Warningf("Fetching tile %s failed", t.Url);
            continue
        }
        defer res.Body.Close()

        m, format, err := image.Decode(res.Body)
        if err != nil {
            q.log.Warningf("Decoding tile image %s failed: %s", t.Url, err);
            continue
        }

        q.log.Debugf("Decoding tile image passed (format: %s)", format);

        // put fetched tile image at proper place in final image
        finalPoint := image.Point{t.Left * ip.Scale, t.Top * ip.Scale}
        finalRect := m.Bounds().Add(finalPoint)

        q.log.Debugf("Putting tile at %v", finalRect)
        draw.Draw(final, finalRect, m, image.Point{}, draw.Over)
    }

    // save to image file (PNG)
    out, err := os.Create(filepath.Join(q.dir, GetImageFileName(request.Id)))
    if err != nil {
        return err
    }

    err = png.Encode(out, final)
    if err != nil {
        return err
    }

    return nil
}

func (q *Queue) removeRequest(request *QueueRequest) {
    q.log.Debugf("Remove request %v", request.Id)

    // delete generated png file
    if err := os.Remove(path.Join(q.dir, GetImageFileName(request.Id))); err != nil {
        q.log.Warningf("Remove image file for request %s failed: %s", request.Id, err)
    }

    // delete json file
    if err := os.Remove(path.Join(q.dir, GetJsonFileName(request.Id))); err != nil {
        q.log.Warningf("Remove json file for request %s failed: %s", request.Id, err)
    }
}

func (q Queue) writeRequestJson(request *QueueRequest) error {

    // store queue record attributes to json file
    requestJson, err := json.MarshalIndent(request, "", " ")
    if err != nil {
        return err
    }

    err = ioutil.WriteFile(path.Join(q.dir, GetJsonFileName(request.Id)), requestJson, 0644)
    return err
}

func (q *Queue) Enqueue(ip *InputParams) (*QueueRequest, error) {

    // generate unique id
    id := UniqueId()

    q.log.Debugf("New request in queue: %s", id)

    // create new queue record
    request:= QueueRequest{id, *ip, QUEUE_REQUEST_STATE_NEW, time.Now().Unix()}

    if err := q.writeRequestJson(&request); err != nil {
        return nil, err
    }
    return &request, nil
}

    
