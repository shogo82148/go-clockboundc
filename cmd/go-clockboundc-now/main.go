package main

import (
	"log"

	"github.com/shogo82148/go-clockboundc"
)

func main() {
	c, err := clockboundc.New()
	if err != nil {
		log.Fatal(err)
	}
	now, err := c.Now()
	if err != nil {
		log.Fatal(err)
	}

	if now.Header.Unsynchronized {
		log.Println("Unsynchronized")
	} else {
		log.Println("Synchronized")
	}
	log.Println("Current: ", now.Time)
	log.Println("Earliest:", now.Bound.Earliest)
	log.Println("Latest:  ", now.Bound.Latest)
	log.Println("Range:   ", now.Bound.Latest.Sub(now.Bound.Earliest))

	if err := c.Close(); err != nil {
		log.Fatal(err)
	}
}
