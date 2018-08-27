// Package nsq provides a Vice implementation for NSQ.
package common

import (
	"sync"
	"time"
	"fmt"
	"github.com/nsqio/go-nsq"
)

// DefaultTCPAddr is the default NSQ TCP address
const DefaultTCPAddr = "123.206.232.202:4150"

// Err represents a vice error.
// Err 表示vice错误
type Err struct {
	Message []byte
	Name    string
	Err     error
}

func (e Err) Error() string {
	if len(e.Message) > 0 {
		return fmt.Sprintf("%s: |%s| <- `%s`", e.Err, e.Name, string(e.Message))
	}
	return fmt.Sprintf("%s: |%s|", e.Err, e.Name)
}

// Transport is a vice.Transport for NSQ.
type Transport struct {
	// sendmessage 互斥锁
	sm        sync.Mutex
	sendChans map[string]chan []byte

	// recivemessage 互斥锁
	rm           sync.Mutex
	receiveChans map[string]chan []byte

	errChan chan error

	// stopchan is closed when everything has stopped.
	stopchan chan struct{}
	// stopProdChan is closed when producers should stop.
	stopProdChan chan struct{}

	consumers []*nsq.Consumer

	// producersWG tracks running producers
	producersWG sync.WaitGroup

	// NewProducer is a func that creates an nsq.Producer.
	NewProducer func() (*nsq.Producer, error)

	// NewConsumer is a func that creates an nsq.Consumer.
	NewConsumer func(name string) (*nsq.Consumer, error)

	// ConnectConsumer is a func that connects the nsq.Consumer
	// to NSQ.
	ConnectConsumer func(consumer *nsq.Consumer) error
}

// New makes a new Transport.
// 新建一个Transport
func New() *Transport {

	// 新建一个Transport
	// sendChans是  字符串  []byte类型chan的映射
	// receiveChans是 字符串 []byte了下chan的映射

	// stopchan是一个chan结构体
	// stopProchan是一个chan结构体
	// errchan是一个error chan

	// consumer是nsq的消费者
	//
	return &Transport{
		sendChans:    make(map[string]chan []byte),
		receiveChans: make(map[string]chan []byte),

		stopchan:     make(chan struct{}),
		stopProdChan: make(chan struct{}),
		errChan:      make(chan error, 10),

		consumers: []*nsq.Consumer{},

		NewProducer: func() (*nsq.Producer, error) {
			return nsq.NewProducer(DefaultTCPAddr, nsq.NewConfig())
		},
		NewConsumer: func(name string) (*nsq.Consumer, error) {
			return nsq.NewConsumer(name, "vice", nsq.NewConfig())
		},
		ConnectConsumer: func(consumer *nsq.Consumer) error {
			return consumer.ConnectToNSQD(DefaultTCPAddr)
		},
	}
}

// ErrChan gets the channel on which errors are sent.
func (t *Transport) ErrChan() <-chan error {
	return t.errChan
}

// Receive gets a channel on which to receive messages
// with the specified name.
func (t *Transport) Receive(name string) <-chan []byte {
	t.rm.Lock()
	defer t.rm.Unlock()

	ch, ok := t.receiveChans[name]
	if ok {
		return ch
	}
	var err error
	if ch, err = t.makeConsumer(name); err != nil {
		// failed to make a consumer, so send an error down the
		// ReceiveErrs channel and return an empty channel to
		// avoid panic.
		t.errChan <- Err{Name: name, Err: err}
		return make(chan []byte)
	}
	t.receiveChans[name] = ch
	return ch
}

func (t *Transport) makeConsumer(name string) (chan []byte, error) {
	ch := make(chan []byte)
	consumer, err := t.NewConsumer(name)
	if err != nil {
		return nil, err
	}
	consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		body := message.Body
		message.Finish() // sends the ACK to avoid long blocking
		ch <- body
		return nil
	}))

	err = Do(1*time.Second, 10*time.Minute, 0, func() error {
		return t.ConnectConsumer(consumer)
	})
	if err != nil {
		return nil, err
	}
	t.consumers = append(t.consumers, consumer)
	return ch, nil
}

// Send gets a channel on which messages with the
// specified name may be sent.
func (t *Transport) Send(name string) chan<- []byte {
	t.sm.Lock()
	defer t.sm.Unlock()

	ch, ok := t.sendChans[name]
	if ok {
		return ch
	}
	var err error
	ch, err = t.makeProducer(name)
	if err != nil {
		// failed to make a producer, send an error down the
		// sendErrsChan and return an empty channel so we don't
		// panic.
		t.errChan <- Err{Name: name, Err: err}
		return make(chan []byte)
	}
	t.sendChans[name] = ch
	return ch
}

func (t *Transport) makeProducer(name string) (chan []byte, error) {
	ch := make(chan []byte)
	producer, err := t.NewProducer()
	if err != nil {
		return nil, err
	}
	t.producersWG.Add(1)
	go func() {
		defer func() {
			producer.Stop()
			t.producersWG.Done()
		}()
		for {
			select {
			case <-t.stopProdChan:
				return
			case msg := <-ch:
				err = Do(1*time.Second, 10*time.Minute, 10, func() error {
					return producer.Publish(name, msg)
				})
				if err != nil {
					t.errChan <- Err{Message: msg, Name: name, Err: err}
					continue
				}
			}
		}
	}()
	return ch, nil
}

// Stop stops the transport.
// The channel returned from Done() will be closed
// when the transport has stopped.
func (t *Transport) Stop() {
	// stops and waits for the producers
	close(t.stopProdChan)
	t.producersWG.Wait()

	for _, c := range t.consumers {
		c.Stop()
		<-c.StopChan
	}
	close(t.stopchan)
}

// Done gets a channel which is closed when the
// transport has successfully stopped.
func (t *Transport) Done() chan struct{} {
	return t.stopchan
}