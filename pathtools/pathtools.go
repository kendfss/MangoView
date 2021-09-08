package pathtools // catlab/pathtools
// package main

import (
    "errors"
    "fmt"
    "io/fs"
    "log"
    "os"
    "path"
    "strings"
    "time"
    
    et"mangoview/errortools"
    it"mangoview/itertools"
)
func Cd(pth string) {
    err := os.Chdir(pth)
    et.Assert(err)
}
func Pwd() string {
    home, err := os.Getwd()
    et.Assert(err)
    return home
}
// func Normalize(path) {
    
// }
func IsDir(pth string) bool {
    f, err := os.Lstat(pth)
    et.Check(err)
    mode := f.Mode()
    return mode.IsDir()
}
func IsFile(pth string) bool {
    f, err := os.Lstat(pth)
    et.Check(err)
    mode := f.Mode()
    return mode.IsRegular()
}
func IsSymLink(pth string) bool {
    f, err := os.Lstat(pth)
    et.Check(err)
    mode := f.Mode()
    return mode&fs.ModeSymlink != 0
}
func IsNamedPipe(pth string) bool {
    f, err := os.Lstat(pth)
    et.Check(err)
    mode := f.Mode()
    return mode&fs.ModeNamedPipe != 0
}
func ExtensionSupported(pth string) bool {
    switch path.Ext(pth) {
        case ".jpeg":
            return true
        case ".jpg":
            return true
        case ".png":
            return true
        default:
            return false
    }
}
func SuitableImage(dir string) string {
    files := Files(dir)
    var ctr int
    var pth string
    finding:
        for ; !ExtensionSupported(pth); ctr++ {
            if ctr >= len(files) {
                log.Fatalf("Directory \"%s\" doesn't contain any suitable images", dir)
                break finding
            }
            pth = files[it.Randex(len(files))]
        }
    return pth
}

func ProjDir(pth string) string {
    pwd, err := os.Getwd()
    et.Assert(err)
    pth = strings.Replace(pth, "\\", "/", -1)
    _, name := path.Split(pth)
    // fmt.Println(name)
    parts := strings.Split(name, ".")
    ext := parts[len(parts)-1]
    name = strings.Join(parts[:len(parts)-1], ".") + "-" + ext
    new := path.Join(pwd, "cat_maps", name)
    return NameSpacer(new)
}
func Files(root string) []string {
    paths := make([]string, 0)
    for _, name := range Listdir(root) {
        pth := path.Join(root, name)
        stat, err := os.Lstat(pth)
        if err != nil {
            log.Fatal(err)
        }
        switch mode := stat.Mode(); {
            case mode.IsRegular():
                paths = append(paths, pth)
            case mode.IsDir():
                go it.Merge(&paths, Files(pth))
        }
    }
    return paths
}
func Folders(root string) []string {
    paths := make([]string, 0)
    for _, name := range Listdir(root) {
        pth := path.Join(root, name)
        // fmt.Println(pth)
        stat, err := os.Lstat(pth)
        if err != nil {
            log.Fatal(err)
        }
        switch mode := stat.Mode(); {
            case mode.IsRegular():
                continue
            case mode.IsDir():
                paths = append(paths, pth)
                go it.Merge(&paths, Folders(pth))
        }
    }
    return paths
}
func Listdir(pth string) []string {
    files, err := os.ReadDir(pth)
    et.Check(err)
    rack := make([]string, 0)
    for _, file := range files {
        rack = append(rack, file.Name())
    }
    return rack
}
type namespacer struct {
    Format string
    Index int
}
func (ns *namespacer) space(pth string) string {
    _, err := os.Stat(pth)
    if err != nil {
        return pth
    } else {
        n := ns.new(pth)
        _, err = os.Stat(n)
        if err != nil {
            return n
        } else {
            ns.Index += 1
            return ns.space(pth)
        }
    }
}
func (ns namespacer) new(pth string) string {
    ext := path.Ext(pth)
    name := pth[:len(pth)-len(ext)]
    return (name + fmt.Sprintf(ns.Format, ns.Index) + ext)
}
func NameSpacer(pth string) string {
    ns := namespacer{"_%v", 2}
    return ns.space(pth)
}




func Exists(path string) bool {
    _, err := os.Stat(path)
    return !errors.Is(err, os.ErrNotExist)
}



func main() {
    folder := "c:/users/kenneth/pictures"
    folders := Folders(folder)
    it.Show(folders)
    // for _, f := range folders {
    //     fmt.Println(f)
    // }
    files := Files(folder)
    it.Show(files)
    // for _, f := range files {
    //     fmt.Println(f)
    // }
    // Show(Files(folder))
}


func Touch(fileName string) {
    _, err := os.Stat(fileName)
    if os.IsNotExist(err) {
        file, err := os.Create("temp.txt")
        if err != nil {
            log.Printf("Couldn't create file:\n\t%q\n\t%q\n", fileName, err)
        }
        defer file.Close()
    } else {
        currentTime := time.Now().Local()
        err = os.Chtimes(fileName, currentTime, currentTime)
        if err != nil {
            log.Printf("Couldn't touch file:\n\t%q\n\t%q\n", fileName, err)
        }
    }
}