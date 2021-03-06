package main

import (
	"./pubsub"
	"flag"
)

func main() {
	var quiet bool
	var mangos bool
	var broker bool
	flag.BoolVar(&quiet, "quiet", false, "")
	flag.BoolVar(&mangos, "mangos", false, "")
	flag.BoolVar(&broker, "broker", false, "")
	flag.Parse()
	if mangos {
		pubsub.MangosServe(quiet)
	} else {
		pubsub.Serve(quiet)
	}
}
