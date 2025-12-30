package main

import (
	"log"
	"time"
)

func main() {
	log.Println("worker started")
	for {
		log.Println("worker tick (placeholder)")
		time.Sleep(5 * time.Second)
	}
}
