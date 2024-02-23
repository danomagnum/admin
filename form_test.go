package gowebstructapi

import (
	"fmt"
	"testing"
)

type testform struct {
	MyBool1 bool
	MyBool2 bool
	MyInt1  int
	MyInt2  int
}

func TestFormCreate(t *testing.T) {

	t1 := testform{
		MyBool1: false,
		MyBool2: true,
		MyInt1:  123,
		MyInt2:  456,
	}

	f := StructToForm(&t1)
	fmt.Printf("result: %v", f)

}
