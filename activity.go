package scyna

import (
	"strings"
	"sync"
	"time"

	"github.com/scylladb/gocqlx/v2/qb"
	"google.golang.org/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)

const tryCount = 10

type ActivityStream struct {
	Entity  string
	Stream  string
	Queries *QueryPool
}

type Activity struct {
	EntityID uint64 `db:"entity_id"`
	Type     int32  `db:"type"`
	Time     int64  `db:"time"`
	Data     []byte `db:"data"`
}

func InitActivityStream(entity string, stream string) *ActivityStream {
	tName := strings.Split(entity, ".")[0] + ".es_" + strings.Split(entity, ".")[1] + "_" + stream
	/*TODO: check if table tName existed, call fatal to exit*/
	return &ActivityStream{
		Entity: entity,
		Stream: stream,
		Queries: &QueryPool{
			sync.Pool{
				New: func() interface{} {
					return qb.Insert(tName).Columns("entity_id", "type", "time", "data").Unique().Query(DB)
				},
			},
		},
	}
}

func (stream *ActivityStream) Add(entity uint64, Type int, event protoreflect.ProtoMessage) {
	t := uint64(time.Now().UnixMicro())

	var data []byte
	if event != nil {
		data, _ = proto.Marshal(event)
	}

	qInsert := stream.Queries.GetQuery()
	defer stream.Queries.Put(qInsert)

	for i := 0; i < tryCount; i++ {
		qInsert.Bind(entity, Type, t, data)
		if applied, err := qInsert.ExecCAS(); applied {
			return
		} else {
			if err != nil {
				LOG.Error("ActivityStream.Add :" + err.Error())
				return
			}
		}
		t++
	}
}
