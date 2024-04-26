package chat

import (
	"context"
	"crypto/tls"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/danielpfeifer02/priority-moqtransport"
	"github.com/danielpfeifer02/priority-moqtransport/quicmoq"
	"github.com/danielpfeifer02/quic-go-prio-packs"
)

type Server struct {
	chatRooms   map[string]*room
	peers       map[*moqtransport.Session]string
	nextTrackID uint64
	lock        sync.Mutex
}

func NewServer() *Server {
	s := &Server{
		chatRooms:   map[string]*room{},
		peers:       map[*moqtransport.Session]string{},
		nextTrackID: 1,
		lock:        sync.Mutex{},
	}
	return s
}

func (s *Server) ListenQUIC(ctx context.Context, addr string, tlsConfig *tls.Config) error {
	listener, err := quic.ListenAddr(addr, tlsConfig, &quic.Config{
		EnableDatagrams: true,
	})
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept(ctx)
		if err != nil {
			return err
		}
		session, err := moqtransport.NewServerSession(quicmoq.New(conn), true)
		if err != nil {
			return err
		}
		go s.handle(session)
	}
}

func (s *Server) handle(p *moqtransport.Session) {
	var name string

	go func() {
		for {
			a, err := p.ReadAnnouncement(context.Background())
			if err != nil {
				panic(err)
			}
			uri := strings.SplitN(a.Namespace(), "/", 4)
			if len(uri) < 4 {
				a.Reject(0, "invalid announcement")
				continue
			}
			moq_chat, id, participant, username := uri[0], uri[1], uri[2], uri[3]
			if moq_chat != "moq-chat" || participant != "participant" {
				a.Reject(0, "invalid moq-chat namespace")
				continue
			}
			a.Accept()
			name = username
			if _, ok := s.chatRooms[id]; !ok {
				s.chatRooms[id] = newChat(id)
			}
			if err := s.chatRooms[id].join(name, p); err != nil {
				log.Println(err)
			}
			log.Printf("announcement accepted: %v", a.Namespace())
		}
	}()
	go func() {
		for {
			sub, err := p.ReadSubscription(context.Background())
			if err != nil {
				panic(err)
			}
			if len(name) == 0 {
				// Subscribe requires a username which has to be announced
				// before subscribing
				sub.Reject(moqtransport.SubscribeErrorUnknownTrack, "subscribe without prior announcement")
				continue
			}
			parts := strings.SplitN(sub.Namespace(), "/", 4)
			if len(parts) < 2 {
				sub.Reject(moqtransport.SubscribeErrorUnknownTrack, "invalid trackname")
				continue
			}
			moq_chat, id := parts[0], parts[1]
			if moq_chat != "moq-chat" {
				sub.Reject(0, "invalid moq-chat namespace")
				continue
			}
			r, ok := s.chatRooms[id]
			if !ok {
				sub.Reject(moqtransport.SubscribeErrorUnknownTrack, "unknown chat id")
				continue
			}
			if sub.Trackname() == "/catalog" {
				sub.Accept()
				go func() {
					// TODO: Improve synchronization (buffer objects before
					// subscription finished)
					time.Sleep(100 * time.Millisecond)
					s.chatRooms[id].subscribe(name, sub)
				}()
				continue
			}

			r.lock.Lock()
			log.Printf("subscribing user %v to publisher %v", name, sub.Trackname())
			if len(parts) < 4 {
				sub.Reject(0, "invalid subscriptions namespace, expected 'moq-chat/<room-id>/participants/<username>")
				continue
			}
			username := parts[3]
			r.publishers[username].subscribe(name, sub)
			r.lock.Unlock()
		}
	}()
}

type subscriber struct {
	name  string
	track *moqtransport.SendSubscription
}
