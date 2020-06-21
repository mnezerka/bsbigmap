package main

import (
    "fmt"
    "testing"
)

func TestIntMin(t *testing.T) {
    Equals(t, 1, IntMin(1, 2))
    Equals(t, 1, IntMin(2, 1))
    Equals(t, 1, IntMin(1, 1))
}

func TestIntMax(t *testing.T) {
    Equals(t, 2, IntMax(1, 2))
    Equals(t, 2, IntMax(2, 1))
    Equals(t, 1, IntMax(1, 1))
}

func TestIntPow2(t *testing.T) {
    Equals(t, 1, IntPow2(0))
    Equals(t, 2, IntPow2(1))
    Equals(t, 4, IntPow2(2))
    Equals(t, 8, IntPow2(3))
    Equals(t, 256, IntPow2(8))
}

func TestUniqueId(t *testing.T) {
    id := UniqueId()
    fmt.Printf("id: %s", id)
}

