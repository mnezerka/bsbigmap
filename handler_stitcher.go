package main

import (
    "fmt"
    "html/template"
    "net/http"
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

    request, err := h.queue.Enqueue(ip)
    if err != nil {
        e := fmt.Errorf("Cannot enqueue: %s", err)
        h.log.Error(err)
        WriteErrorResponse(w, 500, e)
    }

    tmpl := template.Must(template.ParseFiles("html/base.html", "html/stitcher.html"))

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    err = tmpl.Execute(w, request)
    if err != nil {
        w.Write([]byte(fmt.Sprintf("Rendering error: %s", err)))
    }
}
