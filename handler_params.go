package main

import (
    "context"
    "fmt"
    "net/http"
    "github.com/op/go-logging"
)

type HandlerParams struct {
    log *logging.Logger
    providers map[string]Provider
    handler http.Handler
}

func (h *HandlerParams) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    var err error

    h.log.Debugf("Processing request, path: %s", r.URL.Path)

    ip := InputParams{}

    // provider
    providerName := r.URL.Query().Get("provider")
    if len(providerName) == 0 {
        h.log.Debugf("No provider specified, trying to choose first")
        for k := range h.providers {
            providerName = k
            h.log.Debugf("Choosen provider: %s", providerName)
            break
        }
    }

    // choose provider
    var Exists bool
    if ip.Provider, Exists = h.providers[providerName]; !Exists {
        WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("Unknown provider: %s", providerName))
        return
    }

    if ip.Zoom, err = parseParamInt(r, "zoom", 3); err != nil {
        WriteErrorResponse(w, http.StatusBadRequest, err)
        return
    }

    if ip.XMin, err = parseParamInt(r, "xmin", 1); err != nil {
        w.WriteHeader(400)
        return
    }

    if ip.YMin, err = parseParamInt(r, "ymin", 1); err != nil {
        w.WriteHeader(400)
        return
    }

    if ip.XMax, err = parseParamInt(r, "xmax", 3); err != nil {
        w.WriteHeader(400)
        return
    }

    if ip.YMax, err = parseParamInt(r, "ymax", 3); err != nil {
        w.WriteHeader(400)
        return
    }

    if ip.Scale, err = parseParamInt(r, "scale", ip.Provider.Scale); err != nil {
        w.WriteHeader(400)
        return
    }

    ////////////////////////////////////
    // PREPARE MAP DATA

    // normallization of input parameters
    ip.Zoom = IntMax(IntMin(ip.Provider.MaxZoom, ip.Zoom), ip.Provider.MinZoom)
    zoom2 := IntPow2(ip.Zoom)

    h.log.Debugf("Map params before corrections:")
    h.log.Debugf("  zoom: %d", ip.Zoom)
    h.log.Debugf("  zoom2: %d", zoom2)
    h.log.Debugf("  zoom2 - 1: %d", zoom2 - 1)
    h.log.Debugf("  xmin ymin: %d %d", ip.XMin, ip.YMin)
    h.log.Debugf("  xmax ymax: %d %d", ip.XMax, ip.YMax)

    ip.XMin = IntMax(0, ip.XMin)
    ip.YMin = IntMax(0, ip.YMin)
    h.log.Debugf("mintest %d", IntMin(zoom2 - 1, 73));
    h.log.Debugf("mintest %d", IntMin(zoom2 - 1, 71));
    ip.XMax = IntMin(zoom2 - 1, ip.XMax)
    ip.YMax = IntMin(zoom2 - 1, ip.YMax)

    if ip.XMax < ip.XMin { ip.XMax = ip.XMin }
    if ip.YMax < ip.YMin { ip.YMax = ip.YMin }

    h.log.Debugf("Map params after corrections:")
    h.log.Debugf("  zoom: %d", ip.Zoom)
    h.log.Debugf("  zoom2: %d", zoom2)
    h.log.Debugf("  zoom2 - 1: %d", zoom2 - 1)
    h.log.Debugf("  xmin ymin: %d %d", ip.XMin, ip.YMin)
    h.log.Debugf("  xmax ymax: %d %d", ip.XMax, ip.YMax)

    ctx := r.Context()

    ctx = context.WithValue(ctx, "ip", &ip)

    h.handler.ServeHTTP(w, r.WithContext(ctx))
}
