package quicmoq

import (
	"context"

	moqtransport "github.com/danielpfeifer02/priority-moqtransport"
	"github.com/danielpfeifer02/quic-go-prio-packs"
	"github.com/danielpfeifer02/quic-go-prio-packs/priority_setting"
)

type connection struct {
	connection quic.Connection
}

func New(conn quic.Connection) moqtransport.Connection {
	return &connection{conn}
}

func (c *connection) OpenStream() (moqtransport.Stream, error) {
	return c.connection.OpenStream()
}

func (c *connection) OpenStreamWithPriority(p priority_setting.Priority) (moqtransport.SendStream, error) {
	return c.connection.OpenStreamWithPriority(p)
}

func (c *connection) OpenStreamSync(ctx context.Context) (moqtransport.Stream, error) {
	return c.connection.OpenStreamSync(ctx)
}

func (c *connection) OpenStreamSyncWithPriority(ctx context.Context, p priority_setting.Priority) (moqtransport.SendStream, error) {
	return c.connection.OpenStreamSyncWithPriority(ctx, p)
}

func (c *connection) OpenUniStream() (moqtransport.SendStream, error) {
	return c.connection.OpenUniStream()
}

func (c *connection) OpenUniStreamWithPriority(p priority_setting.Priority) (moqtransport.SendStream, error) {
	return c.connection.OpenUniStreamWithPriority(p)
}

func (c *connection) OpenUniStreamSync(ctx context.Context) (moqtransport.SendStream, error) {
	return c.connection.OpenUniStreamSync(ctx)
}

func (c *connection) OpenUniStreamSyncWithPriority(ctx context.Context, p priority_setting.Priority) (moqtransport.SendStream, error) {
	return c.connection.OpenUniStreamSyncWithPriority(ctx, p)
}

func (c *connection) AcceptStream(ctx context.Context) (moqtransport.Stream, error) {
	return c.connection.AcceptStream(ctx)
}

func (c *connection) AcceptUniStream(ctx context.Context) (moqtransport.ReceiveStream, error) {
	return c.connection.AcceptUniStream(ctx)
}

func (c *connection) SendDatagram(b []byte) error {
	return c.connection.SendDatagram(b)
}

func (c *connection) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	return c.connection.ReceiveDatagram(ctx)
}

func (c *connection) CloseWithError(e uint64, msg string) error {
	return c.connection.CloseWithError(quic.ApplicationErrorCode(e), msg)
}

// RTT_STATS_TAG
func (c *connection) GetRTTStats() quic.RTTStatistics {
	return c.connection.GetRTTStats()
}
