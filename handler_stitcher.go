package main

import (
    "image/draw"
    "fmt"
    "html/template"
    "image"
    _ "image/jpeg"
    "image/png"
    "net/http"
    "os"
    "github.com/op/go-logging"
)

type HandlerStitcher struct {
    log *logging.Logger
    providers map[string]Provider
    queue *Queue
}

func (h *HandlerStitcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    var err error

    h.log.Debugf("Processing request, path: %s", r.URL.Path)

    // check http method, GET is required
    if r.Method != http.MethodGet {
        WriteErrorResponse(w, http.StatusMethodNotAllowed, fmt.Errorf("Only GET method is allowed"))
        return
    }

    // check http method, GET is required
    if r.URL.Path != "/stitcher" {
        h.log.Warningf("Ignoring request to path %s", r.URL.Path)
        WriteErrorResponse(w, http.StatusNotFound, fmt.Errorf("Only /stitcher path is supported"))
        return
    }

    // get request context to access params from upper handler
    ctx := r.Context()

    // get input parameters
    ip := ctx.Value("ip").(*InputParams)

    err = h.queue.Enqueue(ip)
    if err != nil {
        e := fmt.Errorf("Cannot enqueue: %s", err)
        h.log.Error(err)
        WriteErrorResponse(w, 500, e)
    }

    // create final image (canvas)
    h.log.Debugf("Input params: %v", ip);
    finalRect := image.Rectangle{image.Point{0, 0}, image.Point{(ip.XMax - ip.XMin + 1) * ip.Scale, (ip.YMax - ip.YMin + 1) * ip.Scale}}
    h.log.Debugf("Final image size: %v", finalRect);
    final := image.NewRGBA(finalRect)

    // get tiles for current set of parameters
    tiles := ip.Provider.getTiles(ip.XMin, ip.YMin, ip.XMax, ip.YMax, ip.Zoom, ip.Scale)

    // loop through all tiles
    for i := 0; i < len(*tiles); i++ {

        t := (*tiles)[i]

        h.log.Debugf("Fetching tile %s (%d:%d)", t.Url, t.Left, t.Top);
        res, err := http.Get(t.Url)
        if err != nil || res.StatusCode != 200 {
            // silently skip this tile
            //WriteErrorResponse(w, http., fmt.Errorf("Only GET method is allowed"))
            h.log.Warningf("Fetching tile %s failed", t.Url);
            continue
        }
        defer res.Body.Close()

        m, format, err := image.Decode(res.Body)
        if err != nil {
            h.log.Warningf("Decoding tile image %s failed: %s", t.Url, err);
            continue
        }

        h.log.Debugf("Decoding tile image passed (format: %s)", format);

        // put fetched tile image at proper place in final image
        finalPoint := image.Point{t.Left * ip.Scale, t.Top * ip.Scale}
        finalRect := m.Bounds().Add(finalPoint)

        h.log.Debugf("Putting tile at %v", finalRect)
        draw.Draw(final, finalRect, m, image.Point{}, draw.Over)
    }

    // save to file
    out, err := os.Create("./output2.png")
    if err != nil {
        erro := fmt.Errorf("Error creating file: %s", err)
        h.log.Errorf("%s", erro)
        WriteErrorResponse(w, 500., erro)
        return
    }

    err = png.Encode(out, final)
    if err != nil {
        erro := fmt.Errorf("Error encoding PNG to file: %s", err)
        h.log.Errorf("%s", erro)
        WriteErrorResponse(w, 500., erro)
        return
    }

    data := ip.Provider.Name

    tmpl := template.Must(template.ParseFiles("html/base.html", "html/stitcher.html"))

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    err = tmpl.Execute(w, data)
    if err != nil {
        w.Write([]byte(fmt.Sprintf("Rendering error: %s", err)))
    }
}
