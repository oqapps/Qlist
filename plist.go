package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oq-x/go-plist"
)

type Entry struct {
	key           string
	value         interface{}
	children      []Entry
	childrenPaths []string
	path          string
	array         bool
	isParent      bool
	index         int
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

func ValueToDict(entry Entry) plist.OrderedDict {
	dict := plist.OrderedDict{}
	for _, e := range entry.children {
		dict.Keys = append(dict.Keys, e.key)
		if e.isParent {
			dict.Values = append(dict.Values, ValueToDict(e))
		} else {
			dict.Values = append(dict.Values, e.value)
		}
	}
	return dict
}

func ParseEntries(entries Entries) plist.OrderedDict {
	dict := plist.OrderedDict{}
	for _, entry := range entries {
		dict.Keys = append(dict.Keys, entry.key)
		if entry.isParent {
			dict.Values = append(dict.Values, ValueToDict(entry))
		} else {
			dict.Values = append(dict.Values, entry.value)
		}
	}
	return dict
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
}*/

func (entry Entry) SetKey(entries Entries, value string) {
	entry.key = value
	pathSp := strings.Split(entry.path, "\\-\\")
	pathSp[len(pathSp)-1] = entry.key
	entry.path = strings.Join(pathSp, "\\-\\")
	entries[entry.path] = entry
}

func AppendToPath(path string, index ...string) string {
	for _, p := range index {
		path += "\\-\\" + p
	}
	return path
}

func Parse(key string, data interface{}, path string, index int, entries Entries) Entry {
	var children []Entry
	var childrenPaths []string
	switch v := data.(type) {
	case []interface{}:
		for i, item := range v {
			children = append(children, Parse(fmt.Sprintf("%d", i), item, AppendToPath(path, fmt.Sprintf("%d", i)), index, entries))
			childrenPaths = append(childrenPaths, AppendToPath(path, fmt.Sprintf("%d", i)))
		}
		entry := Entry{key: key, children: children, index: index, path: path, isParent: true, childrenPaths: childrenPaths}
		entries[path] = entry
		return entry
	case plist.OrderedDict:
		for f, a := range v.Keys {
			p := AppendToPath(path, a)
			value := v.Values[f]
			x, ok := value.(plist.OrderedDict)
			var ch []Entry
			var chp []string
			if ok {
				for c, k := range x.Keys {
					m := x.Values[c]
					ch = append(ch, Parse(k, m, AppendToPath(p, k), index, entries))
					chp = append(chp, AppendToPath(p, k))
				}
				e := Entry{key: a, children: ch, path: p, index: f, isParent: true, childrenPaths: chp}
				children = append(children, e)
				childrenPaths = append(childrenPaths, p)
				entries[p] = e
			} else {
				m, ok := value.([]interface{})
				if ok {
					for i, q := range m {
						in := strconv.Itoa(i)
						ch = append(ch, Parse(in, q, AppendToPath(p, in), index, entries))
						chp = append(chp, AppendToPath(p, in))
					}
					e := Entry{key: a, children: ch, path: p, index: f, array: true, isParent: true, childrenPaths: chp}
					children = append(children, e)
					childrenPaths = append(childrenPaths, p)
					entries[p] = e
				} else {
					e := Entry{key: a, value: value, path: p, index: f}
					children = append(children, e)
					childrenPaths = append(childrenPaths, p)
					entries[p] = e
				}
			}
		}
		entry := Entry{key: key, children: children, childrenPaths: childrenPaths, path: path, index: index, isParent: true}
		entries[path] = entry
		return entry
	default:
		entry := Entry{key: key, value: v, path: path, index: index}
		entries[path] = entry
		return entry
	}
}
