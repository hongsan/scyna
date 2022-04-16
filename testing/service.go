package scyna_test

import (
	"fmt"
	"scyna"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
)

func CallService(t *testing.T, url string, request proto.Message) *scyna.Response {
	var req scyna.Request
	var res scyna.Response

	if request != nil {
		var err error
		if req.Body, err = proto.Marshal(request); err != nil {
			t.Fatal("Bad request")
		}
	}

	req.CallID = scyna.ID.Next()
	req.JSON = false

	reqBytes, err := proto.Marshal(&req)
	if err != nil {
		t.Fatal("Bad request")
	}

	msg, respErr := scyna.Connection.Request(scyna.PublishURL(url), reqBytes, 10*time.Second)
	if respErr != nil {
		t.Fatal("Service unavailable")
	}

	err = res.ReadFrom(msg.Data)
	if err != nil {
		t.Fatal("Can not parse response")
	}
	return &res
}

func TestService(t *testing.T, url string, request proto.Message, response proto.Message, code int32) {
	res := CallService(t, url, request)
	if res.Code != code {
		t.Fatal("Code not match:", res.Code)
	}

	tmp := proto.Clone(response)
	if err := proto.Unmarshal(res.Body, tmp); err != nil {
		t.Fatal("Can not parse response")
	}
	fmt.Printf("tmp %s", tmp)
	fmt.Printf("respone %s", response)
	if !proto.Equal(tmp, response) {
		t.Fatal("Response not match")
	}
}

func CallServiceCheckCode(t *testing.T, url string, request proto.Message, code int32) {
	res := CallService(t, url, request)
	if res.Code != code {
		t.Fatal("Code not match:", res.Code)
	}
}

func CallServiceParseResponse(t *testing.T, url string, request proto.Message, response proto.Message, code int32) {
	res := CallService(t, url, request)
	if res.Code != code {
		t.Fatal("Code not match")
	}

	if err := proto.Unmarshal(res.Body, response); err != nil {
		t.Fatal("Can not parse response")
	}
}
