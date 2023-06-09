package main

import (
	"fmt"
	"strconv"

	"github.com/deitrix/go-plist"
)

type Entry struct {
	key      string
	value    interface{}
	children []Entry
	path     string
	array    bool
	isParent bool
	index    int
}

var plistData plist.OrderedDict
var arrayPlist []interface{}

type Entries = map[string]Entry

func GetEntry(entries Entries, index int) Entry {
	var entry Entry
	for _, e := range entries {
		if e.index == index {
			entry = e
			break
		}
	}
	return entry
}

func Get(dict plist.OrderedDict, key string) interface{} {
	var index int
	for i, k := range dict.Keys {
		if k == key {
			index = i
			break
		}
	}
	return dict.Values[index]
}

/*func (entry Entry) SetValue(value interface{}) {
	entry.value = value
	var element = Get(plistData, entry.path[0])
	for i, p := range entry.path {
		if i == 0 {
			continue
		}
		element = (element.(map[string]interface{}))[p]
	}
}

func (entry Entry) SetKey(value string) {
	var parentEntry Entry
	for _, entry := range entries {
		var p []string
		for index, value := range entry.path {
			if index == len(entry.path)-1 {
				continue
			} else {
				p = append(p, value)
			}
		}
		for index, value := range p {
			if entry.path[index] != value {
				return
			}
		}
		parentEntry = entry
	}
	for index, child := range parentEntry.children {
		if child.key == entry.key {
			parentEntry.children[index] = child
			entries[parentEntry.index] = parentEntry
			break
		}
	}
	var element = Get(plistData, entry.path[0])
	for index, p := range entry.path {
		if index == 0 {
			continue
		}
		if index == len(entry.path)-1 {
			continue
		}

		element, err := element.(map[string]interface{})
		if !err {
			break
		} else {
			element = element[p].(map[string]interface{})
		}
	}
	data := (element.(map[string]interface{}))[entry.key]
	delete(element.(map[string]interface{}), entry.key)
	(element.(map[string]interface{}))[value] = data
}*/

func AppendToPath(path string, index ...int) string {
	for _, p := range index {
		i := strconv.Itoa(p)
		path += "\\-\\" + i
	}
	return path
}

func Parse(key string, data interface{}, path string, index int, entries Entries) Entry {
	switch v := data.(type) {
	case []interface{}:
		var children []Entry
		for i, item := range v {
			children = append(children, Parse(fmt.Sprintf("%v", i), item, AppendToPath(path, i), index, entries))
		}
		entry := Entry{key: key, children: children, index: index, path: path, isParent: true}
		entries[path] = entry
		return entry
	case plist.OrderedDict:
		var children []Entry
		for f, a := range v.Keys {
			p := AppendToPath(path, f)
			value := v.Values[f]
			x, ok := value.(plist.OrderedDict)
			if ok {
				var ch []Entry
				for c, k := range x.Keys {
					m := x.Values[c]
					ch = append(ch, Parse(k, m, AppendToPath(p, c), index, entries))
				}
				e := Entry{key: a, children: ch, path: p, index: f, isParent: true}
				children = append(children, e)
				entries[p] = e
			} else {
				m, ok := value.([]interface{})
				if ok {
					var ch []Entry
					for i, q := range m {
						in := strconv.Itoa(i)
						ch = append(ch, Parse(in, q, AppendToPath(p, i), index, entries))
					}
					e := Entry{key: a, children: ch, path: p, index: f, array: true, isParent: true}
					children = append(children, e)
					entries[p] = e
				} else {
					e := Entry{key: a, value: value, path: p, index: f}
					children = append(children, e)
					entries[p] = e
				}
			}
		}
		entry := Entry{key: key, children: children, path: path, index: index, isParent: true}
		entries[path] = entry
		return entry
	default:
		entry := Entry{key: key, value: v, path: path, index: index}
		entries[path] = entry
		return entry
	}
}
