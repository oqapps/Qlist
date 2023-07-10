package main

import (
	"fmt"

	"github.com/fatih/color"
)

type Logger struct {
	FilePath string
}

func (logger Logger) Info(data ...interface{}) {
	blue := color.New(color.BgBlue).Add(color.FgWhite).Add(color.Bold).SprintFunc()
	str := ""
	for _, d := range data {
		str += fmt.Sprintf("%v ", d)
	}
	fmt.Printf("%s %s\n", blue("INFO"), str)
}

func (logger Logger) Error(data ...interface{}) {
	red := color.New(color.BgRed).Add(color.FgWhite).Add(color.Bold).SprintFunc()
	str := ""
	for _, d := range data {
		str += fmt.Sprintf("%v ", d)
	}
	fmt.Printf("%s %s\n", red("ERROR"), str)
}
