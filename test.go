package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/pub"
	"github.com/go-mangos/mangos/protocol/sub"
	"github.com/go-mangos/mangos/transport/inproc"
	"github.com/go-mangos/mangos/transport/tcp"
)

var addr1 = "tcp://127.0.0.1:4455"
var addr2 = "tcp://127.0.0.1:4456"

//var addr1 = "inproc://127.0.0.1:4455"
//var addr2 = "inproc://127.0.0.1:4456"

func client(loops int) {
	p, e := pub.NewSocket()
	if e != nil {
		panic(e.Error())
	}
	defer p.Close()
	s, e := sub.NewSocket()
	if e != nil {
		panic(e.Error())
	}
	defer s.Close()

	p.AddTransport(tcp.NewTransport())
	s.AddTransport(tcp.NewTransport())
	p.AddTransport(inproc.NewTransport())
	s.AddTransport(inproc.NewTransport())

	s.SetOption(mangos.OptionSubscribe, []byte{})

	if e = p.Dial(addr1); e != nil {
		panic(e.Error())
	}
	if e = s.Dial(addr2); e != nil {
		panic(e.Error())
	}

	msg := mangos.NewMessage(8)
	msg.Body = append(msg.Body, []byte("hello")...)
	now := time.Now()

	time.Sleep(time.Millisecond * 100)

	for i := 0; i < loops; i++ {
		if e = p.SendMsg(msg); e != nil {
			panic(e.Error())
		}
		if msg, e = s.RecvMsg(); e != nil {
			panic(e.Error())
		}
	}
	end := time.Now()
	delta := float64(end.Sub(now)) / float64(time.Second)

	fmt.Printf("Client %d RTTs in %f secs (%f rtt/sec)\n",
		loops, delta, float64(loops)/delta)
}

func server() {
	p, e := pub.NewSocket()
	if e != nil {
		panic(e.Error())
	}
	defer p.Close()
	s, e := sub.NewSocket()
	if e != nil {
		panic(e.Error())
	}
	defer s.Close()
	p.AddTransport(tcp.NewTransport())
	s.AddTransport(tcp.NewTransport())
	s.AddTransport(inproc.NewTransport())
	p.AddTransport(inproc.NewTransport())
	s.SetOption(mangos.OptionSubscribe, []byte{})

	s.Listen(addr1)
	p.Listen(addr2)

	for {
		msg, e := s.RecvMsg()
		if e != nil {
			println(e.Error())
			return
		}
		e = p.SendMsg(msg)
		if e != nil {
			println(e.Error())
			return
		}
	}
}

func main() {
	clients := 32
	loops := 10000

	//go server()
	time.Sleep(time.Millisecond * 100)

	wg := sync.WaitGroup{}
	wg.Add(clients)

	for i := 0; i < clients; i++ {
		go func() {
			defer wg.Done()
			client(loops)
		}()
	}

	wg.Wait()
}
