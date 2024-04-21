package moqtransport

import (
	"context"
	"crypto/tls"
	"errors"

	"github.com/danielpfeifer02/quic-go-prio-packs"
)

func DialQUIC(ctx context.Context, addr string) (*Peer, error) {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"moq-00"},
	}
	conn, err := quic.DialAddr(context.TODO(), addr, tlsConf, &quic.Config{
		GetConfigForClient:             nil,
		Versions:                       nil,
		HandshakeIdleTimeout:           0,
		MaxIdleTimeout:                 1<<63 - 1,
		TokenStore:                     nil,
		InitialStreamReceiveWindow:     1 << 30, // TODONOW: what should this be?
		MaxStreamReceiveWindow:         1 << 30, // TODONOW: what should this be?
		InitialConnectionReceiveWindow: 1 << 30, // TODONOW: what should this be?
		MaxConnectionReceiveWindow:     1 << 30, // TODONOW: what should this be?
		AllowConnectionWindowIncrease:  nil,
		MaxIncomingStreams:             0,
		MaxIncomingUniStreams:          0,
		KeepAlivePeriod:                0,
		DisablePathMTUDiscovery:        false,
		Allow0RTT:                      false,
		EnableDatagrams:                true,
		Tracer:                         nil,
	})
	if err != nil {
		return nil, err
	}
	qc := &quicConn{
		conn: conn,
	}
	p, err := newClientPeer(ctx, qc, true)
	if err != nil {
		if errors.Is(err, errUnsupportedVersion) {
			conn.CloseWithError(SessionTerminatedErrorCode, errUnsupportedVersion.Error())
		}
		conn.CloseWithError(GenericErrorCode, "internal server error")
		return nil, err
	}
	return p, nil
}
