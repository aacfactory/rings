package rings_test

import (
	"fmt"
	"github.com/aacfactory/rings"
	"testing"
)

func TestNew(t *testing.T) {
	r := rings.New[*Item]("1")
	for i := 0; i < 10; i++ {
		_ = r.Push(&Item{
			key:   fmt.Sprintf("%d", i),
			value: i,
		})
	}
	fmt.Println(r)
	fmt.Println(r.Push(&Item{
		key:   fmt.Sprintf("%d", 1),
		value: 1,
	}))
	fmt.Println(r.Get("1"))
	r.Remove("1")
	_, _ = r.Pop()
	_ = r.SeekTo("5")
	fmt.Println(r.Head())
	for i := 0; i < 11; i++ {
		fmt.Print(r.Next(), " ")
	}
	fmt.Println()

}
