package main

import (
    "encoding/json"
    "fmt"
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

type QueueRequestMeta struct {
    Id string
    Params InputParams
    State string
    Created int64
}

type Queue struct {
    log *logging.Logger
    dir string
}

// constructor
func NewQueue(log *logging.Logger, dir string) (*Queue, error) {

    if _, err := os.Stat(dir); os.IsNotExist(err) {
        log.Infof("Creating queue directory: %s", dir)
        err = os.MkdirAll(dir, os.ModePerm)
        if err != nil {
            return nil, err
        }
    }

    q := &Queue{log: log, dir: dir}

    go q.Monitor()

    return q, nil
}

func (q *Queue) Monitor() {
    for {
        q.log.Debugf("Queue monitor iteration")

        // read all files in queue
        files, err := ioutil.ReadDir(q.dir)
        if err != nil {
            q.log.Errorf("Cannot read content of directory %s: %s", q.dir, err)
        }

        for _, file := range files {
            // skip directories
            if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
                continue
            }

            q.log.Debugf("file %s", file.Name())
        }

        time.Sleep(5000 * time.Millisecond)
    }
}


func (q *Queue) Enqueue(ip *InputParams) error {

    // generate unique id
    id := UniqueId()

    q.log.Debugf("New request in queue: %s", id)

    // create new queue record
    requestMeta:= QueueRequestMeta{id, *ip, QUEUE_REQUEST_STATE_NEW, time.Now().Unix()}

    // store queue record attributes to json file
    requestMetaJson, err := json.MarshalIndent(requestMeta, "", " ")
    if err != nil {
        return err
    }

    err = ioutil.WriteFile(path.Join(q.dir, id + ".json"), requestMetaJson, 0644)
    return err
}

func (q *Queue) Process(ip *InputParams) {

    // create final image (canvas)
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

    // save to file
    out, err := os.Create("./output2.png")
    if err != nil {
        erro := fmt.Errorf("Error creating file: %s", err)
        q.log.Errorf("%s", erro)
        //WriteErrorResponse(w, 500., erro)
        return
    }

    err = png.Encode(out, final)
    if err != nil {
        erro := fmt.Errorf("Error encoding PNG to file: %s", err)
        q.log.Errorf("%s", erro)
        //WriteErrorResponse(w, 500., erro)
        return
    }
}

