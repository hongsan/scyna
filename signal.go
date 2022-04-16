package scyna

import (
	"log"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

type SignalHandler func(data []byte)

func RegisterSignal(channel string, handler SignalHandler) {
	_, err := Connection.QueueSubscribe(channel, module, func(m *nats.Msg) {
		handler(m.Data)
	})

	if err != nil {
		log.Fatal("Error in register event")
	}
}

func EmitEmptySignal(channel string) {
	var data []byte
	err := Connection.Publish(channel, data)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func EmitSignal(channel string, event proto.Message) {
	data, err := proto.Marshal(event)
	if err != nil {
		log.Print(err.Error())
	}
	if err := Connection.Publish(channel, data); err != nil {
		log.Print(err.Error())
	}
}
