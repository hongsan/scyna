package scyna

import (
	"log"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

type EventHandler func(data []byte)

func RegisterEvent(channel string, consumer string, handler EventHandler) {
	_, err := JetStream.QueueSubscribe(channel, module, func(m *nats.Msg) {
		handler(m.Data)
	}, nats.Durable(consumer))

	if err != nil {
		log.Fatal("JetStream Error: ", err)
	}
}

func PostEmptyEvent(channel string) {
	if _, err := JetStream.Publish(channel, nil); err != nil {
		log.Print(err.Error())
	}
}

func PostEvent(channel string, event proto.Message) {
	data, err := proto.Marshal(event)
	if err != nil {
		log.Print(err.Error())
	}
	if _, err := JetStream.Publish(channel, data); err != nil {
		log.Print(err.Error())
	}
}
