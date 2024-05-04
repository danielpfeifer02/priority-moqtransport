package moqtransport

import (
	"context"
	"io"

	"github.com/danielpfeifer02/quic-go-prio-packs/priority_setting"
)

type Stream interface {
	ReceiveStream
	SendStream
}

type ReceiveStream interface {
	io.Reader
}

type SendStream interface {
	io.WriteCloser
}

type Connection interface {
	OpenStream() (Stream, error)
	OpenStreamWithPriority(priority_setting.Priority) (SendStream, error)
	OpenStreamSync(context.Context) (Stream, error)
	OpenStreamSyncWithPriority(context.Context, priority_setting.Priority) (SendStream, error)
	OpenUniStream() (SendStream, error)
	OpenUniStreamWithPriority(priority_setting.Priority) (SendStream, error)
	OpenUniStreamSync(context.Context) (SendStream, error)
	OpenUniStreamSyncWithPriority(context.Context, priority_setting.Priority) (SendStream, error)
	AcceptStream(context.Context) (Stream, error)
	AcceptUniStream(context.Context) (ReceiveStream, error)
	SendDatagram([]byte) error
	ReceiveDatagram(context.Context) ([]byte, error)
	CloseWithError(uint64, string) error
}
