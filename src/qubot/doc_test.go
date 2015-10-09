package qubot

import (
	"fmt"
	"os"
	"os/signal"
)

func ExampleQubot() {
	q := Init(&Config{})
	q.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	select {
	case <-sigChan:
		q.Close()
	case <-q.Done():
		fmt.Println("Bye!")
	}
}
