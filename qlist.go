package main

import (
	"fmt"
	"os"

	"io/ioutil"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/andybrewer/mack"
	"github.com/fstanis/screenresolution"

	"github.com/deitrix/go-plist"
	"github.com/sqweek/dialog"
)


func main() {

	a := app.New()
	w := a.NewWindow("Qlist Plist Editor")
	var root treeNode
	tree := widget.NewTree(
		func(tni widget.TreeNodeID) (nodes []widget.TreeNodeID) {
			if tni == "" {
				nodes = root.ChildrenKeys()
			} else {
				node := root.PathToNode(tni)
				if node != nil {
					for _, label := range node.ChildrenKeys() {
						nodes = append(nodes, tni+"/"+label)
					}
				}
			}
			return
		},
		func(tni widget.TreeNodeID) bool {
			if node := root.PathToNode(tni); node != nil && node.CountChildren() > 0 {
				return true
			}
			return false
		},
		func(b bool) fyne.CanvasObject {
			key := canvas.NewText("Key", theme.TextColor())
			typ := canvas.NewText("Type", theme.TextColor())
			value := canvas.NewText("Value", theme.TextColor())
			return container.New(layout.NewGridLayout(3), key, typ, value)
		},
		func(tni widget.TreeNodeID, b bool, co fyne.CanvasObject) {
			node := root.PathToNode(tni)
			container, _ := co.(*fyne.Container)
			key := container.Objects[0].(*canvas.Text)
			typ := container.Objects[1].(*canvas.Text)
			value := container.Objects[2].(*canvas.Text)
			if node == nil || &node.entry == nil {
				key.Text = "N/A"
				typ.Text = "N/A"
				value.Text = "N/A"

			} else {
				t, v := GetType(node.entry)
				key.Text = node.entry.key
				typ.Text = t
				value.Text = v.display
			}

		},
	)
	text := canvas.NewText("Please upload a plist file", theme.TextColor())
	text.Alignment = fyne.TextAlignCenter
	text.TextSize = 25
	text.Refresh()

	fileitem := fyne.NewMenuItem("Open", func() {
		filename, err := dialog.File().Filter("Property-List File", "plist").Load()
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}

		file, err := os.Open(filename)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer file.Close()

		content, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}

		_, err = plist.Unmarshal(content, &plistData)
		if err != nil {
			fmt.Println("Error parsing plist data:", err)
			mack.Alert("Error", "Failed to parse plist", "critical")
			w.Close()
			return
		}
		for index, key := range plistData.Keys {
			entry := Parse(key, plistData.Values[index], []string{key})
			entries = append(entries, entry)
		}
		for _, e := range entries {
			t := root.AddChild(e)
			if len(e.children) != 0 {
				for _, c := range e.children {
					node := t.AddChild(c)
					if len(c.children) != 0 {
						for _, m := range c.children {
							ParseChildren(node, m)
						}
					}
				}
			}
		}
		tree.Refresh()
		tree.OpenAllBranches()
		w.SetTitle(filename)
		w.SetContent(tree)
	})

	filemenu := fyne.NewMenu("File", fileitem)
	mainmenu := fyne.NewMainMenu(filemenu)
	w.SetMainMenu(mainmenu)
	resolution := screenresolution.GetPrimary()
	w.Resize(fyne.Size{Width: float32(resolution.Width), Height: float32(resolution.Height)})

	w.SetContent(text)

	w.ShowAndRun()
}
