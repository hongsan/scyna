package scyna

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

type ServiceHandler func(s *Service)
type Service struct {
	Request  Request
	Response Response
	Reply    string
	LOG      Logger
}

type ServicePool struct {
	sync.Pool
}

func (service *Service) reset() {
	service.Request.Body = service.Request.Body[0:0]
	service.Request.CallID = uint64(0)
	service.Response.Body = service.Response.Body[0:0]
	service.Response.Code = int32(0)
	service.Response.SessionID = uint64(0)
	service.LOG.reset(0)
}

func NewService() *Service {
	return &Service{
		Request: Request{
			Body:   make([]byte, 4096),
			CallID: 0,
		},
		Response: Response{
			Body:      make([]byte, 0),
			SessionID: 0,
			Code:      200,
		},
		LOG: &logger{false, 0},
	}
}

func (p *ServicePool) GetService() *Service {
	service, _ := p.Get().(*Service)
	return service
}

func (p *ServicePool) PutService(service *Service) {
	service.reset()
	p.Put(service)
}

func NewServicePool() ServicePool {
	return ServicePool{
		sync.Pool{
			New: func() interface{} { return NewService() },
		}}
}

func RegisterService(url string, handler ServiceHandler) {
	subscribeURL := SubscribreURL(url)
	log.Println("[Register] Sub url: ", url)
	_, err := Connection.QueueSubscribe(subscribeURL, "API", func(m *nats.Msg) {
		service := Services.GetService()
		defer Services.PutService(service)

		if err := proto.Unmarshal(m.Data, &service.Request); err != nil {
			log.Print("Register unmarshal error response data:", err.Error())
			return
		}

		service.Reply = m.Reply
		handler(service)
	})

	if err != nil {
		log.Fatal("Can not register service:", url)
	}
}

func CallService(url string, request proto.Message, response proto.Message) *Error {
	callID := ID.Next()
	start := time.Now()
	day := GetDayByTime(start)
	clientID := module

	var req Request
	var res Response

	if request != nil {
		var err error
		if req.Body, err = proto.Marshal(request); err != nil {
			saveErrorCall(clientID, 400, callID, day, start, url)
			return BAD_REQUEST
		}
	}

	req.CallID = callID
	req.JSON = false

	reqBytes, err := proto.Marshal(&req)
	if err != nil {
		saveErrorCall(clientID, 400, callID, day, start, url)
		return BAD_REQUEST
	}

	msg, respErr := Connection.Request(PublishURL(url), reqBytes, 10*time.Second)
	if respErr != nil {
		saveErrorCall(clientID, 400, callID, day, start, url)
		return SERVER_ERROR
	}

	err = res.ReadFrom(msg.Data)
	if err != nil {
		saveErrorCall(clientID, 400, callID, day, start, url)
		return SERVER_ERROR
	}

	if res.Code == 200 {
		if err := proto.Unmarshal(res.Body, response); err == nil {
			duration := time.Now().UnixMicro() - start.UnixMicro()
			saveCall(clientID, callID, day, start, duration, url, &req, &res)
			return OK
		}
	} else {
		var ret Error
		if err := proto.Unmarshal(res.Body, &ret); err == nil {
			saveErrorCall(clientID, 400, callID, day, start, url)
			return &ret
		}
	}
	saveErrorCall(clientID, 400, callID, day, start, url)
	return SERVER_ERROR
}

func (s *Service) Parse(request proto.Message) bool {
	if s.Request.JSON {
		if err := json.Unmarshal(s.Request.Body, request); err != nil {
			log.Print("Bad Request: " + err.Error())
			s.Error(BAD_REQUEST)
			return false
		}
	} else {
		if err := proto.Unmarshal(s.Request.Body, request); err != nil {
			log.Print("Bad Request: " + err.Error())
			s.Error(BAD_REQUEST)
			return false
		}
	}
	s.LOG.reset(s.Request.CallID)
	return true
}

func (s *Service) Done(response proto.Message) {
	s.Response.Code = 200
	var err error
	if s.Request.JSON {
		s.Response.Body, err = json.Marshal(response)
	} else {
		s.Response.Body, err = proto.Marshal(response)
	}
	if err != nil {
		s.writeError(err)
	}
	s.flushResponse()
}

func (s *Service) Error(e *Error) {
	s.Response.Code = 400
	var err error
	if s.Request.JSON {
		s.Response.Body, err = json.Marshal(e)
	} else {
		s.Response.Body, err = proto.Marshal(e)
	}
	if err != nil {
		s.writeError(err)
	}
	s.flushResponse()
}

func (s *Service) writeError(err error) {
	s.Response.Code = int32(500)
	s.Response.Body = []byte(err.Error())
}

func (s *Service) flushResponse() {
	s.Response.SessionID = Session.ID()
	bytes, err := proto.Marshal(&s.Response)
	if err != nil {
		log.Print("Register marshal error response data:", err.Error())
		return
	}
	err = Connection.Publish(s.Reply, bytes)
	if err != nil {
		LOG.Error(fmt.Sprintf("Nats publish to [%s] error: %s", s.Reply, err.Error()))
	}
}

func (r *Request) Build(req *http.Request) error {
	if req == nil {
		return errors.New("natsproxy: Request cannot be nil")
	}

	buf := bytes.NewBuffer(r.Body)
	buf.Reset()
	if req.Body != nil {
		if _, err := io.Copy(buf, req.Body); err != nil {
			return err
		}
		if err := req.Body.Close(); err != nil {
			return err
		}
	}

	r.Body = buf.Bytes()
	return nil
}

func (r *Response) ReadFrom(responseData []byte) error {
	if len(responseData) == 0 {
		return errors.New("natsproxy: No response content found")
	}
	if err := proto.Unmarshal(responseData, r); err != nil {
		return err
	}
	return nil
}

func appendPrefix(message string) string {
	pc, file, line, ok := runtime.Caller(3)
	if !ok {
		return fmt.Sprintf("[?:0 - ?] %s", message)
	}
	path := strings.Split(file, "/")
	filename := path[len(path)-1]

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return fmt.Sprintf("[%s:%d - ?] %s", filename, line, message)
	}
	fPath := strings.Split(fn.Name(), "/")
	funcName := fPath[len(fPath)-1]
	return fmt.Sprintf("[%s:%d - %s] %s", filename, line, funcName, message)
}

func saveErrorCall(client string, status int, id uint64, day int, start time.Time, url string) {
	EmitSignal(CALL_WRITE_CHANNEL, &WriteCallSignal{
		Id:       id,
		Source:   url,
		CallerId: client,
		Status:   uint32(status),
		Time:     uint64(start.UnixMicro()),
		Day:      uint32(day),
	})
}

func saveCall(client string, id uint64, day int, start time.Time, duration int64, url string, request *Request, response *Response) {
	EmitSignal(CALL_WRITE_CHANNEL, &WriteCallSignal{
		Id:        id,
		Source:    url,
		CallerId:  client,
		SessionId: response.SessionID,
		Duration:  uint32(duration),
		Request:   request.Body,
		Response:  response.Body,
		Status:    uint32(200),
		Time:      uint64(start.UnixMicro()),
		Day:       uint32(day),
	})
}

func (s *Service) Auth(org string, secret string, apps []string, userID string) (bool, string) {
	request := CreateAuthRequest{
		Organization: org,
		Secret:       secret,
		Apps:         apps,
		UserID:       userID,
	}

	var response CreateAuthResponse
	if err := CallService(AUTH_CREATE_URL, &request, &response); err != OK {
		return false, ""
	}
	s.Response.Token = response.Token
	s.Response.Expired = response.Expired
	return true, response.Token
}
