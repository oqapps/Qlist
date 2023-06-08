package main

import (
	"fmt"
	"reflect"
	"strings"

	valid "github.com/asaskevich/govalidator"
	"github.com/iancoleman/strcase"
)


type Value struct {
	real    interface{}
	display string
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
