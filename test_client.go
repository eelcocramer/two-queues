package main

import (
	"./pubsub"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	host        string
	numSeconds  float64
	numClients  int
	numChannels int
	messageSize int
	useRedis    bool
	useMangos   bool
	broker      bool
	quiet       bool
	channels    []string
)

// Returns a new pubsub client instance - either the Redis or ZeroMQ
// client, based on command-line arg.
func NewClient() pubsub.Client {
	var client pubsub.Client
	if useRedis {
		client = pubsub.NewRedisClient(host)
	} else if useMangos {
		var err error
		client, err = pubsub.NewMangosClient(host)
		if err != nil {
			log.Panicln(err)
		}
	} else {
		var err error
		client, err = pubsub.NewZMQClient(host)
		if err != nil {
			log.Panicln(err)
		}
	}
	return client
}

// Loops forever, publishing messages to random channels.
func Publisher() {
	client := NewClient()
	message := strings.Repeat("x", messageSize)
	for {
		channel := channels[rand.Intn(len(channels))]
		if err := client.Publish(channel, message); err != nil {
			log.Panicln(err)
		}
	}
}

// Subscribes to all channels, keeping a count of the number of
// messages received. Publishes and resets the total every second.
func Subscriber() {
	client := NewClient()
	for _, channel := range channels {
		if err := client.Subscribe(channel); err != nil {
			log.Panicln(err)
		}
	}
	last := time.Now()
	messages := 0
	for {
		if _, err := client.Receive(); err != nil {
			log.Panicln(err)
		}
		messages += 1
		now := time.Now()
		if now.Sub(last).Seconds() > 1 {
			if !quiet {
				println(messages, "msg/sec")
			}
			if err := client.Publish("metrics", strconv.Itoa(messages)); err != nil {
				log.Panicln(err)
			}
			last = now
			messages = 0
		}
	}
}

// Creates goroutines * --num-clients, running the given target
// function for each.
func RunWorkers(target func()) {
	for i := 0; i < numClients; i++ {
		go target()
	}
}

// Subscribes to the metrics channel and returns messages from
// it until --num-seconds has passed.
func GetMetrics() []int {
	client := NewClient()
	if err := client.Subscribe("metrics"); err != nil {
		log.Panicln(err)
	}
	metrics := []int{}
	start := time.Now()
	for time.Now().Sub(start).Seconds() <= numSeconds {
		message, err := client.Receive()
		if err != nil {
			log.Panicln(err)
		}
		if message.Type == "message" {
			messages, _ := strconv.Atoi(message.Data)
			metrics = append(metrics, messages)
		}
	}
	return metrics
}

func main() {

	// Set up and parse command-line args.
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.StringVar(&host, "host", "127.0.0.1", "")
	flag.Float64Var(&numSeconds, "num-seconds", 10, "")
	flag.IntVar(&numClients, "num-clients", 1, "")
	flag.IntVar(&numChannels, "num-channels", 50, "")
	flag.IntVar(&messageSize, "message-size", 20, "")
	flag.BoolVar(&useRedis, "redis", false, "")
	flag.BoolVar(&useMangos, "mangos", false, "")
	flag.BoolVar(&quiet, "quiet", false, "")
	flag.BoolVar(&broker, "broker", false, "")
	flag.Parse()
	for i := 0; i < numChannels; i++ {
		channels = append(channels, strconv.Itoa(i))
	}

	// Create publisher/subscriber goroutines, pausing to allow
	// publishers to hit full throttle
	RunWorkers(Publisher)
	time.Sleep(1 * time.Second)
	RunWorkers(Subscriber)

	// Consume metrics until --num-seconds has passed, and display
	// the median value.
	metrics := GetMetrics()
	sort.Ints(metrics)
	fmt.Println(metrics[len(metrics)/2], "median msg/sec")

}
