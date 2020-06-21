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
    SubDomains string
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

        // load balancing
        // [abc] => {s} + remember "abc" for distribution to a or b or c
        re := regexp.MustCompile(`[[][a-z0-9]+[]]`)
        if loc := re.FindStringIndex(p.Url); loc != nil {
            p.SubDomains = p.Url[loc[0] + 1:loc[1] - 1]
            p.Url = strings.ReplaceAll(p.Url, p.Url[loc[0]:loc[1]], "{s}")
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

            url := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(p.Url, "{z}", strconv.Itoa(zoom)), "{y}", strconv.Itoa(y)), "{x}", strconv.Itoa(x))

            // load balancing
            // pick random character from subdomains
            if p.SubDomains != "" {
                charPos := rand.Intn(len(p.SubDomains) - 1)
                char := p.SubDomains[charPos]
                url = strings.ReplaceAll(url, "{s}", string(char))
            }

            tiles = append(tiles, Tile{xp, yp, url})
        }
    }
    return &tiles
}
