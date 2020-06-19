package main

import (
    "encoding/csv"
    "io"
    "os"
    "strconv"
)

type Provider struct {
    Name string
    MinZoom int
    MaxZoom int
    Scale int
    Url string
    Attribution string
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
