package wait

import (
	"log"
	"time"
)

func For(name string, fn func() error) {
	for i := 1; ; i++ {
		if err := fn(); err == nil {
			return
		} else if i == 30 {
			log.Fatalf("%s not ready: %v", name, err)
		} else {
			log.Printf("waiting for %s...", name)
			time.Sleep(time.Second)
		}
	}
}
