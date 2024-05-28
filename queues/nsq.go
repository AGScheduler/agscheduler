package queues

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"runtime/debug"

	"github.com/nsqio/go-nsq"
)

const (
	NSQ_TOPIC = "agscheduler_topic"
)

// Queue jobs in NSQ.
type NsqQueue struct {
	Producer *nsq.Producer
	Consumer *nsq.Consumer
	Mh       *NsqMessageHandler
	Topic    string
	HttpAddr string

	size int
	jobC chan []byte
}

func (q *NsqQueue) Init(ctx context.Context) error {
	if q.Topic == "" {
		q.Topic = NSQ_TOPIC
	}

	q.size = int(math.Abs(float64(q.size)))
	q.jobC = make(chan []byte, q.size)
	q.Mh.jobC = q.jobC

	return nil
}

func (q *NsqQueue) PushJob(bJ []byte) error {
	err := q.Producer.Publish(q.Topic, bJ)
	if err != nil {
		return err
	}

	return nil
}

func (q *NsqQueue) PullJob() <-chan []byte {
	return q.jobC
}

func (q *NsqQueue) Clear() error {
	defer close(q.jobC)

	// Delete NSQ topic should use the nsqd http api or nsqlookupd http api
	// https://github.com/nsqio/go-nsq/issues/335
	u, err := url.JoinPath(q.HttpAddr, "/topic/delete")
	if err != nil {
		return err
	}
	resp, err := http.Post(u+"?topic="+q.Topic, "text/plain", nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to delete topic: `%s`", body)
	}

	return nil
}

type NsqMessageHandler struct {
	jobC chan []byte
}

func (h *NsqMessageHandler) HandleMessage(m *nsq.Message) error {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("NsqQueue handle message error: `%s`", err))
			slog.Debug(string(debug.Stack()))
		}
	}()

	if len(m.Body) == 0 {
		return nil
	}
	h.jobC <- m.Body

	return nil
}
