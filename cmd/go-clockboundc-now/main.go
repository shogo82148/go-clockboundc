package main

import (
	"log"
	"os"

	"github.com/shogo82148/go-clockboundc"
)

func main() {
	var path string
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	if path == "" {
		path = clockboundc.DefaultSocketPath
	}

	c, err := clockboundc.NewWithPath(path)
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
