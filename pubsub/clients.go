package pubsub

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	mangos "github.com/go-mangos/mangos"
	pub "github.com/go-mangos/mangos/protocol/pub"
	sub "github.com/go-mangos/mangos/protocol/sub"
	"github.com/go-mangos/mangos/transport/ipc"
	"github.com/go-mangos/mangos/transport/tcp"
	nats "github.com/nats-io/go-nats"
	zmq "github.com/pebbe/zmq4"
	"strings"
	"sync"
	"time"
)

// A pub-sub message - defined to support Redis receiving different
// message types, such as subscribe/unsubscribe info.
type Message struct {
	Type    string
	Channel string
	Data    string
}

// Client interface for both Redis and ZMQ pubsub clients.
type Client interface {
	Subscribe(channels ...interface{}) (err error)
	Unsubscribe(channels ...interface{}) (err error)
	Publish(channel string, message string) error
	Receive() (Message, error)
}

// Redis client - defines the underlying connection and pub-sub
// connections, as well as a mutex for locking write access,
// since this occurs from multiple goroutines.
type RedisClient struct {
	conn redis.Conn
	redis.PubSubConn
	sync.Mutex
}

// ZMQ client - just defines the pub and sub ZMQ sockets.
type ZMQClient struct {
	ctx *zmq.Context
	pub *zmq.Socket
	sub *zmq.Socket
}

// Mangos client - just defines the pub and sub Mangos sockets
type MangosClient struct {
	pub  mangos.Socket
	sub  mangos.Socket
	host string
}

type NatsClient struct {
	conn   *nats.Conn
	pubsub *nats.EncodedConn
	ch     chan *Message
	subs   map[string]*nats.Subscription
	sync.Mutex
}

// Returns a new Redis client. The underlying redigo package uses
// Go's bufio package which will flush the connection when it contains
// enough data to send, but we still need to set up some kind of timed
// flusher, so it's done here with a goroutine.
func NewRedisClient(host string) *RedisClient {
	host = fmt.Sprintf("%s:6379", host)
	conn, _ := redis.Dial("tcp", host)
	pubsub, _ := redis.Dial("tcp", host)
	client := RedisClient{conn, redis.PubSubConn{pubsub}, sync.Mutex{}}
	go func() {
		for {
			time.Sleep(200 * time.Millisecond)
			client.Lock()
			client.conn.Flush()
			client.Unlock()
		}
	}()
	return &client
}

func (client *RedisClient) Publish(channel, message string) error {
	client.Lock()
	client.conn.Send("PUBLISH", channel, message)
	client.Unlock()

	return nil
}

func (client *RedisClient) Receive() (Message, error) {
	switch message := client.PubSubConn.Receive().(type) {
	case redis.Message:
		return Message{"message", message.Channel, string(message.Data)}, nil
	case redis.Subscription:
		return Message{message.Kind, message.Channel, string(message.Count)}, nil
	}
	return Message{}, nil
}

func NewZMQClient(host string) (*ZMQClient, error) {
	var err error
	var context *zmq.Context
	context, err = zmq.NewContext()
	if err != nil {
		return nil, err
	}
	var pub *zmq.Socket
	pub, err = context.NewSocket(zmq.PUSH)
	if err != nil {
		return nil, err
	}
	pub.Connect(fmt.Sprintf("tcp://%s:%d", host, 5562))
	var sub *zmq.Socket
	sub, err = context.NewSocket(zmq.SUB)
	if err != nil {
		return nil, err
	}
	sub.Connect(fmt.Sprintf("tcp://%s:%d", host, 5561))
	return &ZMQClient{context, pub, sub}, nil
}

func (client *ZMQClient) Subscribe(channels ...interface{}) error {
	for _, channel := range channels {
		if err := client.sub.SetSubscribe(channel.(string)); err != nil {
			return err
		}
	}
	return nil
}

func (client *ZMQClient) Unsubscribe(channels ...interface{}) error {
	for _, channel := range channels {
		if err := client.sub.SetUnsubscribe(channel.(string)); err != nil {
			return err
		}
	}
	return nil
}

func (client *ZMQClient) Publish(channel, message string) error {
	_, err := client.pub.Send(channel+" "+message, 0)
	return err
}

func (client *ZMQClient) Receive() (Message, error) {
	message, err := client.sub.Recv(0)
	if err != nil {
		return Message{}, err
	}
	parts := strings.SplitN(string(message), " ", 2)
	return Message{Type: "message", Channel: parts[0], Data: parts[1]}, nil
}

func NewMangosClient(host string) (*MangosClient, error) {
	var err error
	var p mangos.Socket
	p, err = pub.NewSocket()

	if p, err = pub.NewSocket(); err != nil {
		return nil, err
	}

	p.AddTransport(ipc.NewTransport())
	p.AddTransport(tcp.NewTransport())
	p.Dial(fmt.Sprintf("tcp://%s:%d", host, 40899))

	var s mangos.Socket
	s, err = sub.NewSocket()

	if s, err = sub.NewSocket(); err != nil {
		return nil, err
	}

	s.AddTransport(ipc.NewTransport())
	s.AddTransport(tcp.NewTransport())

	time.Sleep(time.Millisecond * 100)

	return &MangosClient{p, s, host}, nil
}

func (client *MangosClient) Subscribe(channels ...interface{}) error {
	for _, channel := range channels {
		if err := client.sub.SetOption(mangos.OptionSubscribe, channel.(string)); err != nil {
			return err
		}
	}

	client.sub.Dial(fmt.Sprintf("tcp://%s:%d", client.host, 40898))
	return nil
}

func (client *MangosClient) Unsubscribe(channels ...interface{}) error {
	for _, channel := range channels {
		if err := client.sub.SetOption(mangos.OptionUnsubscribe, channel.(string)); err != nil {
			return err
		}
	}
	return nil
}

func (client *MangosClient) Publish(channel, message string) error {
	msg := mangos.NewMessage(len(channel + " " + message))
	msg.Body = []byte(channel + " " + message)
	err := client.pub.SendMsg(msg)
	return err
}

func (client *MangosClient) Receive() (Message, error) {
	message, err := client.sub.RecvMsg()
	if err != nil {
		return Message{}, err
	}
	parts := strings.SplitN(string(message.Body), " ", 2)
	return Message{Type: "message", Channel: parts[0], Data: parts[1]}, nil
}

// Returns a new Redis client. The underlying redigo package uses
// Go's bufio package which will flush the connection when it contains
// enough data to send, but we still need to set up some kind of timed
// flusher, so it's done here with a goroutine.
func NewNatsClient(host string) *NatsClient {
	host = fmt.Sprintf("nats://%s:4222", host)
	conn, _ := nats.Connect(host)
	pubsub, _ := nats.NewEncodedConn(conn, nats.JSON_ENCODER)
	client := NatsClient{conn, pubsub, make(chan *Message, 64), make(map[string]*nats.Subscription), sync.Mutex{}}
	go func() {
		for {
			time.Sleep(200 * time.Millisecond)
			client.Lock()
			client.conn.Flush()
			client.Unlock()
		}
	}()
	return &client
}

func (client *NatsClient) Publish(channel, message string) error {
	client.Lock()
	m := &Message{Type: "message", Channel: channel, Data: message}
	client.pubsub.Publish(channel, m)
	client.Unlock()
	return nil
}

func (client *NatsClient) Subscribe(channels ...interface{}) error {
	for _, channel := range channels {
		client.subs[channel.(string)], _ = client.pubsub.Subscribe(channel.(string), func(m *Message) {
			client.ch <- m
		})
	}
	return nil
}

func (client *NatsClient) Unsubscribe(channels ...interface{}) error {
	for _, channel := range channels {
		client.subs[channel.(string)].Unsubscribe()
		delete(client.subs, channel.(string))
	}

	return nil
}

func (client *NatsClient) Receive() (Message, error) {
	message := <-client.ch
	return Message{"message", message.Channel, string(message.Data)}, nil
}
