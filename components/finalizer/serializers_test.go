package finalizer

import (
	"fmt"
	"testing"
	"time"

	"github.com/itrabbit/just"
)

type testStruct0 struct {
	Value string `json:"value" xml:"value"`
}

type testStruct1 struct {
	ID          uint64       `json:"-" xml:"-"`
	UUID        string       `json:"uuid,omitempty" xml:"uuid,omitempty"`
	PhoneNumber *testStruct0 `json:"phone_number,omitempty" xml:"phone_number,omitempty" group:"moderate" export:"Value"`
	FirstName   string       `json:"first_name,omitempty" xml:"first_name,omitempty"`
	LastName    string       `json:"last_name,omitempty" xml:"last_name,omitempty"`
	MiddleName  string       `json:"middle_name,omitempty" xml:"middle_name,omitempty"`
	CreatedAt   time.Time    `json:"created_at,omitempty" xml:"created_at,omitempty" group:"moderate"`
	UpdatedAt   time.Time    `json:"updated_at,omitempty" xml:"updated_at,omitempty" group:"moderate" exclude:"equal:CreatedAt"`
}

func TestJsonSerializer_Serialize(t *testing.T) {
	// Отключаем режим отладки
	just.SetDebugMode(false)

	users := make([]*testStruct1, 10, 10)
	now := time.Unix(1024, 0)
	for i := 0; i < 10; i++ {
		users[i] = &testStruct1{
			ID:   uint64(i + 1),
			UUID: "0000-0000-00000000",
			PhoneNumber: &testStruct0{
				Value: "+7900000000",
			},
			FirstName: "Alex",
			LastName:  "Grimm",
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	s := NewJsonSerializer("utf-8")
	d, err := s.Serialize(Input(users, "moderate"))
	if err != nil {
		t.Error(err)
		return
	}
	if len(d) != 1401 {
		fmt.Println("1401 !=", len(d))
		fmt.Println(string(d))
		t.Fail()
	}
}

func TestXmlSerializer_Serialize(t *testing.T) {
	// Отключаем режим отладки
	just.SetDebugMode(false)

	users := make([]*testStruct1, 10, 10)
	now := time.Unix(1024, 0)
	for i := 0; i < 10; i++ {
		users[i] = &testStruct1{
			ID:   uint64(i + 1),
			UUID: "0000-0000-00000000",
			PhoneNumber: &testStruct0{
				Value: "+7900000000",
			},
			FirstName: "Alex",
			LastName:  "Grimm",
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	s := NewXmlSerializer("utf-8")
	d, err := s.Serialize(Input(users, "moderate"))
	if err != nil {
		t.Error(err)
		return
	}
	if len(d) != 1910 {
		fmt.Println("1910 !=", len(d))
		fmt.Println(string(d))
		t.Fail()
	}
}
