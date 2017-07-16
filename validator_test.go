package just

import (
	"fmt"
	"testing"
)

type validationStructTest01 struct {
	IntValue      int     `valid:"min(1);max(100)"`
	FloatValue    float64 `valid:"min(1);max(100)"`
	UuidValue     string  `valid:"uuid"`
	IntStrValue   string  `valid:"int"`
	FloatStrValue string  `valid:"float"`
	BoolStrValue  string  `valid:"bool"`
}

func TestValidation(t *testing.T) {
	v := validationStructTest01{
		IntValue:      10,
		FloatValue:    10,
		UuidValue:     "00002a37-0000-1000-8000-00805f9b34fb",
		IntStrValue:   "10",
		FloatStrValue: "10.1",
		BoolStrValue:  "true",
	}
	if list := Validation(&v); len(list) > 0 {
		fmt.Println(list)
		t.Fail()
	}
}
