package main

import (
	"fmt"
	"os"
	"qlist/widgets"
	"strconv"

	"runtime"

	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/andybrewer/mack"
	"github.com/asaskevich/govalidator"

	"github.com/fstanis/screenresolution"

	"github.com/oq-x/go-plist"
	"github.com/sqweek/dialog"
)

var plistType string
var filename string
var manager = Manager{}

func ParsePlist(filename string, w fyne.Window, tree *widget.Tree, entries Entries) *Entries {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	plistData = plist.OrderedDict{}
	arrayPlist = []interface{}{}
	entries = make(Entries)
	_, err = plist.Unmarshal(content, &plistData)
	if err != nil {
		_, e := plist.Unmarshal(content, &arrayPlist)
		if e != nil {
			fmt.Println("Error parsing dict plist data:", err)
			fmt.Println("Error parsing array plist data:", e)
			mack.Alert("Error", "Failed to parse plist", "critical")
			w.Close()
		} else {
			fmt.Printf("[INFO] Parsed array plist %s\n", filename)
			plistType = "Array"
			for index, value := range arrayPlist {
				key := fmt.Sprintf("%v", index)
				path := strconv.Itoa(index)
				entry := Parse(key, value, path, len(entries), entries)
				entries[path] = entry
			}
		}
	} else {
		fmt.Printf("[INFO] Parsed dict plist %s\n", filename)
		plistType = "Dictionary"
		for index, key := range plistData.Keys {
			Parse(key, plistData.Values[index], key, len(entries), entries)
		}
	}
	tree.OpenAllBranches()
	w.SetTitle(filename)
	w.SetContent(tree)
	return &entries
}

func main() {
	entries := make(Entries)
	a := app.New()
	w := a.NewWindow("Qlist Plist Editor")
	for _, a := range os.Args {
		if b, _ := govalidator.IsFilePath(a); b {
			if strings.HasSuffix(a, ".plist") {
				filename = a
			}
		}
	}
	w.SetCloseIntercept(func() {
		if runtime.GOOS == "darwin" {
			if len(plistData.Keys) > 0 {
				response, _ := mack.AlertBox(mack.AlertOptions{
					Title:   "Do you want to save this file?",
					Style:   "critical",
					Buttons: "No, Yes, Cancel",
				})
				if response.Clicked == "No" {
					w.Close()
				}
			} else {
				w.Close()
			}
		} else {
			w.Close()
		}
	})
	tree := widget.NewTree(
		func(path widget.TreeNodeID) []widget.TreeNodeID {
			children := []string{}
			if path == "" {
				children = append(children, "Root")
			} else {
				if path == "Root" {
					children = manager.Keys()
				} else {
					entry := entries[path]
					children = entry.childrenPaths
				}
			}
			return children
		},
		func(path widget.TreeNodeID) bool {
			if path == "" || path == "Root" {
				return true
			}
			entry := entries[path]
			if &entry == nil {
				return false
			}
			if entry.value == nil {
				return true
			}
			return false
		},
		func(branch bool) fyne.CanvasObject {
			key := widgets.NewText("Key")
			keyEntry := widgets.NewEntry(key.Resource.Text)
			keyEntry.Hide()
			key.SetDoubleTapEvent(func(_ *fyne.PointEvent) {
				key.Hide()
				keyEntry.SetText(key.Resource.Text)
				keyEntry.Show()
			})
			typ := widgets.NewText("Type")
			value := widgets.NewText("Value")
			valueEntry := widgets.NewEntry(value.Resource.Text)
			valueEntry.Hide()
			value.SetDoubleTapEvent(func(_ *fyne.PointEvent) {
				value.Hide()
				valueEntry.Show()
			})

			return container.New(layout.NewGridLayout(3), key, keyEntry, typ, value)
		},
		func(path widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
			container, _ := o.(*fyne.Container)
			key := container.Objects[0].(*widgets.Text)
			typ := container.Objects[2].(*widgets.Text)
			value := container.Objects[3].(*widgets.Text)
			if path == "Root" {
				key.SetText("Root")
				typ.SetText(plistType)
				if manager.Length() == 1 {
					value.SetText("1 key/value entries")
				} else {
					value.SetText(fmt.Sprintf("%v key/value entries", manager.Length()))
				}
			} else {
				entry := entries[path]
				if &entry == nil {
					key.SetText("N/A")
					typ.SetText("N/A")
					value.SetText("N/A")
				} else {
					t, v := GetType(entry)
					key.SetText(entry.key)
					typ.SetText(t)
					value.SetText(v.display)
				}
			}
			return
		})
	text := canvas.NewText("Please upload a plist file", theme.TextColor())
	text.Alignment = fyne.TextAlignCenter
	text.TextSize = 25

	fileitem := fyne.NewMenuItem("Open", func() {
		filename, err := dialog.File().Filter("Property-List File", "plist").Load()
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		entries = *ParsePlist(filename, w, tree, entries)
	})

	filemenu := fyne.NewMenu("File", fileitem)
	mainmenu := fyne.NewMainMenu(filemenu)
	w.SetMainMenu(mainmenu)
	resolution := screenresolution.GetPrimary()
	w.Resize(fyne.Size{Width: float32(resolution.Width), Height: float32(resolution.Height)})

	if filename == "" {
		w.SetContent(text)
	} else {
		entries = *ParsePlist(filename, w, tree, entries)
	}

	w.ShowAndRun()
}
