package media

import (
	"fmt"

	"github.com/pion/webrtc/v3"
)

// StreamManager manages media streams
type StreamManager struct {
	tracks map[string]webrtc.RTPSender
}

// TrackInfo contains information about a media track
type TrackInfo struct {
	Type       string // audio or video
	Codec      string
	Bandwidth  uint32
	SampleRate uint32
}

// NewStreamManager creates a new stream manager
func NewStreamManager() *StreamManager {
	return &StreamManager{
		tracks: make(map[string]webrtc.RTPSender),
	}
}

// AddTrack adds a media track to a peer connection
func (sm *StreamManager) AddTrack(pc *webrtc.PeerConnection, trackType, codec string) (webrtc.RTPSender, error) {
	if trackType != "audio" && trackType != "video" {
		return nil, fmt.Errorf("invalid track type: %s", trackType)
	}

	mediaStream := &webrtc.MediaStream{
		ID: "stream-" + codec,
	}

	track := &webrtc.Track{
		PayloadType: 111,
		Kind:        webrtc.RTPCodecTypeVideo,
		ID:          "track-" + codec,
		StreamID:    mediaStream.ID,
	}

	if trackType == "audio" {
		track.Kind = webrtc.RTPCodecTypeAudio
	}

	sender, err := pc.AddTrack(track)
	if err != nil {
		return nil, err
	}

	return sender, nil
}

// RemoveTrack removes a media track
func (sm *StreamManager) RemoveTrack(pc *webrtc.PeerConnection, sender webrtc.RTPSender) error {
	return pc.RemoveTrack(sender)
}

// GetConnectionStats returns statistics for a peer connection
func GetConnectionStats(pc *webrtc.PeerConnection) (map[string]interface{}, error) {
	stats := pc.GetStats()
	result := make(map[string]interface{})

	for _, stat := range stats {
		result[stat.Type.String()] = stat
	}

	return result, nil
}

// ConfigureMediaCodecs configures media codecs for a peer connection
func ConfigureMediaCodecs(mediaEngine *webrtc.MediaEngine, videoCodec, audioCodec string) error {
	// Register video codec
	if videoCodec != "" {
		switch videoCodec {
		case "VP8":
			if err := mediaEngine.RegisterCodec(
				webrtc.RTPCodecParameters{
					RTPCodecCapability: webrtc.RTPCodecCapability{
						MimeType: webrtc.MimeTypeVP8,
					},
					PayloadType: 96,
				},
				webrtc.RTPCodecTypeVideo,
			); err != nil {
				return err
			}
		case "VP9":
			if err := mediaEngine.RegisterCodec(
				webrtc.RTPCodecParameters{
					RTPCodecCapability: webrtc.RTPCodecCapability{
						MimeType: webrtc.MimeTypeVP9,
					},
					PayloadType: 98,
				},
				webrtc.RTPCodecTypeVideo,
			); err != nil {
				return err
			}
		case "H264":
			if err := mediaEngine.RegisterCodec(
				webrtc.RTPCodecParameters{
					RTPCodecCapability: webrtc.RTPCodecCapability{
						MimeType: webrtc.MimeTypeH264,
					},
					PayloadType: 102,
				},
				webrtc.RTPCodecTypeVideo,
			); err != nil {
				return err
			}
		}
	}

	// Register audio codec
	if audioCodec != "" {
		switch audioCodec {
		case "OPUS":
			if err := mediaEngine.RegisterCodec(
				webrtc.RTPCodecParameters{
					RTPCodecCapability: webrtc.RTPCodecCapability{
						MimeType:  webrtc.MimeTypeOpus,
						ClockRate: 48000,
						Channels:  2,
					},
					PayloadType: 111,
				},
				webrtc.RTPCodecTypeAudio,
			); err != nil {
				return err
			}
		}
	}

	return nil
}
