package main

import (
	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/pub"
	"github.com/go-mangos/mangos/protocol/sub"
	"github.com/go-mangos/mangos/transport/inproc"
	"github.com/go-mangos/mangos/transport/tcp"
)

var addr1 = "tcp://127.0.0.1:4455"
var addr2 = "tcp://127.0.0.1:4456"

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
	server()
}
