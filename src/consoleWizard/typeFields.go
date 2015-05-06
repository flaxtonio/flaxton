package consoleWizard

import (
	"fmt"
	"os"
	"bufio"
	"strings"
	"strconv"
)

type FieldValueCallback func(value string) bool

func ConsoleHandleField(label string, cb FieldValueCallback) {
	fmt.Println(label)
	for {
		reader := bufio.NewReader(os.Stdin)
		value_input, _ := reader.ReadString('\n')
		value_input = strings.Replace(value_input, "\n", "", -1)
		if !cb(value_input) {
			continue
		}
		break
	}
}

// Field Types structures
type FieldNumber struct {
	Label string	`json:"label"`
	Default int 	`json:"default"`
	Value int 		`json:"value"`
}

func (f *FieldNumber) HandleConsole() {
	fmt.Println(f.Label)
	cb := func(value string) bool {
		if len(value) == 0 {
			f.Value = f.Default
		} else {
			var conv_error error
			f.Value, conv_error = strconv.Atoi(value)
			if conv_error != nil {
				fmt.Println("Please Try again and Input correct Number ")
				return false
			}
		}
		return true
	}
	num_str, _ := strconv.Atoi(f.Default)
	ConsoleHandleField(fmt.Sprintf("%[0] \n Default: [%[1]]", f.Label, num_str), cb)
}

type FieldText struct {
	Label string 	`json:"label"`
	Default string 	`json:"default"`
	Value string  	`json:"value"`
}

func (f *FieldText) HandleConsole() {
	fmt.Println(f.Label)
	cb := func(value string) bool {
		if len(value) == 0 {
			f.Value = f.Default
		} else {
			f.Value = value
		}
		return true
	}
	ConsoleHandleField(fmt.Sprintf("%[0] \n Default: [%[1]]", f.Label, f.Default), cb)
}


type FieldMultiInput struct {
	Label string 		`json:"label"`
	Default []string 	`json:"default"`
	Separator string 	`json:"separator"`
	Values []string 	`json:"values"`
}

func (f *FieldMultiInput) HandleConsole() {
	fmt.Println(f.Label)
	cb := func(value string) bool {
		if len(value) == 0 {
			f.Values = f.Default
		} else {
			if len(f.Separator) > 0 {
				f.Values = strings.Split(value, f.Separator)
			} else {
				f.Values = strings.Split(value, " ")
			}
		}
		return true
	}
	ConsoleHandleField(fmt.Sprintf("%[0] \n Default: [%[1]]", f.Label, f.Default), cb)
}

type FieldChoice struct {
	Label string    			`json:"label"`
	DefaultKey string           `json:"default_key"`
	DefaultValue string         `json:"default_value"`
	Choices map[string]string 	`json:"choices"`
	SelectedKey string 			`json:"selected_key"`
	SelectedValue string 		`json:"selected_value"`
}


func (f *FieldChoice) HandleConsole() {
	fmt.Println(f.Label)
	cb := func(value string) bool {
		if len(value) == 0 {
			f.SelectedValue = f.DefaultValue
			f.SelectedKey = f.DefaultKey
		} else {

		}
		return true
	}
	choices_out := f.Label
	index := 1
	var index_str string
	for _, val := range f.Choices {
		index_str, _ = strconv.Atoi(index)
		choices_out = fmt.Sprintf("%[0] \n %[1]: %[2]", choices_out, index_str, val)
		index++
	}
	choices_out = fmt.Sprintf("%[0] \n Default: %[1]", choices_out, f.DefaultValue)
	ConsoleHandleField(choices_out, cb)
}