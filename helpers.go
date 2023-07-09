package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	valid "github.com/asaskevich/govalidator"
	"github.com/iancoleman/strcase"
)

type Value struct {
	real    interface{}
	display string
}

type Display struct {
	value    Value
	typeText string
}

type Manager struct {
	Displays map[string]Display
}

func (manager Manager) Type() string {
	if len(arrayPlist) != 0 {
		return "array"
	} else if len(plistData.Keys) != 0 {
		return "dict"
	} else {
		return "unknown"
	}
}

func (manager Manager) Length() int {
	switch manager.Type() {
	case "array":
		return len(arrayPlist)
	case "dict":
		return len(plistData.Keys)
	default:
		return 0
	}
}

func (manager Manager) Keys() []string {
	children := []string{}
	switch manager.Type() {
	case "array":
		for i := 0; i < len(arrayPlist); i++ {
			children = append(children, strconv.Itoa(i))
		}
		return children
	case "dict":
		return plistData.Keys
	default:
		return children
	}
}

func (manager Manager) Display(entry Entry) Display {
	if manager.Displays == nil {
		manager.Displays = make(map[string]Display)
	}
	display := manager.Displays[entry.path]
	if display.typeText == "" {
		typeText, value := GetType(entry)
		display = Display{typeText: typeText, value: value}
		manager.Displays[entry.path] = display
	}
	return display
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
		if entry.isParent {
			if entry.array {
				t = "Array"
				value = Value{display: fmt.Sprintf("%v children", len(entry.children))}
				if len(entry.children) == 1 {
					value.display = "1 child"
				}
			} else {
				t = "Dictionary"
				value = Value{display: fmt.Sprintf("%v key/value entries", len(entry.children))}
				if len(entry.children) == 1 {
					value.display = "1 key/value entry"
				}
			}
		} else {
			t = "String"
			value = Value{display: fmt.Sprintf("%v", entry.value)}
		}
	} else {
		value = Value{display: fmt.Sprintf("%v", entry.value), real: entry.value}
		_, isBool := entry.value.(bool)
		_, isString := entry.value.(string)
		ti, isDate := entry.value.(time.Time)
		if isBool {
			t = "Boolean"
			value.display = strcase.ToCamel(value.display)
		} else if isString {
			t = "String"
		} else if isDate {
			value.display = ti.String()
			t = "Date"
		} else {
			if valid.IsInt(value.display) {
				t = "Number"
			} else {
				value.display = dataString(entry.value.([]uint8))
				t = "Data"
			}
		}
	}
	return t, value
}
