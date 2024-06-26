package moqtransport

import (
	"io"

	"github.com/danielpfeifer02/quic-go-prio-packs/quicvarint"
)

type ReceiveSubscription struct {
	responseCh chan subscribeIDer

	session *Session

	writeBuffer *io.PipeWriter
	readBuffer  *io.PipeReader

	subscribeID uint64
}

func newReceiveSubscription(id uint64, s *Session) *ReceiveSubscription {
	r, w := io.Pipe()
	return &ReceiveSubscription{
		responseCh:  make(chan subscribeIDer),
		session:     s,
		writeBuffer: w,
		readBuffer:  r,
		subscribeID: id,
	}
}

type payloader interface {
	payload() []byte
}

func (s *ReceiveSubscription) push(p payloader) (int, error) {
	return s.writeBuffer.Write(p.payload())
}

func (s *ReceiveSubscription) Read(buf []byte) (int, error) {
	return s.readBuffer.Read(buf)
}

func (s *ReceiveSubscription) Unsubscribe() error {
	if err := s.writeBuffer.Close(); err != nil {
		return err
	}
	return s.session.unsubscribe(s.subscribeID)
}

func (s *ReceiveSubscription) readTrackHeaderStream(rs ReceiveStream) {
	parser := newParser(quicvarint.NewReader(rs))
	for {
		msg, err := parser.parseStreamHeaderTrackObject()
		if err != nil {
			if err == io.EOF {
				return
			}
			panic(err)
		}
		if _, err = s.push(msg); err != nil {
			panic(err)
		}
	}
}

func (s *ReceiveSubscription) readGroupHeaderStream(rs ReceiveStream) {
	parser := newParser(quicvarint.NewReader(rs))
	for {
		msg, err := parser.parseStreamHeaderGroupObject()
		if err != nil {
			if err == io.EOF {
				return
			}
			panic(err)
		}
		if _, err = s.push(msg); err != nil {
			panic(err)
		}
	}
}
