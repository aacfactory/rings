package rings_test

import (
	"fmt"
	"github.com/aacfactory/rings"
	"math/rand"
	"testing"
)

func TestHashed(t *testing.T) {
	hashed := rings.NewHashed(
		&Item{
			key:   "1",
			value: 1,
		},
		&Item{
			key:   "2",
			value: 2,
		},
		&Item{
			key:   "3",
			value: 3,
		},
	)
	fmt.Println(hashed)
	for i := 0; i < 5; i++ {
		key := []byte(fmt.Sprintf("%d", i*rand.Intn(9999999)))
		fmt.Print(hashed.Get(key))
		fmt.Print(", ")
	}
	fmt.Println()
	prev, cLow, cHigh, active, cancel, ok := hashed.Add(&Item{
		key:   "4",
		value: 4,
	})
	if !ok {
		fmt.Println(ok)
		return
	}
	fmt.Println("prev:", prev, cLow, cHigh)
	fmt.Println(hashed)
	active()
	fmt.Println(hashed)
	cancel()
	fmt.Println(hashed)
	fmt.Println("-----")
	for i := 0; i < 5; i++ {
		_, cLow, cHigh, active, _, _ = hashed.Add(&Item{
			key:   fmt.Sprintf("%d", 4+i),
			value: 4 + i,
		})
		active()
		fmt.Println(hashed, cLow, cHigh)
	}
}
