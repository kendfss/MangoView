package iotools // mangoview/iotools
// package main

import (
    "fmt"
    "io"
    "image"
    "image/png"
    "image/jpeg"
    "strings"
    "path"
    "os"
    "os/exec"
    
    pt"mangoview/pathtools"
    et"mangoview/errortools"
)

// func LoadImage(filePath string) (image.Image) {
// func decoder(ext string) (func(io.Reader) (image.Image, str, err)) {
//     opts := map[string](func(io.Reader) (image.Image, str, err)) {
//         ".jpeg": jpeg.Decode
//         ".jpg": jpeg.Decode
//         ".png": png.Decode
//         // ".gif": gif.Decode
//     }
//     return opts[ext]
// }
func Decoder(handle *os.File) (func(io.Reader) (image.Image, error)) {
    pth := handle.Name()
    
    switch strings.ToLower(path.Ext(pth)) {
        case ".jpeg":
            return jpeg.Decode
        case ".jpg":
            return jpeg.Decode
        case ".png":
            return png.Decode
        default:
            panic("Couldn't find an appropriate decoder for this image")
    }
}
func LoadImageFace(filePath string) image.Image {
    f, err := os.Open(filePath)
    defer f.Close()
    et.Assert(err)
    // img, _, err := image.Decode(f)
    img, err := Decoder(f)(f)
    if err != nil {
        panic(fmt.Sprintf("Couldn't decode image file:\n\t\"%s\"\n%s", filePath, err))
    }
    return img
}
// func LoadImage(filePath string) ([]uint8, int) {
func LoadImage(filePath string) (*image.NRGBA) {
    f, err := os.Open(filePath)
    defer f.Close()
    et.Assert(err)
    // img, _, err := image.Decode(f)
    img, err := Decoder(f)(f)
    if err != nil {
        panic(fmt.Sprintf("Couldn't decode image file:\n\t\"%s\"\n%s", filePath, err))
    }
    // pix := make([]uint8, 0)
    // stride := img.Stride
    out := image.NewNRGBA(img.Bounds())
    for y := 0; y < img.Bounds().Max.Y; y++ {
        for x := 0; x < img.Bounds().Max.X; x++ {
            out.Set(x, y, img.At(x, y))
        }
    }
    // return img, stride
    return out
}
func SavePNG(m *image.NRGBA) string {
    name := pt.NameSpacer("image.png")
    w, err := os.Create(name)
    defer w.Close()
    et.Assert(err)
    encoder := &png.Encoder{CompressionLevel: png.BestCompression}
    err = encoder.Encode(w, m)
    et.Assert(err)
    return name
}
func set_atTensor(x, y int, im *image.NRGBA, vals []uint8) {
    i := y*im.Stride + x*4
    (*im).Pix[i] = (vals)[0] // red
    (*im).Pix[i+1] = (vals)[1] // green
    (*im).Pix[i+2] = (vals)[2] // blue
    (*im).Pix[i+3] = (vals)[3] // alpha
}
func RenderPNGTensor(mat [][][]uint8) string {
    height := len(mat)
    width := len(mat[0])
    im := image.NewNRGBA(image.Rect(0, 0, width, height))
    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            go set_atTensor(x, y, im, mat[y][x])
        }
    }
    return SavePNG(im)
}

func set_at(x, y int, im *image.NRGBA, vals []uint8) {
    i := y*im.Stride + x*4
    (*im).Pix[i] = vals[i] // red
    (*im).Pix[i+1] = vals[i+1] // green
    (*im).Pix[i+2] = vals[i+2] // blue
    (*im).Pix[i+3] = vals[i+3] // alpha
}
func RenderPNG(mat []uint8, width, height int) string {
    im := image.NewNRGBA(image.Rect(0, 0, width, height))
    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            go set_at(x, y, im, mat)
        }
    }
    return SavePNG(im)
}

func Pop(name string) {
    err := exec.Command("explorer", name).Run()
    et.Check(err)
}

func main() {
    folder := "c:/users/kenneth/pictures"
    folders := pt.Folders(folder)
    for _, f := range folders {
        fmt.Println(f)
    }
    files := pt.Files(folder)
    for _, f := range files {
        fmt.Println(f)
    }
    // Show(Files(folder))
}


