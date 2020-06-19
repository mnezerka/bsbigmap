package main

import (
    "fmt"
    "html/template"
    "math/rand"
    "net/http"
    "net/url"
    "regexp"
    "strconv"
    "strings"
    "github.com/op/go-logging"
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

type HandlerRoot struct {
    log *logging.Logger
    providers map[string]Provider
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

func WriteErrorResponse(w http.ResponseWriter, status int, err error) {
    w.Header().Set("Content-Type", "text/plain")
    w.WriteHeader(status)
    w.Write([]byte(err.Error()))
}

func (h *HandlerRoot) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    var err error

    h.log.Debugf("Processing request, path: %s", r.URL.Path)

    // check http method, GET is required
    if r.Method != http.MethodGet {
        WriteErrorResponse(w, http.StatusMethodNotAllowed, fmt.Errorf("Only GET method is allowed"))
        return
    }

    // check http method, GET is required
    if r.URL.Path != "/" {
        h.log.Warningf("Ignoring request to path %s", r.URL.Path)
        WriteErrorResponse(w, http.StatusNotFound, fmt.Errorf("Only root path is supported"))
        return
    }

    ip := InputParams{}

    ////////////////////////////////////
    // PROCESS QUERY PARAMETERS

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

    mp := MapParams{}

    urlBase := *r.URL
    urlBase.Scheme = "http"
    urlBase.Host = r.Host
    urlBase.Path = r.URL.Path

    // normallization of input parameters
    ip.Zoom = IntMax(IntMin(ip.Provider.MaxZoom, ip.Zoom), ip.Provider.MinZoom)
    zoom2 := ip.Zoom * ip.Zoom

    ip.XMin = IntMax(0, ip.XMin)
    ip.YMin = IntMax(0, ip.YMin)
    ip.XMax = IntMin(zoom2 - 1, ip.XMax)
    ip.YMax = IntMin(zoom2 - 1, ip.YMax)

    if ip.XMax < ip.XMin { ip.XMax = ip.XMin }
    if ip.YMax < ip.YMin { ip.YMax = ip.YMin }

    mp.WidthTiles = ip.XMax - ip.XMin + 1
    mp.HeightTiles = ip.YMax - ip.YMin + 1
    mp.WidthPx = mp.WidthTiles * ip.Provider.Scale
    mp.HeightPx = mp.HeightTiles * ip.Provider.Scale

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

    data := HtmlMap{InputParams: ip, MapParams: mp}

    for y := ip.YMin; y <= ip.YMax; y++ {
        for x := ip.XMin; x <= ip.XMax; x++ {

            xp := ip.Scale * (x - ip.XMin)
            yp := ip.Scale * (y - ip.YMin)

            style := fmt.Sprintf("position: absolute; left: %dpx; top: %dpx; width: %dpx; height: %dpx", xp, yp, ip.Scale, ip.Scale);

            bg := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(ip.Provider.Url, "!z", strconv.Itoa(ip.Zoom)), "!y", strconv.Itoa(y)), "!x", strconv.Itoa(x))

            // load balancing
            // {abc} => a or b or c
            re := regexp.MustCompile(`{[a-z0-9]+}`)
            if loc := re.FindStringIndex(bg); loc != nil {
                charPos := rand.Intn(loc[1] - loc[0] - 2)
                char := bg[loc[0] + 1 + charPos]
                bg = strings.ReplaceAll(bg, bg[loc[0]:loc[1]], string(char))
            }

            data.Images = append(data.Images, HtmlImage{Url: bg, Style: template.CSS(style)})
        }
    }

    tmpl := template.Must(template.ParseFiles("html/main.html"))

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    err = tmpl.Execute(w, data)
    if err != nil {
        w.Write([]byte(fmt.Sprintf("Rendering error: %s", err)))
    }
}



