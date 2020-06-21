package main

import (
    "fmt"
    "html/template"
    "net/http"
    "github.com/op/go-logging"
)

type MapParams struct {
    UrlExpandRight string
    UrlExpandLeft string
    UrlExpandTop string
    UrlExpandBottom string

    UrlShiftRight string
    UrlShiftLeft string
    UrlShiftTop string
    UrlShiftBottom string

    UrlShrinkRight string
    UrlShrinkLeft string
    UrlShrinkTop string
    UrlShrinkBottom string

    UrlZoomInDouble string
    UrlZoomInKeep string
    UrlZoomOutHalf string
    UrlZoomOutKeep string

    UrlGeneratePng string

    WidthTiles int
    HeightTiles int
    WidthPx int
    HeightPx int
}

type HtmlImage struct {
    Url string
    Style template.CSS
}

type HtmlMap struct {
    InputParams InputParams
    MapParams MapParams
    Images []HtmlImage
}

type HandlerMap struct {
    log *logging.Logger
    providers map[string]Provider
}

func (h *HandlerMap) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    var err error

    h.log.Debugf("Processing request, path: %s", r.URL.Path)

    // check http method, GET is required
    if r.Method != http.MethodGet {
        WriteErrorResponse(w, http.StatusMethodNotAllowed, fmt.Errorf("Only GET method is allowed"))
        return
    }

    // check path, only root is allowed
    if r.URL.Path != "/map" {
        h.log.Warningf("Ignoring request to path %s", r.URL.Path)
        WriteErrorResponse(w, http.StatusNotFound, fmt.Errorf("Only /map path is supported"))
        return
    }

    ctx := r.Context()
    ip := ctx.Value("ip").(*InputParams)

    ////////////////////////////////////
    // PREPARE MAP DATA

    mp := MapParams{}

    urlBase := *r.URL
    urlBase.Scheme = "http"
    urlBase.Host = r.Host
    urlBase.Path = r.URL.Path

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

    mp.WidthTiles = ip.XMax - ip.XMin + 1
    mp.HeightTiles = ip.YMax - ip.YMin + 1
    mp.WidthPx = mp.WidthTiles * ip.Provider.Scale
    mp.HeightPx = mp.HeightTiles * ip.Provider.Scale

    h.log.Debugf("Map params after corrections:")
    h.log.Debugf("  zoom: %d", ip.Zoom)
    h.log.Debugf("  zoom2: %d", zoom2)
    h.log.Debugf("  zoom2 - 1: %d", zoom2 - 1)
    h.log.Debugf("  xmin ymin: %d %d", ip.XMin, ip.YMin)
    h.log.Debugf("  xmax ymax: %d %d", ip.XMax, ip.YMax)

    // expand links
    mp.UrlExpandLeft = getMapUrl(urlBase, ip.Zoom, ip.XMin - 1, ip.YMin, ip.XMax, ip.YMax, ip.Provider.Name, ip.XMin == 0)
    mp.UrlExpandRight = getMapUrl(urlBase, ip.Zoom, ip.XMin, ip.YMin, ip.XMax + 1, ip.YMax, ip.Provider.Name, ip.XMax >= zoom2 - 1)
    mp.UrlExpandTop = getMapUrl(urlBase, ip.Zoom, ip.XMin, ip.YMin - 1, ip.XMax, ip.YMax, ip.Provider.Name, ip.YMin == 0)
    mp.UrlExpandBottom = getMapUrl(urlBase, ip.Zoom, ip.XMin, ip.YMin, ip.XMax, ip.YMax + 1, ip.Provider.Name, ip.YMax >= zoom2 - 1)

    // shift links
    mp.UrlShiftLeft = getMapUrl(urlBase, ip.Zoom, ip.XMin - 1, ip.YMin, ip.XMax - 1, ip.YMax, ip.Provider.Name, ip.XMin == 0)
    mp.UrlShiftRight = getMapUrl(urlBase, ip.Zoom, ip.XMin + 1, ip.YMin, ip.XMax + 1, ip.YMax, ip.Provider.Name, ip.XMax >= zoom2 - 1)
    mp.UrlShiftTop = getMapUrl(urlBase, ip.Zoom, ip.XMin, ip.YMin - 1, ip.XMax, ip.YMax - 1, ip.Provider.Name, ip.YMin == 0)
    mp.UrlShiftBottom = getMapUrl(urlBase, ip.Zoom, ip.XMin, ip.YMin + 1, ip.XMax, ip.YMax + 1, ip.Provider.Name, ip.YMax >= zoom2 - 1)

    // shring links
    mp.UrlShrinkLeft = getMapUrl(urlBase, ip.Zoom, ip.XMin + 1, ip.YMin, ip.XMax, ip.YMax, ip.Provider.Name, ip.XMin == ip.XMax)
    mp.UrlShrinkRight = getMapUrl(urlBase, ip.Zoom, ip.XMin, ip.YMin, ip.XMax - 1, ip.YMax, ip.Provider.Name,  ip.XMin == ip.XMax)
    mp.UrlShrinkTop = getMapUrl(urlBase, ip.Zoom, ip.XMin, ip.YMin + 1, ip.XMax, ip.YMax, ip.Provider.Name, ip.YMin == ip.YMax)
    mp.UrlShrinkBottom = getMapUrl(urlBase, ip.Zoom, ip.XMin, ip.YMin, ip.XMax, ip.YMax - 1, ip.Provider.Name, ip.YMin == ip.YMax)

    mp.UrlZoomInDouble = getMapUrl(urlBase, ip.Zoom + 1, ip.XMin * 2, ip.YMin * 2, ip.XMax *2 + 1, ip.YMax * 2 + 1, ip.Provider.Name, ip.Zoom >= ip.Provider.MaxZoom)
    mp.UrlZoomInKeep = getMapUrl(
        urlBase,
        ip.Zoom + 1,
        ip.XMin * 2 + (ip.XMax - ip.XMin) / 2,
        ip.YMin * 2 + (ip.YMax - ip.YMin) / 2,
        ip.XMax * 2 - (ip.XMax - ip.XMin) / 2,
        ip.YMax * 2 - (ip.YMax - ip.YMin) / 2,
        ip.Provider.Name,
        ip.Zoom >= ip.Provider.MaxZoom)

    mp.UrlZoomOutHalf = getMapUrl(urlBase, ip.Zoom - 1, ip.XMin / 2, ip.YMin / 2, ip.XMax / 2, ip.YMax / 2, ip.Provider.Name, ip.Zoom <= ip.Provider.MinZoom)

    mp.UrlZoomOutKeep = getMapUrl(
        urlBase,
        ip.Zoom - 1,
        ip.XMin / 2 - (ip.XMax - ip.XMin) / 4,
        ip.YMin / 2 - (ip.YMax - ip.YMin) / 4,
        ip.XMax / 2 + (ip.XMax - ip.XMin) / 4,
        ip.YMax / 2 + (ip.YMax - ip.YMin) / 4,
        ip.Provider.Name,
        ip.Zoom <= ip.Provider.MinZoom)

    mp.UrlGeneratePng = getStitcherUrl(urlBase, ip)

    data := HtmlMap{InputParams: *ip, MapParams: mp}

    // get tiles for current setting
    tiles := ip.Provider.getTiles(ip.XMin, ip.YMin, ip.XMax, ip.YMax, ip.Zoom, ip.Scale)

    // loop through all tiles - generate style information
    for i := 0; i < len(*tiles); i++ {
        style := fmt.Sprintf("position: absolute; left: %dpx; top: %dpx; width: %dpx; height: %dpx", (*tiles)[i].Left * ip.Scale, (*tiles)[i].Top * ip.Scale, ip.Scale, ip.Scale);
        data.Images = append(data.Images, HtmlImage{Url: (*tiles)[i].Url, Style: template.CSS(style)})
    }

    tmpl := template.Must(template.ParseFiles("html/base.html", "html/map.html"))

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    err = tmpl.Execute(w, data)
    if err != nil {
        w.Write([]byte(fmt.Sprintf("Rendering error: %s", err)))
    }
}
