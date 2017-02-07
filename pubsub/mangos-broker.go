package pubsub

import (
	"fmt"
	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/pub"
	"github.com/go-mangos/mangos/protocol/pull"
	"github.com/go-mangos/mangos/transport/ipc"
	"github.com/go-mangos/mangos/transport/tcp"
	"time"
)

func MangosServe(quiet bool) {
	var receiver mangos.Socket
	var sender mangos.Socket

	sender, _ = pub.NewSocket()
	sender.AddTransport(ipc.NewTransport())
	sender.AddTransport(tcp.NewTransport())
	sender.Listen("tcp://0.0.0.0:40898")

	receiver, _ = pull.NewSocket()
	receiver.AddTransport(ipc.NewTransport())
	receiver.AddTransport(tcp.NewTransport())
	receiver.Listen("tcp://0.0.0.0:40899")

	last := time.Now()
	messages := 0
	for {
		message, err := receiver.Recv()
		if err != nil {
			fmt.Println(err)
		}

		sender.Send(message)
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
