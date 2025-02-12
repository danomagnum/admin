package admin

import (
	"encoding/json"
	"log"
	"reflect"
	"testing"
	"time"
)

type Config struct {
	Thing1a testResolveSubStr `json:"thing1a"`
	Thing1b testResolveSubStr `json:"thing1b"`
}

var config Config

type Options1 int

func (o Options1) Enumerate() map[string]any {
	return map[string]any{
		"unknown": 0,
		"a":       1,
		"b":       2,
		"c":       3,
	}
}

type testResolveSubStr2 struct {
	Name   string  `json:"name"`
	Length float32 `json:"length"`
	OK     bool    `json:"ok"`
}

type testResolveSubStr struct {
	Name   string               `json:"name"`
	Param1 int                  `json:"param1"`
	Param2 float32              `json:"param2"`
	Param3 time.Time            `json:"param3"`
	Param4 time.Duration        `json:"param4"`
	Param5 bool                 `json:"param5"`
	Param6 Options1             `json:"param6"`
	Param7 []int                `json:"param7"`
	Param8 []testResolveSubStr2 `json:"param8"`
}

func TestResolve(t *testing.T) {
	var c Config
	c.Thing1a.Name = "not abcd"
	c.Thing1a.Param7 = make([]int, 10)
	c.Thing1b.Param7 = make([]int, 5)

	//var c2 Config
	//c2.Thing1a.Name = "abcd"
	var c2 testResolveSubStr
	c2.Name = "abcd"

	//v := Viewer{Root: c}

	/*
		x, err := v.ResolvePath("")
		if err != nil {
			t.Errorf("failed to resolve path at all. %v", err)
		}
		log.Print(x.CanSet())
		log.Print(x.CanAddr())
	*/
	x := testResolvefunc(&c)

	d, err := json.Marshal(c2)
	if err != nil {
		t.Errorf("failed to marshal. %v", err)
	}

	log.Printf("before: %+v", c)
	err = json.Unmarshal(d, x)
	if err != nil {
		t.Errorf("failed to unmarshal. %v", err)
	}
	log.Printf("after: %+v", c)

}

func testResolvefunc(t any) any {
	v := reflect.ValueOf(t).Elem()

	n := v.NumField()
	log.Printf("%d fields", n)

	f := v.FieldByName("Thing1a")

	return f.Addr().Interface()
	//return &t
}
