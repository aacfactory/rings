package rings_test

import (
	"fmt"
	"github.com/aacfactory/rings"
	"testing"
)

type Item struct {
	key   string
	value int
}

func (item *Item) Key() (key string) {
	key = item.key
	return
}

func (item *Item) String() (s string) {
	s = fmt.Sprintf("%s:%d", item.key, item.value)
	return
}

func createItems(n int) (items []*Item) {
	items = make([]*Item, 0, 1)
	for i := 0; i < n; i++ {
		items = append(items, &Item{
			key:   fmt.Sprintf("%d", i),
			value: i,
		})
	}
	return
}

func TestNewImmutable(t *testing.T) {
	items := createItems(5)
	r := rings.NewImmutable[*Item]("immutable", items)
	for i := 0; i < 9; i++ {
		fmt.Println(i, "->", r.Next().String())
	}
	fmt.Println(r.Get("0"))
	fmt.Println(r)
}
