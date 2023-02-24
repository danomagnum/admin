package main

import (
	"html/template"
	"net/url"
	"time"
)

type Options1 int

func (o Options1) Enumerate() map[string]any {
	return map[string]any{
		"unknown": 0,
		"a":       1,
		"b":       2,
		"c":       3,
	}
}

type SubConfig struct {
	Name   string
	Length float32
	OK     bool
}

type ConfigType1 struct {
	Name   string
	Param1 int
	Param2 float32
	Param3 time.Time
	Param4 time.Duration
	Param5 bool
	Param6 Options1
	Param7 []int
	Param8 []SubConfig
}

func (t ConfigType1) Handle() string {
	return t.Name
}
func (t ConfigType1) String() string {
	return t.Name
}
func (t ConfigType1) Update(url.Values) error {
	return nil
}
func (t ConfigType1) RenderConfig() template.HTML {
	return ""
}
