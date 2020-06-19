package main

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
