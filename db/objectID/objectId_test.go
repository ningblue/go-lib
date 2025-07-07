package objectID

import (
	"fmt"
	"testing"
	"time"
)

func TestNewObjectID(t *testing.T) {
	for i := 0; i < 12; i++ {
		time.Sleep(100 * time.Millisecond)
		fmt.Println(NewObjectID().Hex())
	}
}
