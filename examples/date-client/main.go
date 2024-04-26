package main

import (
	"context"
	"crypto/tls"
	"flag"
	"io"
	"log"

	"github.com/danielpfeifer02/priority-moqtransport"
	"github.com/danielpfeifer02/priority-moqtransport/quicmoq"
	"github.com/danielpfeifer02/quic-go-prio-packs"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "address to connect to")
	namespace := flag.String("namespace", "clock", "Namespace to subscribe to")
	trackname := flag.String("trackname", "second", "Track to subscribe to")
	flag.Parse()

	if err := run(context.Background(), *addr, *namespace, *trackname); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, addr string, namespace, trackname string) error {
	var session *moqtransport.Session
	var conn moqtransport.Connection
	var err error

	conn, err = dialQUIC(ctx, addr)

	if err != nil {
		return err
	}
	session, err = moqtransport.NewClientSession(conn, moqtransport.IngestionDeliveryRole, true)
	if err != nil {
		return err
	}
	defer session.Close()

	for {
		var a *moqtransport.Announcement
		a, err = session.ReadAnnouncement(context.Background())
		if err != nil {
			panic(err)
		}
		log.Println("got Announcement")
		if a.Namespace() == "clock" {
			a.Accept()
			break
		}
	}

	log.Println("subscribing")
	rs, err := session.Subscribe(context.Background(), 0, 0, namespace, trackname, "")
	if err != nil {
		panic(err)
	}
	log.Println("got subscription")
	buf := make([]byte, 64_000)
	for {
		n, err := rs.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Printf("got last object")
				return nil
			}
			panic(err)
		}
		log.Printf("got object: %v\n", string(buf[:n]))
	}
}

func dialQUIC(ctx context.Context, addr string) (moqtransport.Connection, error) {
	conn, err := quic.DialAddr(ctx, addr, &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"moq-00"},
	}, &quic.Config{
		EnableDatagrams: true,
	})
	if err != nil {
		return nil, err
	}
	return quicmoq.New(conn), nil
}
