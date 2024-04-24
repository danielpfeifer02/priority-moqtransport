package moqtransport

import (
	"errors"
	"fmt"
)

var (
	errInvalidSendMode = errors.New("invalid send mode")
)

type sendMode int

const (
	streamPerObject sendMode = iota
	streamPerGroup
	singleStream
	datagram
)

type SendTrack struct {
	conn   connection
	mode   sendMode
	id     uint64
	stream sendStream
}

func newSendTrack(conn connection) *SendTrack {
	s, err := conn.OpenUniStream()
	if err != nil {
		// TODO
		panic(err)
	}
	return &SendTrack{
		conn:   conn,
		mode:   datagram,
		id:     0,
		stream: s,
	}
}

func (t *SendTrack) writeNewStream(b []byte) (int, error) {
	s, err := t.conn.OpenUniStream()
	if err != nil {
		return 0, err
	}
	om := &objectMessage{
		trackID:         t.id,
		groupSequence:   0,
		objectSequence:  0,
		objectSendOrder: 0,
		objectPayload:   b,
	}
	buf := make([]byte, 0, 64_000)
	buf = om.append(buf)
	defer s.Close()
	return s.Write(buf)
}

// DATAGRAM_TAG
func (t *SendTrack) Write(b []byte) (n int, err error) {
	switch t.mode {
	case streamPerObject:
		return t.writeNewStream(b)
	case streamPerGroup:
		// ...
	case singleStream:
	case datagram:
		om := &objectMessage{
			trackID:         t.id,
			groupSequence:   0,
			objectSequence:  0,
			objectSendOrder: 0,
			objectPayload:   b,
		}
		buf := make([]byte, 0, 64_000)
		buf = om.append(buf)

		if t.mode == singleStream {
			return t.stream.Write(buf)
		}

		fmt.Println("DG length:", len(buf))
		err = t.sendDatagramSplit(buf)
		return len(b), err
	}
	return 0, errInvalidSendMode
}

func (t *SendTrack) sendDatagramSplit(b []byte) error {
	// max_dg_len := 16380 // TODO: why not working even though this is max size?
	n := 1_200 // TODO: limit due to MTU?
	for len(b) > 0 {
		if len(b) < n {
			n = len(b)
		}
		err := t.conn.SendDatagram(b[:n])
		if err != nil {
			return err
		}
		b = b[n:]
	}
	return nil
}
