package moqtransport

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

var drop = false

func (s *objectStream) Write(payload []byte) (int, error) {

	// fmt.Println("LEN: ", len(payload))

	if true {
		// REMOVENOW

		tmo := min(len(payload), 100)
		// fmt.Println("---------------------------")
		req := payload[0]
		xbit := (req >> 7) & 0x01
		// nbit := (req >> 5) & 0x01
		sbit := (req >> 4) & 0x01
		pid := req & 0x07
		optionals := 0
		if xbit == 1 {
			optionals += 1
			value := payload[1]
			ibit := (value >> 7) & 0x01
			lbit := (value >> 6) & 0x01
			tbit := (value >> 5) & 0x01
			kbit := (value >> 4) & 0x01

			if ibit == 1 {
				value := payload[2]
				mbit := (value >> 7) & 0x01
				optionals += int(mbit)
			}

			optionals += (int(ibit) + int(lbit) + int(tbit|kbit))

		}
		if sbit == 1 && pid == 0 {
			// fmt.Println("x: ", xbit, " n: ", nbit, " s: ", sbit, " pid: ", pid)
			hdr := payload[optionals]
			// fmt.Println("frame: ", hdr&0x01)

			if (hdr & 0x01) == 1 {
				drop = true
				// fmt.Println("DROPPING")
				return s.streams.lowPriorityStream.Write(payload)
			} else {
				drop = false
			}

			for i := 0; i < tmo; i++ {
				// fmt.Printf("%02x ", payload[i])
			}
		} else {
			if drop {
				// fmt.Println("DROPPING")
				return s.streams.lowPriorityStream.Write(payload)
			}
		}
	}

	return s.streams.highPriorityStream.Write(payload)
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
