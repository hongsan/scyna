package scyna

import (
	"fmt"
	"log"
	sync "sync"
	"time"

	"google.golang.org/protobuf/proto"
)

type SerialNumber struct {
	key    string
	mutex  sync.Mutex
	prefix uint32
	last   uint64
	next   uint64
}

type generator struct {
	mutex  sync.Mutex
	prefix uint32
	last   uint64
	next   uint64
}

func (g *generator) Reset(prefix uint32, last uint64, next uint64) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.prefix = prefix
	g.last = last
	g.next = next
}

func (g *generator) Next() uint64 {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.next < g.last {
		g.next++
	} else {
		if !g.getID() {
			log.Fatal("Can not create generator")
		}
	}
	return (uint64(g.prefix) << 44) + g.next
}

func (g *generator) InitSN(key string) *SerialNumber {
	return &SerialNumber{
		key:    key,
		prefix: 0,
		last:   0,
		next:   0,
	}
}

func (sn *SerialNumber) Next() string {
	sn.mutex.Lock()
	defer sn.mutex.Unlock()

	if sn.next < sn.last {
		sn.next++
	} else {
		request := GetSNRequest{Key: sn.key}
		var response GetSNResponse
		if r := CallService(GEN_GET_SN_URL, &request, &response); r.Code == 0 {
			sn.prefix = response.Prefix
			sn.next = response.Start
			sn.last = response.End
		} else {
			log.Fatal("Can not get SerialNumber")
		}
	}
	return fmt.Sprintf("%d%07d", sn.prefix, sn.next)
}

func (g *generator) getID() bool {
	callID := uint64(0)

	var req Request
	var res Response

	req.CallID = callID
	req.JSON = false

	reqBytes, err := proto.Marshal(&req)
	if err != nil {
		return false
	}

	msg, respErr := Connection.Request(PublishURL(GEN_GET_ID_URL), reqBytes, 10*time.Second)
	if respErr != nil {
		return false
	}

	err = res.ReadFrom(msg.Data)
	if err != nil {
		return false
	}

	if res.Code == 200 {
		var response GetIDResponse
		if err := proto.Unmarshal(res.Body, &response); err == nil {
			g.prefix = response.Prefix
			g.next = response.Start
			g.last = response.End
			return true
		}
	}
	return false
}
