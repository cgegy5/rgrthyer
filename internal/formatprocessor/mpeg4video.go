package formatprocessor

import (
	"fmt"
	"time"

	"github.com/bluenviron/gortsplib/v3/pkg/formats"
	"github.com/bluenviron/gortsplib/v3/pkg/formats/rtpmpeg4video"
	"github.com/pion/rtp"
)

// UnitMPEG4Video is a MPEG-4 Video data unit.
type UnitMPEG4Video struct {
	RTPPackets []*rtp.Packet
	NTP        time.Time
	PTS        time.Duration
	Frame      []byte
}

// GetRTPPackets implements Unit.
func (d *UnitMPEG4Video) GetRTPPackets() []*rtp.Packet {
	return d.RTPPackets
}

// GetNTP implements Unit.
func (d *UnitMPEG4Video) GetNTP() time.Time {
	return d.NTP
}

type formatProcessorMPEG4Video struct {
	udpMaxPayloadSize int
	format            *formats.MPEG4Video
	encoder           *rtpmpeg4video.Encoder
	decoder           *rtpmpeg4video.Decoder
}

func newMPEG4Video(
	udpMaxPayloadSize int,
	forma *formats.MPEG4Video,
	allocateEncoder bool,
) (*formatProcessorMPEG4Video, error) {
	t := &formatProcessorMPEG4Video{
		udpMaxPayloadSize: udpMaxPayloadSize,
		format:            forma,
	}

	if allocateEncoder {
		t.encoder = &rtpmpeg4video.Encoder{
			PayloadMaxSize: t.udpMaxPayloadSize - 12,
		}
		t.encoder.Init()
	}

	return t, nil
}

func (t *formatProcessorMPEG4Video) Process(unit Unit, hasNonRTSPReaders bool) error { //nolint:dupl
	tunit := unit.(*UnitMPEG4Video)

	if tunit.RTPPackets != nil {
		pkt := tunit.RTPPackets[0]

		// remove padding
		pkt.Header.Padding = false
		pkt.PaddingSize = 0

		if pkt.MarshalSize() > t.udpMaxPayloadSize {
			return fmt.Errorf("payload size (%d) is greater than maximum allowed (%d)",
				pkt.MarshalSize(), t.udpMaxPayloadSize)
		}

		// decode from RTP
		if hasNonRTSPReaders || true {
			if t.decoder == nil {
				t.decoder = t.format.CreateDecoder()
			}

			frame, pts, err := t.decoder.Decode(pkt)
			if err != nil {
				if err == rtpmpeg4video.ErrMorePacketsNeeded {
					return nil
				}
				return err
			}

			// fmt.Println("VOP", len(vop), pts)

			tunit.Frame = frame
			tunit.PTS = pts
		}

		// route packet as is
		return nil
	}

	// encode into RTP
	pkts, err := t.encoder.Encode(tunit.Frame, tunit.PTS)
	if err != nil {
		return err
	}
	tunit.RTPPackets = pkts

	return nil
}
