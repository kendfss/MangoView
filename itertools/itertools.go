package itertools // catlab/itertools
// package main

import (
	"fmt"
    "crypto/rand"
    // "math/rand"
    
    et"mangoview/errortools"
)

func Randex(length int) int {
    bites := make([]byte, 1)
    _, err := rand.Read(bites)
    et.Assert(err)
    return int(bites[0]) % length
}

func SameImage(slice1, slice2 []uint8) bool {
    if len(slice1) != len(slice2) {
        return false
    }
    for i, value := range slice1 {
        if value != slice2[i] {
            return false
        }
    }
    return true
}
func SameImageTensor(slice1, slice2 [][][]uint8) bool {
    if len(slice1) != len(slice2) {
        return false
    }
    if len(slice1[0]) != len(slice2[0]) {
        return false
    }
    width := len((slice1)[0])
    height := len(slice1)
    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            c1 := slice1[y][x]
            c2 := slice2[y][x]
            if !SameImage(c1, c2) {
                return false
            }
        }
    }
    return true
}

func Merge(receiver *[]string, giver []string) {
    for _, str := range giver {
        *receiver = append(*receiver, str)
    }
}

func Show(box []string) {
    // print elements from an arbitrary iterable
    for _, element := range box {
        fmt.Println(element)
    }
}

func Filter(f func(interface{})bool, iter []interface{}) []interface{} {
    out := make([]interface{}, 0)
    for _, elem := range iter {
        if f(elem) {
            out = append(out, elem)
        }
    }
    return out
}
