package main

import (
    "fmt"
    "html/template"
    ttemplate "text/template"
    "net/http"
    "github.com/op/go-logging"
)

type HandlerRoot struct {
    log *logging.Logger
    providers map[string]Provider
}

func (h *HandlerRoot) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    var err error

    h.log.Debugf("Processing request, path: %s", r.URL.Path)

    // check http method, GET is required
    if r.Method != http.MethodGet {
        WriteErrorResponse(w, http.StatusMethodNotAllowed, fmt.Errorf("Only GET method is allowed"))
        return
    }

    // check path, only root is allowed
    /*
    if r.URL.Path != "/" {
        h.log.Warningf("Ignoring request to path %s", r.URL.Path)
        WriteErrorResponse(w, http.StatusNotFound, fmt.Errorf("Only root path is supported"))
        return
    }
    */

    if r.URL.Path == "/map.js" {
        data := h.providers
        tmpl := ttemplate.Must(ttemplate.ParseFiles("js/map.js"))
        w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
        err = tmpl.Execute(w, data)
        if err != nil {
            w.Write([]byte(fmt.Sprintf("Rendering error: %s", err)))
        }
    } else if r.URL.Path == "/" {
        tmpl := template.Must(template.ParseFiles("html/base.html", "html/index.html"))
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        err = tmpl.Execute(w, nil)
        if err != nil {
            w.Write([]byte(fmt.Sprintf("Rendering error: %s", err)))
        }
    } else {
        w.WriteHeader(http.StatusNotFound)
    }
}
