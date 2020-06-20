package main

import (
    "encoding/csv"
    "io"
    "math/rand"
    "os"
    "regexp"
    "strconv"
    "strings"
)

type Provider struct {
    Name string
    MinZoom int
    MaxZoom int
    Scale int
    Url string
    Attribution string
}

type Tile struct {
    Left int
    Top int
    Url string
}

func readProviders() (map[string]Provider, error) {

    csvfile, err := os.Open("providers.csv")
    if err != nil {
        return nil, err
    }

    r := csv.NewReader(csvfile)

    providers := make(map[string]Provider)

    // Iterate through the records
    for {
        rec, err := r.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, err
        }

        // if line is comment, ignore it
        if rec[0][0] == '#' {
            continue
        }

        // skip not complete records
        if len(rec) != 6 {
            continue
        }

        p := Provider{Name: rec[0], Url: rec[4], Attribution: rec[5]}

        if p.MinZoom, err = strconv.Atoi(rec[1]); err != nil {
            return nil, err
        }

        if p.MaxZoom, err = strconv.Atoi(rec[2]); err != nil {
            return nil, err
        }

        if p.Scale, err = strconv.Atoi(rec[3]); err != nil {
            return nil, err
        }

        providers[p.Name] = p
    }

    return providers, nil
}

func (p *Provider) getTiles(xmin, ymin, xmax, ymax, zoom, scale int) *[]Tile {

    var tiles []Tile

    for y := ymin; y <= ymax; y++ {
        for x := xmin; x <= xmax; x++ {

            xp := x - xmin
            yp := y - ymin

            url := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(p.Url, "!z", strconv.Itoa(zoom)), "!y", strconv.Itoa(y)), "!x", strconv.Itoa(x))

            // load balancing
            // {abc} => a or b or c
            re := regexp.MustCompile(`{[a-z0-9]+}`)
            if loc := re.FindStringIndex(url); loc != nil {
                charPos := rand.Intn(loc[1] - loc[0] - 2)
                char := url[loc[0] + 1 + charPos]
                url = strings.ReplaceAll(url, url[loc[0]:loc[1]], string(char))
            }

            tiles = append(tiles, Tile{xp, yp, url})
            //data.Images = append(data.Images, HtmlImage{Url: bg, Style: template.CSS(style)})
        }
    }
    return &tiles
}
