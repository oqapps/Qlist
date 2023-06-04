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
	path     []string
	array    bool
	parent   bool
	index    int
}

var plistData plist.OrderedDict

var entries []Entry

func GetEntry(index int) Entry {
	var entry Entry
	for _, i := range entries {
		if i.index == index {
			entry = i
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

func (entry Entry) SetValue(value interface{}) {
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
}

func Parse(key string, data interface{}, path []string) Entry {
	index := len(entries)
	switch v := data.(type) {
	case []interface{}:
		var children []Entry
		for i, item := range v {
			index +=1
			children = append(children, Parse(fmt.Sprintf("%v", i), item, append(path, fmt.Sprintf("%v", i))))
		}
		return Entry{key: key, children: children, index: index, path: path, parent: true}
	case plist.OrderedDict:
		var children []Entry
		for f, a := range v.Keys {
			value := v.Values[f]
			index += 1
			x, ok := value.(plist.OrderedDict)
			if ok {
				var ch []Entry
				for c, k := range x.Keys {
					m := x.Values[c]
					ch = append(ch, Parse(k, m, append(path, k)))
				}

				children = append(children, Entry{key: a, children: ch, path: append(path, a), index: index, parent: true})
			} else {
				m, ok := value.([]interface{})
				if ok {
					var ch []Entry
					for i, a := range m {
						in := strconv.Itoa(i)
						ch = append(ch, Parse(in, a, append(path, in)))
					}
					children = append(children, Entry{key: a, children: ch, path: append(path, a), index: index, array: true, parent: true})
				} else {
					children = append(children, Entry{key: a, value: value, path: append(path, a), index: index})
				}
			}
		}
		return Entry{key: key, children: children, path: path, index: index, parent: true}
	default:
		return Entry{key: key, value: v, path: append(path, key), index: index}
	}
}
