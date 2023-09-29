package transport

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"time"

	"gitlab.lrz.de/cm/moqtransport/varint"
	"golang.org/x/exp/slices"
)

var (
	errUnexpectedMessage = errors.New("got unexpected message")
	errClosed            = errors.New("connection was closed")
)

// TODO: Streams must be wrapped properly for quic and webtransport The
// interfaces need to add CancelRead and CancelWrite for STOP_SENDING and
// RESET_STREAM purposes. The interface should allow implementations for quic
// and webtransport.
type Stream interface {
	ReadStream
	SendStream
}

type ReadStream interface {
	io.Reader
}

type SendStream interface {
	io.WriteCloser
}

type Connection interface {
	OpenStream() (Stream, error)
	OpenStreamSync(context.Context) (Stream, error)
	OpenUniStream() (SendStream, error)
	OpenUniStreamSync(context.Context) (SendStream, error)
	AcceptStream(context.Context) (Stream, error)
	AcceptUniStream(context.Context) (ReadStream, error)
	ReceiveMessage(context.Context) ([]byte, error)
}

type Peer struct {
	conn                Connection
	inMsgCh             chan Message
	ctrlMessageCh       chan keyedResponseHandler
	role                role
	receiveTracks       map[uint64]*ReceiveTrack
	sendTracks          map[string]*SendTrack
	subscribeHandler    func(*SendTrack)
	announcementHandler func()
	closeCh             chan struct{}
}

func NewServerPeer(ctx context.Context, conn Connection) (*Peer, error) {
	stream, err := conn.AcceptStream(ctx)
	if err != nil {
		return nil, err
	}
	m, err := ReadNext(varint.NewReader(stream), serverRole)
	if err != nil {
		return nil, err
	}
	msg, ok := m.(*ClientSetupMessage)
	if !ok {
		return nil, errUnexpectedMessage
	}
	// TODO: Algorithm to select best matching version
	if !slices.Contains(msg.SupportedVersions, DRAFT_IETF_MOQ_TRANSPORT_00) {
		// TODO: Close conn with error
		log.Println("TODO: Close conn with error")
		return nil, nil
	}
	_, ok = msg.SetupParameters[roleParameterKey]
	if !ok {
		// ERROR: role is required
		log.Println("TODO: ERROR: role is required")
		return nil, nil
	}
	// TODO: save role parameter
	ssm := ServerSetupMessage{
		SelectedVersion: DRAFT_IETF_MOQ_TRANSPORT_00,
		SetupParameters: map[parameterKey]Parameter{},
	}
	buf := ssm.Append(make([]byte, 0, 1500))
	_, err = stream.Write(buf)
	if err != nil {
		return nil, err
	}
	p := &Peer{
		conn:          conn,
		inMsgCh:       make(chan Message),
		ctrlMessageCh: make(chan keyedResponseHandler),
		role:          serverRole,
	}
	go p.controlStreamLoop(ctx, stream)
	go p.acceptBidirectionalStreams(ctx)
	go p.acceptUnidirectionalStreams(ctx)
	// TODO: Configure if datagrams enabled?
	go p.acceptDatagrams(ctx)
	return p, nil
}

func NewClientPeer(ctx context.Context, conn Connection) (*Peer, error) {
	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		return nil, err
	}
	csm := ClientSetupMessage{
		SupportedVersions: []Version{Version(DRAFT_IETF_MOQ_TRANSPORT_00)},
		SetupParameters: map[parameterKey]Parameter{
			roleParameterKey: ingestionDeliveryRole,
		},
	}
	buf := csm.Append(make([]byte, 0, 1500))
	_, err = stream.Write(buf)
	if err != nil {
		return nil, err
	}
	p := &Peer{
		conn:          conn,
		inMsgCh:       make(chan Message),
		ctrlMessageCh: make(chan keyedResponseHandler),
		role:          clientRole,
	}
	m, err := ReadNext(varint.NewReader(stream), clientRole)
	if err != nil {
		return nil, err
	}
	_, ok := m.(*ServerSetupMessage)
	if !ok {
		return nil, errUnexpectedMessage
	}

	go p.controlStreamLoop(ctx, stream)
	go p.acceptBidirectionalStreams(ctx)
	go p.acceptUnidirectionalStreams(ctx)
	// TODO: Configure if datagrams enabled?
	go p.acceptDatagrams(ctx)
	return p, nil
}

func (p *Peer) readMessages(r MessageReader, stream io.Reader) {
	for {
		log.Println("reading message")
		msg, err := ReadNext(r, p.role)
		if err != nil {
			// TODO: Handle/log error?
			log.Println(err)
			return
		}
		log.Printf("read message: %v\n", msg)
		object, ok := msg.(*ObjectMessage)
		if !ok {
			// TODO: ERROR: We only expect object messages here. All other
			// messages should be sent on the control stream.
		}
		p.handleObjectMessage(object)
	}
}

type messageKey struct {
	mt MessageType
	id string
}

type keyer interface {
	key() messageKey
}

type keyedMessage interface {
	Message
	keyer
}

type responseHandler interface {
	handle(Message)
}

type keyedResponseHandler interface {
	keyedMessage
	responseHandler
}

func (p *Peer) controlStreamLoop(ctx context.Context, s Stream) {
	inCh := make(chan Message)
	errCh := make(chan error)
	transactions := make(map[messageKey]keyedMessage)

	go func(s Stream, ch chan<- Message, errCh chan<- error) {
		for {
			msg, err := ReadNext(varint.NewReader(s), p.role)
			if err != nil {
				errCh <- err
				return
			}
			ch <- msg
		}
	}(s, inCh, errCh)
	// control stream interactions are tricky.
	// We need to:
	// 1. Make sure we always accept any control message (the peer might want to
	// start a transaction)
	// 2. Send out messages passed to us on the outMsgCh
	// 3. Wait for responses to messages we sent and match them to the correct
	// oneoutMsgCh
	// 4. (Maybe optional?) Be able to send out multiple messages without doing
	// lock-steps?
	// How can we map a message to the corresponding transaction?
	for {
		select {
		case m := <-inCh:
			switch v := m.(type) {
			case *SubscribeRequestMessage, *AnnounceMessage, *GoAwayMessage:
				// TODO: Handle
			case keyedMessage:
				t, ok := transactions[v.key()]
				if !ok {
					// TODO: Error: This an error, because all keyed messages
					// that occur without responding to a transaction started by
					// us should be handled by the case above. I.e., if we get
					// an Ok or Error for Announce or subscribe, that should
					// only happen when we also stored an associated transaction
					// earlier.
					panic("TODO")
				}
				rh, ok := t.(responseHandler)
				if !ok {
					// TODO: Error: This is also an error, because we shouldn't
					// have started a transaction if we cannot handle the
					// response
					panic("TODO")
				}
				rh.handle(t)
			default:
				// error unexpected message, close conn?
				panic("TODO")
			}
		case m := <-p.ctrlMessageCh:
			transactions[m.key()] = m
			buf := make([]byte, 0, 1500)
			buf = m.Append(buf)
			_, err := s.Write(buf)
			if err != nil {
				// TODO
				log.Println(err)
			}
		case err := <-errCh:
			// TODO
			log.Println(err)
			panic(err)
		}
	}
}

func (p *Peer) acceptBidirectionalStreams(ctx context.Context) {
	for {
		stream, err := p.conn.AcceptStream(ctx)
		if err != nil {
			// TODO: Handle/log error?
			log.Println(err)
			return
		}
		log.Println("got new bidirectional stream")
		go p.readMessages(varint.NewReader(stream), stream)
	}
}

func (p *Peer) acceptUnidirectionalStreams(ctx context.Context) {
	for {
		stream, err := p.conn.AcceptUniStream(ctx)
		if err != nil {
			// TODO: Handle/log error?
			log.Println(err)
			return
		}
		log.Println("got new unidirectional stream")
		go p.readMessages(varint.NewReader(stream), stream)
	}
}

func (p *Peer) acceptDatagrams(ctx context.Context) {
	for {
		dgram, err := p.conn.ReceiveMessage(ctx)
		if err != nil {
			// TODO: Handle/log error?
			log.Println(err)
			return
		}
		r := bytes.NewReader(dgram)
		go p.readMessages(r, nil)
	}
}

func (p *Peer) handleObjectMessage(msg *ObjectMessage) error {
	t, ok := p.receiveTracks[msg.TrackID]
	if !ok {
		// handle unknown track?
		panic("TODO")
	}
	t.push(msg)
	return nil
}

func (p *Peer) handleSubscribeRequest(msg SubscribeRequestMessage) error {
	if p.subscribeHandler == nil {
		panic("TODO")
	}
	t := newSendTrack(p.conn)
	p.sendTracks[msg.FullTrackName] = t
	p.subscribeHandler(t)
	return nil
}

func (p *Peer) handleAnnounceMessage(msg AnnounceMessage) error {
	if p.announcementHandler == nil {
		panic("TODO")
	}
	p.announcementHandler()
	return nil
}

type ctrlMessage struct {
	keyedMessage
	responseCh chan Message
}

func (m *ctrlMessage) handle(msg Message) {
	m.responseCh <- msg
}

func (p *Peer) Announce(namespace string) error {
	am := &AnnounceMessage{
		TrackNamespace:  namespace,
		TrackParameters: map[parameterKey]Parameter{},
	}
	responseCh := make(chan Message)
	select {
	case p.ctrlMessageCh <- &ctrlMessage{
		keyedMessage: am,
		responseCh:   responseCh,
	}:
	case <-p.closeCh:
		return errClosed
	}
	var resp Message
	select {
	case resp = <-responseCh:
	case <-time.After(time.Second): // TODO: Make timeout configurable?
		panic("TODO: timeout error")
	case <-p.closeCh:
		return errClosed
	}
	switch v := resp.(type) {
	case *AnnounceOkMessage:
		if v.TrackNamespace != am.TrackNamespace {
			panic("TODO")
		}
	case *AnnounceErrorMessage:
		return errors.New(v.ReasonPhrase) // TODO: Wrap error string?
	default:
		return errUnexpectedMessage
	}
	return nil
}

func (p *Peer) Subscribe(namespace string) (*ReceiveTrack, error) {
	sm := &SubscribeRequestMessage{
		FullTrackName:          namespace,
		TrackRequestParameters: map[parameterKey]Parameter{},
	}
	responseCh := make(chan Message)
	select {
	case p.ctrlMessageCh <- &ctrlMessage{
		keyedMessage: sm,
		responseCh:   responseCh,
	}:
	case <-p.closeCh:
		return nil, errClosed
	case <-time.After(time.Second):
		panic("TODO: timeout error")
	}
	var resp Message
	select {
	case resp = <-responseCh:
	case <-time.After(time.Second): // TODO: Make timeout configurable?
		panic("TODO: timeout error")
	case <-p.closeCh:
		return nil, errClosed
	}
	switch v := resp.(type) {
	case *SubscribeOkMessage:
		if v.FullTrackName != sm.FullTrackName {
			panic("TODO")
		}
		t := newReceiveTrack()
		p.receiveTracks[v.TrackID] = t
		return t, nil

	case *SubscribeErrorMessage:
		return nil, errors.New(v.ReasonPhrase)
	}
	return nil, errUnexpectedMessage
}

func (p *Peer) OnAnnouncement(callback func()) {
	p.announcementHandler = callback
}
