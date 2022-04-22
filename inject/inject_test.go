package inject

import (
	"fmt"
	"testing"

	"github.com/stella-go/logger"
	"github.com/stella-go/siu/common"
)

type SS struct {
	SSint      int
	SSintslice *[]int `@siu:""`
}

func (s *SS) String() string {
	return fmt.Sprintf("{%v, %v}", s.SSint, s.SSintslice)
}

func (s *SS) Init() {
	fmt.Printf("*SS Init\n")
}

type S struct {
	SprtSS *SS  `@siu:"name='abc',default='zero'"`
	SB     bool `@siu:"value='${a.b.c}'"`
}

func (s *S) String() string {
	return fmt.Sprintf("{%v, %v}", s.SprtSS, s.SB)
}

func (s *S) Init() {
	fmt.Printf("*S Init\n")
}

type C struct{}

func (c *C) Resolve(key string) (interface{}, bool) {
	return true, true
}

func TestInject(t *testing.T) {
	ints := []int{1, 2, 3}
	RegisterNamed("abc", &SS{999, &ints})
	s := &S{}
	fmt.Println(s)
	c := &C{}
	err := Inject(c, s)
	if err != nil {
		fmt.Println("err:", err)
	}
	fmt.Println(s)
}

type Resolver struct{}

func (r *Resolver) Resolve(key string) (interface{}, bool) {
	m := map[string]string{
		"a.b.c": "123",
	}
	v, ok := m[key]
	return v, ok
}

func TestValue(t *testing.T) {
	common.SetLevel(logger.DebugLevel)
	{
		r := &Resolver{}
		type St struct {
			S string `@siu:"value='${a.b.c}'"`
		}
		st := &St{}
		err := Inject(r, st)
		if err != nil {
			t.Fatal(err)
		}
		if st.S != "123" {
			t.FailNow()
		}
	}
	{
		r := &Resolver{}
		type St struct {
			S string `@siu:"value='${abc}'"`
		}
		st := &St{}
		err := Inject(r, st)
		if err == nil {
			t.FailNow()
		}
	}
	{
		r := &Resolver{}
		type St struct {
			S string `@siu:"value='${abc:999}'"`
		}
		st := &St{}
		err := Inject(r, st)
		if err != nil {
			t.Fatal(err)
		}
		if st.S != "999" {
			t.FailNow()
		}
	}
	{
		r := &Resolver{}
		type St struct {
			S string `@siu:"value='${abc}',default='999'"`
		}
		st := &St{}
		err := Inject(r, st)
		if err != nil {
			t.Fatal(err)
		}
		if st.S != "999" {
			t.FailNow()
		}
	}
	{
		r := &Resolver{}
		type St struct {
			S string `@siu:"default='999'"`
		}
		st := &St{}
		err := Inject(r, st)
		if err != nil {
			t.Fatal(err)
		}
		if st.S != "999" {
			t.FailNow()
		}
	}
	{
		r := &Resolver{}
		type St struct {
			S string `@siu:"value='${}',default='999'"`
		}
		st := &St{}
		err := Inject(r, st)
		if err != nil {
			t.Fatal(err)
		}
		if st.S != "999" {
			t.FailNow()
		}
	}
	{
		r := &Resolver{}
		type St struct {
			S string `@siu:""`
		}
		st := &St{}
		err := Inject(r, st)
		if err == nil {
			t.FailNow()
		}
	}
	{
		r := &Resolver{}
		type St struct {
			S string `@siu:"value='${abc:666}',default='999'"`
		}
		st := &St{}
		err := Inject(r, st)
		if err != nil {
			t.Fatal(err)
		}
		if st.S != "666" {
			t.FailNow()
		}
	}
}
