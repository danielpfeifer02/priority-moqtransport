package moqtransport

import (
	"fmt"

	"github.com/danielpfeifer02/quic-go-prio-packs/packet_setting"
)

type objectStream struct {
	streams streamCollection
}

type streamCollection struct {
	highPriorityStream SendStream
	lowPriorityStream  SendStream
	noPriorityStream   SendStream
}

func newObjectStream(streams streamCollection, subscribeID, trackAlias, groupID, objectID, objectSendOrder uint64) (*objectStream, error) {
	osm := &objectMessage{
		datagram:        false,
		SubscribeID:     subscribeID,
		TrackAlias:      trackAlias,
		GroupID:         groupID,
		ObjectID:        objectID,
		ObjectSendOrder: objectSendOrder,
		ObjectPayload:   nil,
	}
	// PRIORITY_TAG
	sl := []SendStream{
		streams.highPriorityStream,
		streams.lowPriorityStream,
		streams.noPriorityStream,
	}
	for _, stream := range sl {
		buf := make([]byte, 0, 48)
		buf = osm.append(buf)
		_, err := stream.Write(buf)
		if err != nil {
			return nil, err
		}
	}
	return &objectStream{
		streams: streams,
	}, nil
}

// In case one wants to showcase the issue of error propagation
// when dropping packets, uncomment the following lines and
// set minimum dist between I-frames to 20 and max dist to 30+.
// var last_was_i = 0

func (s *objectStream) Write(payload []byte) (int, error) {

	// Before the vp8 payload, there is an 8 byte timestamp
	// saved as metadata.
	// This is a little hacky, but it works for now.
	vp8_offset := 8
	hdr := payload[vp8_offset]

	if false {
		size0 := (hdr >> 5) & 0x07
		ver := (hdr >> 1) & 0x07
		fmt.Println("size0: ", size0, " ver: ", ver, "len: ", len(payload))
	}

	if (hdr&0x01) == 1 && ((hdr>>4)&0x01) == 1 {
		packet_setting.DebugPrintln("LOW PRIORITY STREAM")
		// if last_was_i > 0 {
		// 	last_was_i -= 1
		// 	return len(payload), nil
		// }
		return s.streams.lowPriorityStream.Write(payload)
	} else {
		packet_setting.DebugPrintln("HIGH PRIORITY STREAM")
		// last_was_i = 20
		return s.streams.highPriorityStream.Write(payload)
	}

}

func (s *objectStream) Close() error {
	sl := []SendStream{
		s.streams.highPriorityStream,
		s.streams.lowPriorityStream,
		s.streams.noPriorityStream,
	}
	for _, stream := range sl {
		if err := stream.Close(); err != nil { // TODO: if err occurs, later streams will not be closed
			return err
		}
	}
	return nil
}
