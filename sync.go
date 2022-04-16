package scyna

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
)

type SyncHandler func(data []byte) *http.Request

func RegisterSync(channel string, consumer string, group string, handler SyncHandler) {
	LOG.Info(fmt.Sprintf("channel %s, consummer: %s, group: %s", channel, consumer, group))
	_, err := JetStream.QueueSubscribe(channel, group, func(m *nats.Msg) {
		request := handler(m.Data)
		if sendRequest(request) {
			m.Ack()
		} else {
			for i := 0; i < 3; i++ {
				request := handler(m.Data)
				if sendRequest(request) {
					m.Ack()
					return
				}
				time.Sleep(time.Second * 30)
			}
			time.Sleep(time.Minute * 10)
			m.Nak()
		}
	}, nats.Durable(consumer), nats.ManualAck())

	if err != nil {
		log.Fatal("JetStream Error: ", err)
	}
}

func sendRequest(request *http.Request) bool {
	if request == nil {
		return true
	}

	response, err := HttpClient().Do(request)
	if err != nil {
		LOG.Warning("Sync:" + err.Error())
		return false
	} else {
		defer response.Body.Close()
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			LOG.Info("Sync error: " + err.Error())
			return true
		}
		bodyString := string(bodyBytes)
		LOG.Info(fmt.Sprintf("Sync: %s - %d - %s", request.URL, response.StatusCode, bodyString))

		if response.StatusCode == 500 {
			return false
		}
	}
	return true
}
