package moqtransport

import (
	"context"
	"crypto/tls"
	"errors"
	"log"

	"github.com/danielpfeifer02/quic-go-prio-packs"
)

type PeerHandlerFunc func(*Peer)

func (h PeerHandlerFunc) Handle(p *Peer) {
	h(p)
}

type PeerHandler interface {
	Handle(*Peer)
}

type Server struct {
	Handler   PeerHandler
	TLSConfig *tls.Config
}

type listener interface {
	Accept(context.Context) (connection, error)
}

type quicListener struct {
	ql *quic.Listener
}

func (l *quicListener) Accept(ctx context.Context) (connection, error) {
	c, err := l.ql.Accept(ctx)
	if err != nil {
		return nil, err
	}
	qc := &quicConn{
		conn: c,
	}
	return qc, nil
}

func (s *Server) ListenQUIC(ctx context.Context, addr string) error {
	ql, err := quic.ListenAddr(addr, s.TLSConfig, &quic.Config{
		GetConfigForClient:             nil,
		Versions:                       nil,
		HandshakeIdleTimeout:           0,
		MaxIdleTimeout:                 1<<63 - 1,
		TokenStore:                     nil,
		InitialStreamReceiveWindow:     0,
		MaxStreamReceiveWindow:         0,
		InitialConnectionReceiveWindow: 0,
		MaxConnectionReceiveWindow:     0,
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
		return err
	}
	l := &quicListener{
		ql: ql,
	}
	return s.Listen(ctx, l)
}

func (s *Server) Listen(ctx context.Context, l listener) error {
	_, enableDatagrams := l.(*quicListener)
	for {
		conn, err := l.Accept(context.TODO())
		if err != nil {
			return err
		}
		peer, err := newServerPeer(ctx, conn)
		if err != nil {
			log.Printf("failed to create new server peer: %v", err)
			switch {
			case errors.Is(err, errUnsupportedVersion):
				conn.CloseWithError(SessionTerminatedErrorCode, err.Error())
			case errors.Is(err, errMissingRoleParameter):
				conn.CloseWithError(SessionTerminatedErrorCode, err.Error())
			default:
				conn.CloseWithError(GenericErrorCode, "internal server error")
			}
			continue
		}
		// TODO: This should probably be a map keyed by the MoQ-URI the request
		// is targeting
		go func() {
			peer.run(ctx, enableDatagrams)
		}()
		if s.Handler != nil {
			s.Handler.Handle(peer)
		}
	}
}
