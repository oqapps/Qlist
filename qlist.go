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
	fdialog "fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/andybrewer/mack"
	"github.com/asaskevich/govalidator"

	//"github.com/fstanis/screenresolution"

	"github.com/oq-x/go-plist"
	// "github.com/sqweek/dialog"
)

var plistType string
var selectedFilePath string
var manager = Manager{}
var types = []string{"String", "Number", "Data", "Dictionary", "Array", "Date", "Boolean"}
var topTypes = []string{types[3], types[4]}
var log = Logger{}
var rootEntryKeys []string
var tree *widget.Tree

func CreateTree(entries Entries) *widget.Tree {
	tree := widget.NewTree(
		func(path widget.TreeNodeID) []widget.TreeNodeID {
			if path == "" {
				return []string{"Root"}
			} else {
				if path == "Root" {
					return rootEntryKeys
				} else {
					return entries[path].childrenPaths
				}
			}
		},
		func(path widget.TreeNodeID) bool {
			if path == "" || path == "Root" {
				return true
			}
			entry := entries[path]
			if entry.key == "" {
				return false
			}
			return entry.isParent
		},
		func(branch bool) fyne.CanvasObject {
			key := widgets.NewText("Key")
			typ := widgets.NewText("Type")
			value := widgets.NewText("Value")

			container := container.NewGridWithColumns(3, key, typ, value)

			typeText, isText := container.Objects[1].(*widgets.Text)
			if isText && typeText.DoubleTapEvent == nil {
				typeText.SetDoubleTapEvent(func(_ *fyne.PointEvent) {
					sel := widget.NewSelect(types, nil)
					container.Objects[1] = sel
					container.Refresh()
				})
			}
			return container
		},
		func(path widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
			container, _ := o.(*fyne.Container)
			key := container.Objects[0].(*widgets.Text)
			typeSelect, isSelect := container.Objects[1].(*widget.Select)
			typeText, isText := container.Objects[1].(*widgets.Text)
			/*if isSelect {
				fmt.Println("is select!")
				var currentType string
				var entry Entry
				var display Display
				if path == "" {
					currentType = plistType
				} else {
					entry = entries[path]
					display = manager.Display(entry)
					currentType = display.typeText
				}
				typeSelect.OnChanged = func(s string) {
					if currentType == "Dictionary" && s == "Array" {
						var keys []string
						var entry Entry
						if path == "" {
							plistType = s
							keys = rootEntryKeys
						} else {
							entry = entries[path]
							entry.array = true
							keys = entry.childrenPaths
						}
						for i, k := range keys {
							e := entries[k]
							e.SetKey(entries, fmt.Sprint(i))
							key.SetText(e.key)
							keys[i] = fmt.Sprint(i)
							if path == "" {
								rootEntryKeys[i] = keys[i]
							} else {
								entry.children[i].SetKey(entries, keys[i])
								entry.childrenPaths[i] = keys[i]
								entries[entry.path] = entry
							}
						}
						tree.OpenAllBranches()
					} else if typeText.Resource.Text == "Array" && s == "Dictionary" {
						plistType = s
					}
					typeText.SetText(s)
					container.Objects[1] = typeText
				}
				if path == "" {
					typeSelect.SetSelected(plistType)
					typeSelect.Options = topTypes
				} else {
					typeSelect.SetSelected(currentType)
				}
			}*/
			value := container.Objects[2].(*widgets.Text)
			if path == "Root" {
				key.SetID("")
				key.SetText("Root")
				if isSelect {
					typeSelect.Options = topTypes
					typeSelect.SetSelected(plistType)
				} else if isText {
					typeText.SetText(plistType)
				}
				length := len(rootEntryKeys)
				if length == 1 {
					value.SetText("1 key/value entry")
				} else {
					value.SetText(fmt.Sprintf("%v key/value entries", length))
				}
			} else {
				entry := entries[path]
				display := manager.Display(entry)
				key.SetText(entry.key)
				if isSelect {
					typeSelect.SetSelected(display.typeText)
				} else if isText {
					typeText.SetText(display.typeText)
				}
				value.SetText(display.value.display)
			}
		})

	tree.OpenAllBranches()
	return tree
}

func ParsePlist(selectedFilePath string, w fyne.Window, entries Entries) *Entries {
	content, err := os.ReadFile(selectedFilePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	plistData = plist.OrderedDict{}
	arrayPlist = []interface{}{}
	rootEntryKeys = []string{}
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
			log.Info("Parsed array plist", selectedFilePath)
			plistType = "Array"
			for index, value := range arrayPlist {
				key := fmt.Sprintf("%v", index)
				path := strconv.Itoa(index)
				entry := Parse(key, value, path, len(entries), entries)
				entries[path] = entry
			}
		}
	} else {
		log.Info("Parsed dict plist", selectedFilePath)
		plistType = "Dictionary"
		for index, key := range plistData.Keys {
			rootEntryKeys = append(rootEntryKeys, key)
			Parse(key, plistData.Values[index], key, len(entries), entries)
		}
	}
	tree = CreateTree(entries)
	w.SetTitle(selectedFilePath)
	w.SetContent(tree)
	return &entries
}

func main() {
	entries := make(Entries)
	app := app.New()
	window := app.NewWindow("Qlist Plist Editor")

	for _, a := range os.Args {
		if b, _ := govalidator.IsFilePath(a); b {
			if strings.HasSuffix(a, ".plist") {
				selectedFilePath = a
			}
		}
	}
	window.SetCloseIntercept(func() {
		if runtime.GOOS == "darwin" {
			if len(plistData.Keys) > 0 {
				response, _ := mack.AlertBox(mack.AlertOptions{
					Title:   "Do you want to save this file?",
					Style:   "critical",
					Buttons: "No, Yes, Cancel",
				})
				if response.Clicked == "No" {
					window.Close()
				}
			} else {
				window.Close()
			}
		} else {
			var dialog *fdialog.CustomDialog
			var yesButton *widget.Button
			if selectedFilePath != "" {
				dialog = fdialog.NewCustomWithoutButtons("Do you want to save this file?", widget.NewRichTextFromMarkdown(fmt.Sprintf("## You have unsaved changes in the file %s.", GetFileName(selectedFilePath))), window)
				yesButton = widget.NewButtonWithIcon("Save", theme.ConfirmIcon(), func() {})
			} else {
				dialog = fdialog.NewCustomWithoutButtons("Do you want to create this file?", container.NewWithoutLayout(), window)
				yesButton = widget.NewButtonWithIcon("Create", theme.ConfirmIcon(), func() {
					var filename string
					//var err error
					//if runtime.GOOS == "android" || runtime.GOOS == "ios" {
					fdialog.NewFileSave(func(uc fyne.URIWriteCloser, err error) {
						filename = uc.URI().String()
					}, window)
					/*} else {
						filename, err = ndialog.File().Filter("Property-List File", "plist").SetStartFile("Untitled.plist").Save()
						if err != nil {
							dialog.Hide()
						}
					}*/
					data := ParseEntries(entries)
					file, _ := os.Create(filename)
					plist.NewEncoder(file).Encode(data)
					window.Close()
				})
			}
			yesButton.Importance = widget.HighImportance
			noButton := widget.NewButtonWithIcon("Don't Save", theme.DeleteIcon(), func() {
				window.Close()
			})
			noButton.Importance = widget.DangerImportance
			cancelButton := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
				dialog.Hide()
			})
			buttons := []fyne.CanvasObject{yesButton, noButton, cancelButton}
			dialog.SetButtons(buttons)
			dialog.Show()
		}
	})
	text := canvas.NewText("Please upload a plist file", theme.TextColor())
	text.Alignment = fyne.TextAlignCenter
	text.TextSize = 25

	openFile := fyne.NewMenuItem("Open", func() {
		//if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		fdialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
			selectedFilePath = uc.URI().String()
		}, window)
		/*} else {
			var err error
			selectedFilePath, err = ndialog.File().Filter("Property-List File", "plist").Load()
			if err != nil {
				return
			}
		}*/
		entries = *ParsePlist(selectedFilePath, window, entries)
	})

	newFile := fyne.NewMenuItem("New", func() {
		plistType = "Dictionary"
		log.Info("Created new dict plist")
		entries = make(Entries)
		tree = CreateTree(entries)
		window.SetContent(tree)
	})

	filemenu := fyne.NewMenu("File", openFile, newFile)
	mainmenu := fyne.NewMainMenu(filemenu)
	window.SetMainMenu(mainmenu)
	//if runtime.GOOS == "linux" || runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
	//	resolution := screenresolution.GetPrimary()
	//	window.Resize(fyne.Size{Width: float32(resolution.Width), Height: float32(resolution.Height)})
	//}

	if selectedFilePath == "" {
		window.SetContent(text)
	} else {
		entries = *ParsePlist(selectedFilePath, window, entries)
	}
	window.ShowAndRun()
}
