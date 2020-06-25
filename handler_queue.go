package main

import (
    "fmt"
    "html/template"
    "net/http"
    "path"
    "github.com/op/go-logging"
)

type tplRequest struct {
    QueueRequest *QueueRequest
    Url string
    WidthTiles int
    HeightTiles int
    WidthPx int
    HeightPx int
}

type HandlerQueue struct {
    log *logging.Logger
    queue *Queue
}

func (h *HandlerQueue) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    var err error

    h.log.Debugf("Processing request, path: %s", r.URL.Path)

    // check http method, GET is required
    if r.Method != http.MethodGet {
        WriteErrorResponse(w, http.StatusMethodNotAllowed, fmt.Errorf("Only GET method is allowed"))
        return
    }

    // check path
    if r.URL.Path != "/queue" {
        h.log.Warningf("Ignoring request to path %s", r.URL.Path)
        WriteErrorResponse(w, http.StatusNotFound, fmt.Errorf("Only /queue path is supported"))
        return
    }

    // get id of the request
    request := r.URL.Query().Get("request")
    if len(request) != 0 {
        h.log.Debugf("Request to be shown: %s", request)
    }

    // get current list of all queued requests
    requests, err := h.queue.GetRequests()
    if err != nil {
        WriteErrorResponse(w, 500, fmt.Errorf("Fetching queue request failed: %s", err))
    }

    var tplData []tplRequest
    for _, r := range requests {

        // if request id was specify, filter requests
        if len(request) != 0 {
            if request != r.Id {
                continue
            }
        }

        tr := tplRequest{
            r,
            path.Join(h.queue.dir, GetImageFileName(r.Id)),
            r.Params.XMax - r.Params.XMin,
            r.Params.YMax - r.Params.YMin,
            (r.Params.XMax - r.Params.XMin) * r.Params.Scale,
            (r.Params.YMax - r.Params.YMin) * r.Params.Scale,
        }

        tplData = append(tplData, tr)
    }

    tmpl := template.Must(template.ParseFiles("html/base.html", "html/queue.html"))

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    err = tmpl.Execute(w, tplData)
    if err != nil {
        w.Write([]byte(fmt.Sprintf("Rendering error: %s", err)))
    }
}
