package queues

import "github.com/nsqio/go-nsq"

const (
	NSQ_TOPIC = "agscheduler_topic"
)

type NsqMessageHandler struct {
	jobC chan []byte
}

func (h *NsqMessageHandler) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		return nil
	}
	h.jobC <- m.Body

	return nil
}

// Queue jobs in NSQ.
type NsqQueue struct {
	Producer *nsq.Producer
	Consumer *nsq.Consumer
	Mh       *NsqMessageHandler
	Topic    string

	jobC chan []byte
}

func (q *NsqQueue) Init() error {
	if q.Topic == "" {
		q.Topic = NSQ_TOPIC
	}

	q.jobC = make(chan []byte, 5)
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

	q.Producer.Stop()
	q.Consumer.Stop()

	return nil
}
