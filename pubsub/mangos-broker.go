package pubsub

import (
	"fmt"
	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/pub"
	"github.com/go-mangos/mangos/protocol/sub"
	"github.com/go-mangos/mangos/transport/inproc"
	"github.com/go-mangos/mangos/transport/tcp"
	"time"
)

func MangosServe(quiet bool) {
	var receiver mangos.Socket
	var sender mangos.Socket

	sender, _ = pub.NewSocket()
	sender.AddTransport(tcp.NewTransport())
	sender.AddTransport(inproc.NewTransport())
	sender.Listen("tcp://0.0.0.0:40898")

	receiver, _ = sub.NewSocket()
	receiver.SetOption(mangos.OptionSubscribe, []byte{})
	receiver.AddTransport(tcp.NewTransport())
	receiver.AddTransport(inproc.NewTransport())
	receiver.Listen("tcp://0.0.0.0:40899")
	time.Sleep(time.Millisecond * 100)

	last := time.Now()
	messages := 0
	for {
		message, err := receiver.RecvMsg()
		if err != nil {
			fmt.Println(err)
		}
		sender.SendMsg(message)
		if !quiet {
			messages += 1
			now := time.Now()
			if now.Sub(last).Seconds() > 1 {
				println(messages, "msg/sec")
				last = now
				messages = 0
			}
		}
	}
}
