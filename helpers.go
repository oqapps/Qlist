package main

import (
	"fmt"
	"reflect"
	"strings"

	valid "github.com/asaskevich/govalidator"
	"github.com/iancoleman/strcase"
)

type treeNode struct {
	entry    Entry
	children []*treeNode
}

type Value struct {
	real interface{}
	display string
}

func ParseChildren(parent *treeNode, entry Entry) {
	node := parent.AddChild(entry)
	if len(entry.children) != 0 {
		for _, m := range entry.children {
			node.AddChild(m)
		}
	}
}

func (n *treeNode) AddChild(entry Entry) *treeNode {
	if n.GetChild(entry.key) != nil {
		return nil
	}

	new := &treeNode{
		entry: entry,
	}
	n.children = append(n.children, new)
	return new
}


func (n *treeNode) ChildrenKeys() (ret []string) {
	for i := 0; i < len(n.children); i++ {
		ret = append(ret, n.children[i].entry.key)
	}
	return
}

func (n *treeNode) GetChild(label string) (ret *treeNode) {
	for i := 0; i < len(n.children); i++ {
		if n.children[i].entry.key == label {
			ret = n.children[i]
			break
		}
	}
	return
}

func (n *treeNode) CountChildren() int {
	return len(n.children)
}

func (n *treeNode) PathToNode(path string) *treeNode {
	currNode := n
	for _, elem := range strings.Split(path, "/") {
		if elem == "" {
			continue
		}
		currNode = currNode.GetChild(elem)
		if currNode == nil {
			break
		}p
	}
	return currNode
}

func dataString(data []byte) string {
    var builder strings.Builder
    builder.WriteString("<")
    for i, b := range data {
        builder.WriteString(fmt.Sprintf("%02X", b))
        if (i+1)%4 == 0 && i != len(data)-1 {
            builder.WriteString(" ")
        }
    }
    builder.WriteString(">")
    return builder.String()
}

func GetType(entry Entry) (string, Value) {
	var t string
	var value Value
	if entry.value == nil {
		if entry.parent {
			if entry.array {
				t = "Array"
				value = Value{display:fmt.Sprintf("%v children", len(entry.children))}
				if len(entry.children) == 1 {
					value.display = "1 child"
				}
			} else {
				t = "Dictionary"
				value = Value{display:fmt.Sprintf("%v dictionaries", len(entry.children))}
				if len(entry.children) == 1 {
					value.display = "1 dictionary"
				}
			}
		} else {
			t = "String"
			value = Value{display:fmt.Sprintf("%v", entry.value)}
		}
	} else {
		value = Value{display:fmt.Sprintf("%v", entry.value),real:entry.value}
		switch reflect.TypeOf(entry.value).Name() {
		case "bool":
			{
				t = "Boolean"
				value.display = strcase.ToCamel(value.display)
			}
		case "string":
			{
				t = "String"
			}
		default:
			{
				if valid.IsInt(value.display) {
					t = "Number"
				} else {
					data := entry.value.([]uint8)
					
					value.display = dataString(data)
					t = "Data"
				}	
			}
		}
	}
	return t, value
}
