package music

import (
	"testing"

	"github.com/metwurcht/torrent-all-in-one/internal/mediainfo"
	"github.com/metwurcht/torrent-all-in-one/internal/theaudiodb"
)

func TestMusicRenamerUsesTheRequiredReleaseFormat(t *testing.T) {
	renamer := NewRenamer("NOTAG")
	album := &theaudiodb.Album{Artist: "The Weeknd", Title: "Blinding Lights", Date: "2019"}
	info := &mediainfo.MediaInfo{Audio: []mediainfo.AudioInfo{{
		Codec: "FLAC", BitDepth: 16, SampleRate: 44100,
	}}}

	got := renamer.GenerateDirectoryName(album, info)
	want := "The.Weeknd.Blinding.Lights.2019.FLAC[16bit.44.1kHz]-NOTAG"
	if got != want {
		t.Fatalf("unexpected release name: got %q, want %q", got, want)
	}
}

func TestMusicRenamerUsesBitrateForLossyAudio(t *testing.T) {
	renamer := NewRenamer("NOTAG")
	album := &theaudiodb.Album{Artist: "Demi Lovato", Title: "It's Not That Deep (Little Bit Extra Version)", Date: "2025"}
	info := &mediainfo.MediaInfo{Audio: []mediainfo.AudioInfo{{
		Codec: "MPEG Audio", Bitrate: 320000,
	}}}

	got := renamer.GenerateDirectoryName(album, info)
	want := "Demi.Lovato.It.s.Not.That.Deep.Little.Bit.Extra.Version.2025.MP3[320kbps]-NOTAG"
	if got != want {
		t.Fatalf("unexpected release name: got %q, want %q", got, want)
	}
}
