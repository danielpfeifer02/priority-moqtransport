package moqtransport

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/danielpfeifer02/quic-go-prio-packs"
	"github.com/danielpfeifer02/quic-go-prio-packs/priority_setting"
)

var errUnsubscribed = errors.New("peer unsubscribed")

type subscribeError struct {
	code   uint64
	reason string
}

type SendSubscription struct {
	lock       sync.RWMutex
	responseCh chan *subscribeError
	closeCh    chan struct{}
	expires    time.Duration

	conn Connection

	subscribeID, trackAlias uint64
	namespace, trackname    string
	startGroup, startObject Location
	endGroup, endObject     Location
	parameters              parameters
}

func (s *SendSubscription) Accept() {
	select {
	case <-s.closeCh:
	case s.responseCh <- nil:
	}
}

func (s *SendSubscription) Reject(code uint64, reason string) {
	select {
	case <-s.closeCh:
	case s.responseCh <- &subscribeError{
		code:   code,
		reason: reason,
	}:
	}
}

func (s *SendSubscription) SetExpires(d time.Duration) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.expires = d
}

func (s *SendSubscription) Namespace() string {
	return s.namespace
}

func (s *SendSubscription) Trackname() string {
	return s.trackname
}

func (s *SendSubscription) StartGroup() Location {
	return s.startGroup
}

func (s *SendSubscription) StartObject() Location {
	return s.startObject
}

func (s *SendSubscription) EndGroup() Location {
	return s.endGroup
}

func (s *SendSubscription) EndObject() Location {
	return s.endObject
}

func (s *SendSubscription) unsubscribe() {
	close(s.closeCh)
}

func (s *SendSubscription) NewObjectStream(groupID, objectID, objectSendOrder uint64) (*objectStream, error) {
	select {
	case <-s.closeCh:
		return nil, errUnsubscribed
	default:
	}

	// PRIORITY_TAG
	high_stream, err := s.conn.OpenUniStreamWithPriority(priority_setting.HighPriority)
	if err != nil {
		return nil, err
	}
	low_stream, err := s.conn.OpenUniStreamWithPriority(priority_setting.LowPriority)
	if err != nil {
		return nil, err
	}
	no_stream, err := s.conn.OpenUniStream()
	if err != nil {
		return nil, err
	}
	streams := streamCollection{
		highPriorityStream: high_stream,
		lowPriorityStream:  low_stream,
		noPriorityStream:   no_stream,
	}
	return newObjectStream(streams, s.subscribeID, s.trackAlias, groupID, objectID, objectSendOrder)
}

func (s *SendSubscription) NewObjectPreferDatagram(groupID, objectID, objectSendOrder uint64, payload []byte) error {
	select {
	case <-s.closeCh:
		return errUnsubscribed
	default:
	}
	o := objectMessage{
		datagram:        true,
		SubscribeID:     s.subscribeID,
		TrackAlias:      s.trackAlias,
		GroupID:         groupID,
		ObjectID:        objectID,
		ObjectSendOrder: objectSendOrder,
		ObjectPayload:   payload,
	}
	buf := make([]byte, 0, 48+len(o.ObjectPayload))
	buf = o.append(buf)
	err := s.conn.SendDatagram(buf)
	if err == nil {
		fmt.Println("Sent datagram")
		return nil
	}
	fmt.Println("datagram too large", len(buf))
	if !errors.Is(err, &quic.DatagramTooLargeError{}) {
		return err
	}
	os, err := s.NewObjectStream(groupID, objectID, objectSendOrder)
	if err != nil {
		return err
	}
	_, err = os.Write(buf)
	if err != nil {
		return err
	}
	return os.Close()
}

func (s *SendSubscription) NewTrackHeaderStream(objectSendOrder uint64) (*TrackHeaderStream, error) {
	select {
	case <-s.closeCh:
		return nil, errUnsubscribed
	default:
	}
	stream, err := s.conn.OpenUniStream()
	if err != nil {
		return nil, err
	}
	return newTrackHeaderStream(stream, s.subscribeID, s.trackAlias, objectSendOrder)
}

func (s *SendSubscription) NewGroupHeaderStream(groupID, objectSendOrder uint64) (*groupHeaderStream, error) {
	select {
	case <-s.closeCh:
		return nil, errUnsubscribed
	default:
	}
	stream, err := s.conn.OpenUniStream()
	if err != nil {
		return nil, err
	}
	return newGroupHeaderStream(stream, s.subscribeID, s.trackAlias, groupID, objectSendOrder)
}
