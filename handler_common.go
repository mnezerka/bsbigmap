package main

import (
    "fmt"
    "net/http"
    "net/url"
    "strconv"
)

type InputParams struct {
    Zoom int
    XMin int
    YMin int
    XMax int
    YMax int
    Scale int
    Provider Provider
}

func WriteErrorResponse(w http.ResponseWriter, status int, err error) {
    w.Header().Set("Content-Type", "text/plain")
    w.WriteHeader(status)
    w.Write([]byte(err.Error()))
}

func parseParamInt(r *http.Request, name string, defaultValue int) (int, error) {
    var err error
    valueInt := defaultValue

    value := r.URL.Query().Get(name)
    if len(value) == 0 {
        return valueInt, nil
    }

    if valueInt, err = strconv.Atoi(value); err != nil {
        return 0, fmt.Errorf("Cannot parse %s query parameter: %s", name, err)
    }

    return valueInt, nil
}

func getMapUrl(base url.URL, zoom, xmin, ymin, xmax, ymax int, provider string, empty bool) string {
    if empty {
        return ""
    }

    q := base.Query()
    q.Set("zoom", strconv.Itoa(zoom))
    q.Set("xmin", strconv.Itoa(xmin))
    q.Set("ymin", strconv.Itoa(ymin))
    q.Set("xmax", strconv.Itoa(xmax))
    q.Set("ymax", strconv.Itoa(ymax))
    base.RawQuery = q.Encode()

    return base.String()
}

func getStitcherUrl(base url.URL, ip *InputParams) string {
    q := base.Query()
    q.Set("zoom", strconv.Itoa(ip.Zoom))
    q.Set("xmin", strconv.Itoa(ip.XMin))
    q.Set("ymin", strconv.Itoa(ip.YMin))
    q.Set("xmax", strconv.Itoa(ip.XMax))
    q.Set("ymax", strconv.Itoa(ip.YMax))
    base.RawQuery = q.Encode()
    base.Path = "stitcher"

    return base.String()
}


