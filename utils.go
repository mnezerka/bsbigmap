package main

import (
    "fmt"
    "math/rand"
    "time"
)

func IntMin(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func IntMax(a, b int) int {
    if a > b {
        return a
    }
    return b
}

func IntPow2(a int) int {
    if a < 1 {
        return 1
    }
    result := 2
    for i := 1; i < a; i++ {
        result = result * 2
    }
    return result
}

// yyyy-mm-dd-hh-mm-ss-rrrr, where rrrr is random string
func UniqueId() string {
    t := time.Now()
    uuid := fmt.Sprintf("%d-%02d-%02d-%02d-%02d-%02d-%09d-%04d",
        t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), rand.Intn(9999))
    return uuid
}
