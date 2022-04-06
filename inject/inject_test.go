package inject

import (
	"fmt"
	"reflect"
	"testing"
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

func TestInterfaceTypeof(t *testing.T) {
	s := reflect.TypeOf((*Initializable)(nil)).Elem()
	fmt.Println(s)
}
