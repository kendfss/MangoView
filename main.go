/*
TODO
	save state
		statefilepath -> sfname
					  -> sfpath() -> root / sfname
	    structify
	dark theme
	tabs
	explore
	patch leaky keyboard input
*/
package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"math"
	"os"
	fp"path/filepath"
	"sort"
	"strings"
	"strconv"
	"time"

	"gioui.org/app"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	_"golang.org/x/exp/shiny/materialdesign/icons"

	iot"mangoview/iotools"
	pt"mangoview/pathtools"
)

var (
	stateFilePath string = "/Users/kendfss/mangoview.json" 
	// stateFilePath string = "/Users/kendfss/.mangoview" 
	// stateFilePath string = "~/.mangoview" 

	readRTL bool = false

	transformTime time.Time
	gtx layout.Context
	window *app.Window
	theme *material.Theme

	keyHandler = new(int)

	yFactor int = 565

	list = &layout.List{
		Axis: layout.Vertical,
		Alignment: layout.Middle,
	}

	root string
	currentChapterIndex int = -1

	pages []string
	chapters []string
	currentPageIndex int = -1
	currentPagePath string
	currentPage paint.ImageOp

	loaded bool = false
	progress = float32(0)
	progressIncrementer chan float32

	prevChapButton = new(widget.Clickable)	
	prevPageButton = new(widget.Clickable)
	nextPageButton = new(widget.Clickable)
	nextChapButton = new(widget.Clickable)

	prevChapButtonClicked bool 
	prevPageButtonClicked bool 
	nextPageButtonClicked bool 
	nextChapButtonClicked bool 
)

type state struct{
	root string `json:"root"`
	page int `json:"page"`
	chapter int `json:"chapter"`
	rtl bool `json:"rtl"`
}
func (self state) save() {
	pt.Touch(stateFilePath)
	// state := make(map[string]string)
	// state["currentPageIndex"] = fmt.Sprintf("%v", currentPageIndex)
	// state["currentChapterIndex"] = fmt.Sprintf("%v", currentChapterIndex)
	// state["root"] = root
	// fmt.Printf("state:\n\t%s\n", state)
	data, err := json.Marshal(self)
	if err != nil {
		log.Printf("Couldn't marshal state:\n\t%v\n\t%q\n", self, err)
		return
	}
	// fobj, err := os.OpenFile(stateFilePath, os.O_WRONLY, os.ModeType)
	// fobj, err := os.OpenFile(stateFilePath, os.O_WRONLY, os.ModeAppend)
	fobj, err := os.OpenFile(stateFilePath, os.O_WRONLY, os.ModePerm)
	// defer fobj.Close()  
	if err != nil {
		log.Printf("Couldn't access state file:\n\t%q\n\t%q\n", stateFilePath, err)
		return
	}
	encoder := json.NewEncoder(fobj) 
	err = encoder.Encode(data)
	if err != nil {
		log.Printf("Couldn't encode state file:\n\t%q\n\t%q\n", stateFilePath, err)
		return
	}
	err = fobj.Close()
	if err != nil {
		log.Printf("Couldn't close state file:\n\t%q\n\t%q\n", stateFilePath, err)
		return 
	}
	log.Println("Successfully saved State File")
}
func (self state) load() {
	if pt.Exists(stateFilePath) {
		// state := make(map[string]string)
		data, err := ioutil.ReadFile(stateFilePath)
		if err != nil {
			log.Printf("Couldn't read state file:\n\t%q\n", stateFilePath)
			return
		}
		err = json.Unmarshal(data, self) 
		if err != nil {
			log.Printf("Couldn't Unmarshal state file:\n\t%q\n", stateFilePath)
			return 
		}
		// currentPageIndex_, err := strconv.ParseInt(self.page, 10, 0)
		// // currentImageIndex_, err := strconv.ParseInt(state["currentPageIndex"], 10, 0)
		// if err != nil {
		// 	log.Printf("Couldn't parse current image index:\n\t%q", err)
		// 	return 
		// }
		// currentPageIndex = int(currentPageIndex_)
		// currentChapterIndex_, err := strconv.ParseInt(self.chapter, 10, 0)
		// // currentChapterIndex_, err := strconv.ParseInt(state["currentChapterIndex"], 10, 0)
		// if err != nil {
		// 	log.Printf("Couldn't parse current image index:\n\t%q", err)
		// 	return 
		// }
		// currentChapterIndex = int(currentChapterIndex_)
		// // root = state["root"]
		// root = self.root
		// readRTL = self.rtl
		// log.Println("Successfully loaded State File")
	} else {
		log.Printf("State file does not exist:\n\t%q\n", stateFilePath)
		pt.Touch(stateFilePath)
	}
}

func absInt(arg int) int {
	return int(math.Abs(float64(arg)))
}
func normInt(arg, ceil, min int) int {
	if arg < 0 {
		arg = ceil-1
	} else if arg >= ceil {
		arg = 0
	}
	return arg

}
func loadImage(path string) paint.ImageOp {
	return paint.NewImageOp(iot.LoadImageFace(path))
}
func incrementChapter(delta int) {
	currentChapterIndex += delta
	currentChapterIndex = normInt(currentChapterIndex, len(chapters), 0)

	chapters = pt.Folders(root)
	sort.SliceStable(chapters, compareChapterNames)
}
func incrementImage(delta int) paint.ImageOp {
	if len(pages) == 0 || len(chapters) == 0 {
		incrementChapter(1)
	}

	pages = pt.Files(chapters[currentChapterIndex])

	currentPageIndex += delta
	currentPageIndex = normInt(currentPageIndex, len(pages), 0)
	
	sort.SliceStable(pages, compareImageNames)
	currentPagePath = pages[currentPageIndex]
	return loadImage(currentPagePath)
}
func currentTitle() string {
	titleParts := strings.Split(currentPagePath, fmt.Sprintf("%c", os.PathSeparator))
	return fp.Join(titleParts[len(titleParts)-3:]...)
}
func imageRoot(path string) string {
	d, _ := fp.Split(path)
	d, _ = fp.Split(path)
	return d
}

func zoom(delta int) {
	yFactor += delta
}
func main() {
	root = "/Users/kendfss/Downloaded Manga/Ichi The Killer"
	if len(root) > 0 {
		currentPage = incrementImage(1)
	}
	
	loadState()
	// readRTL = !readRTL
	// fmt.Println(chapters)
	go func() {
		w := app.NewWindow()
		app.Title("MangoView")
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		saveState()
		os.Exit(0)
	}()
	app.Main()
}

func loop(w *app.Window) error {
	theme = material.NewTheme(gofont.Collection())
	var ops op.Ops
	for e := range w.Events() {
		// e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		// case system.StageEvent:
		// 	if e.Stage >= system.StageRunning {
		// 		if app.ctxCancel == nil {
		// 			app.ctx, app.ctxCancel = context.WithCancel(context.Background())
		// 		}
		// 		if app.ui.users == nil {
		// 			go app.fetchContributors()
		// 		}
		// 	} else {
		// 		if app.ctxCancel != nil {
		// 			app.ctxCancel()
		// 			app.ctxCancel = nil
		// 		}
		// 	}
		case key.Event:
			switch e.Name {
			case key.NameUpArrow:
				prevChapButtonClicked = true
			case key.NameLeftArrow:
				prevPageButtonClicked = true
			case key.NameRightArrow:
				nextPageButtonClicked = true
			case key.NameDownArrow:
				nextChapButtonClicked = true
			case key.NameEscape:
				saveState()
				return nil
			}
			w.Invalidate()
		case system.FrameEvent:
			switch {
			case prevChapButton.Clicked():
				prevChapButtonClicked = true
			case prevPageButton.Clicked():
				prevPageButtonClicked = true
			case nextPageButton.Clicked():
				nextPageButtonClicked = true
			case nextChapButton.Clicked():
				nextChapButtonClicked = true
			}
			gtx = layout.NewContext(&ops, e)
			viewer(gtx, theme)
			e.Frame(gtx.Ops)
			saveState()
		}
		w.Option(app.Title(currentTitle()))
	}
	saveState()
	return nil
}

func parseIndex(path string) int64 {
	var prefix string
	_, name := fp.Split(path)
	
	ext := fp.Ext(name)
	if strings.Contains(strings.ToLower(name), "page") {
		prefix = "page"
	} else if strings.Contains(strings.ToLower(name), "chapter") {
		prefix = "chapter "
	}
	name = name[len(prefix):len(name) - len(ext)]
	i, err := strconv.ParseInt(name, 10, 0)
	if err != nil {
		log.Printf("Couldn't parse index:\n\t%q\n\t%q", path, name)
	}
	
	return i
}
func compareImageNames(i, j int) bool {
	if readRTL {
		return parseIndex(pages[i]) > parseIndex(pages[j])
	}
	return parseIndex(pages[i]) < parseIndex(pages[j])
}
func compareChapterNames(i, j int) bool {
	if readRTL {
		return parseIndex(chapters[i]) > parseIndex(chapters[j])
	}
	return parseIndex(chapters[i]) < parseIndex(chapters[j])
}

func viewer(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	in := layout.UniformInset(unit.Dp(8))
	widgets := []layout.Widget{
		func(gtx layout.Context) layout.Dimensions {
			// Title
			return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return in.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						maroon := color.NRGBA{R: 127, G: 0, B: 0, A: 255}
						l := material.H6(theme, currentTitle())
						l.Color = maroon
						l.Alignment = text.Middle
						pointer.CursorNameOp{Name: pointer.CursorPointer}.Add(gtx.Ops)
						return l.Layout(gtx)
					})
				}),
			)
		},
		func(gtx layout.Context) layout.Dimensions {
			// Image
			return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return in.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						img := widget.Image{Src: currentPage}
						img.Position = layout.Center
						img.Scale = float32(yFactor) / float32(gtx.Px(unit.Dp(float32(currentPage.Size().Y))))
						return img.Layout(gtx)
					})
				}),
			)
		},
		func(gtx layout.Context) layout.Dimensions {
			// Control buttons
			return layout.Flex{Spacing: layout.SpaceSides, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return in.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						if prevChapButtonClicked {
							toggler(true, -1)
							prevChapButtonClicked = false
						}
						dims := material.Button(theme, prevChapButton, "<<").Layout(gtx)
						pointer.CursorNameOp{Name: pointer.CursorPointer}.Add(gtx.Ops)
						return dims
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return in.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						if prevPageButtonClicked {
							toggler(false, -1)
							prevPageButtonClicked = false
						}
						dims := material.Button(theme, prevPageButton, "<").Layout(gtx)
						pointer.CursorNameOp{Name: pointer.CursorPointer}.Add(gtx.Ops)
						return dims
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return in.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						if nextPageButtonClicked {
							toggler(false, 1)
							nextPageButtonClicked = false
						}
						dims := material.Button(theme, nextPageButton, ">").Layout(gtx)
						pointer.CursorNameOp{Name: pointer.CursorPointer}.Add(gtx.Ops)
						return dims
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return in.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						if nextChapButtonClicked {
							toggler(true, 1)
							nextChapButtonClicked = false
						}
						dims := material.Button(theme, nextChapButton, ">>").Layout(gtx)
						pointer.CursorNameOp{Name: pointer.CursorPointer}.Add(gtx.Ops)
						return dims
					})
				}),
			)
		},
	}
	return list.Layout(gtx, len(widgets), func(gtx layout.Context, i int) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
	})
}

func toggler(chap bool, delta int) {
	if chap {
		incrementChapter(delta)
		currentPage = incrementImage(-currentPageIndex)
	} else {
		currentPage = incrementImage(delta)
	}
}

func saveState() {
	state{
		root: root,
		page: currentPageIndex,
		chapter: currentChapterIndex,
		rtl: readRTL,
	}.save()
}

func loadState() {
	state{}.load()
}


// func saveState() {
// 	pt.Touch(stateFilePath)
// 	state := make(map[string]string)
// 	state["currentImageIndex"] = fmt.Sprintf("%v", currentImageIndex)
// 	state["currentChapterIndex"] = fmt.Sprintf("%v", currentChapterIndex)
// 	state["root"] = root
// 	// fmt.Printf("state:\n\t%s\n", state)
// 	data, err := json.Marshal(state)
// 	if err != nil {
// 		log.Printf("Couldn't marshal state:\n\t%v\n\t%q\n", state, err)
// 		return
// 	}
// 	// fobj, err := os.OpenFile(stateFilePath, os.O_WRONLY, os.ModeType)
// 	// fobj, err := os.OpenFile(stateFilePath, os.O_WRONLY, os.ModeAppend)
// 	fobj, err := os.OpenFile(stateFilePath, os.O_WRONLY, os.ModePerm)
// 	// defer fobj.Close()  
// 	if err != nil {
// 		log.Printf("Couldn't access state file:\n\t%q\n\t%q\n", stateFilePath, err)
// 		return
// 	}
// 	encoder := json.NewEncoder(fobj) 
// 	err = encoder.Encode(data)
// 	if err != nil {
// 		log.Printf("Couldn't encode state file:\n\t%q\n\t%q\n", stateFilePath, err)
// 		return
// 	}
// 	err = fobj.Close()
// 	if err != nil {
// 		log.Printf("Couldn't close state file:\n\t%q\n\t%q\n", stateFilePath, err)
// 		return 
// 	}
// 	log.Println("Successfully saved State File")
// }

// func loadState() {
// 	if pt.Exists(stateFilePath) {
// 		state := make(map[string]string)
// 		data, err := ioutil.ReadFile(stateFilePath)
// 		if err != nil {
// 			log.Printf("Couldn't read state file:\n\t%q\n", stateFilePath)
// 			return
// 		}
// 		err = json.Unmarshal(data, state) 
// 		if err != nil {
// 			log.Printf("Couldn't Unmarshal state file:\n\t%q\n", stateFilePath)
// 			return 
// 		}
// 		currentImageIndex_, err := strconv.ParseInt(state["currentImageIndex"], 10, 0)
// 		if err != nil {
// 			log.Printf("Couldn't parse current image index:\n\t%q", err)
// 			return 
// 		}
// 		currentImageIndex = int(currentImageIndex_)
// 		currentChapterIndex_, err := strconv.ParseInt(state["currentChapterIndex"], 10, 0)
// 		if err != nil {
// 			log.Printf("Couldn't parse current image index:\n\t%q", err)
// 			return 
// 		}
// 		currentChapterIndex = int(currentChapterIndex_)
// 		root = state["root"]
// 		log.Println("Successfully loaded State File")
// 	} else {
// 		log.Printf("State file does not exist:\n\t%q\n", stateFilePath)
// 		pt.Touch(stateFilePath)
// 	}
// }

