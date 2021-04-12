package base

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

var casesResponse = []struct {
	name string
	byts []byte
	res  Response
}{
	{
		"ok",
		[]byte("RTSP/1.0 200 OK\r\n" +
			"CSeq: 2\r\n" +
			"Date: Sat, Aug 16 2014 02:22:28 GMT\r\n" +
			"Session: 645252166\r\n" +
			"WWW-Authenticate: Digest realm=\"4419b63f5e51\", nonce=\"8b84a3b789283a8bea8da7fa7d41f08b\", stale=\"FALSE\"\r\n" +
			"WWW-Authenticate: Basic realm=\"4419b63f5e51\"\r\n" +
			"\r\n",
		),
		Response{
			StatusCode:    StatusOK,
			StatusMessage: "OK",
			Header: Header{
				"CSeq":    HeaderValue{"2"},
				"Session": HeaderValue{"645252166"},
				"WWW-Authenticate": HeaderValue{
					"Digest realm=\"4419b63f5e51\", nonce=\"8b84a3b789283a8bea8da7fa7d41f08b\", stale=\"FALSE\"",
					"Basic realm=\"4419b63f5e51\"",
				},
				"Date": HeaderValue{"Sat, Aug 16 2014 02:22:28 GMT"},
			},
		},
	},
	{
		"ok with payload",
		[]byte("RTSP/1.0 200 OK\r\n" +
			"CSeq: 2\r\n" +
			"Content-Base: rtsp://example.com/media.mp4\r\n" +
			"Content-Length: 444\r\n" +
			"Content-Type: application/sdp\r\n" +
			"\r\n" +
			"m=video 0 RTP/AVP 96\n" +
			"a=control:streamid=0\n" +
			"a=range:npt=0-7.741000\n" +
			"a=length:npt=7.741000\n" +
			"a=rtpmap:96 MP4V-ES/5544\n" +
			"a=mimetype:string;\"video/MP4V-ES\"\n" +
			"a=AvgBitRate:integer;304018\n" +
			"a=StreamName:string;\"hinted video track\"\n" +
			"m=audio 0 RTP/AVP 97\n" +
			"a=control:streamid=1\n" +
			"a=range:npt=0-7.712000\n" +
			"a=length:npt=7.712000\n" +
			"a=rtpmap:97 mpeg4-generic/32000/2\n" +
			"a=mimetype:string;\"audio/mpeg4-generic\"\n" +
			"a=AvgBitRate:integer;65790\n" +
			"a=StreamName:string;\"hinted audio track\"\n",
		),
		Response{
			StatusCode:    200,
			StatusMessage: "OK",
			Header: Header{
				"Content-Base":   HeaderValue{"rtsp://example.com/media.mp4"},
				"Content-Length": HeaderValue{"444"},
				"Content-Type":   HeaderValue{"application/sdp"},
				"CSeq":           HeaderValue{"2"},
			},
			Body: []byte("m=video 0 RTP/AVP 96\n" +
				"a=control:streamid=0\n" +
				"a=range:npt=0-7.741000\n" +
				"a=length:npt=7.741000\n" +
				"a=rtpmap:96 MP4V-ES/5544\n" +
				"a=mimetype:string;\"video/MP4V-ES\"\n" +
				"a=AvgBitRate:integer;304018\n" +
				"a=StreamName:string;\"hinted video track\"\n" +
				"m=audio 0 RTP/AVP 97\n" +
				"a=control:streamid=1\n" +
				"a=range:npt=0-7.712000\n" +
				"a=length:npt=7.712000\n" +
				"a=rtpmap:97 mpeg4-generic/32000/2\n" +
				"a=mimetype:string;\"audio/mpeg4-generic\"\n" +
				"a=AvgBitRate:integer;65790\n" +
				"a=StreamName:string;\"hinted audio track\"\n",
			),
		},
	},
}

func TestResponseRead(t *testing.T) {
	// keep res global to make sure that all its fields are overridden.
	var res Response

	for _, c := range casesResponse {
		t.Run(c.name, func(t *testing.T) {
			err := res.Read(bufio.NewReader(bytes.NewBuffer(c.byts)))
			require.NoError(t, err)
			require.Equal(t, c.res, res)
		})
	}
}

func TestResponseWrite(t *testing.T) {
	for _, c := range casesResponse {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			bw := bufio.NewWriter(&buf)
			err := c.res.Write(bw)
			require.NoError(t, err)
			// do NOT call flush(), write() must have already done it
			require.Equal(t, c.byts, buf.Bytes())
		})
	}
}

func TestResponseReadErrors(t *testing.T) {
	for _, ca := range []struct {
		name string
		byts []byte
	}{
		{
			"empty",
			[]byte{},
		},
		{
			"missing code, message, eol",
			[]byte("RTSP/1.0"),
		},
		{
			"missing message, eol",
			[]byte("RTSP/1.0 200"),
		},
		{
			"missing eol",
			[]byte("RTSP/1.0 200 OK"),
		},
		{
			"missing eol 2",
			[]byte("RTSP/1.0 200 OK\r"),
		},
		{
			"invalid protocol",
			[]byte("RTSP/2.0 200 OK\r\n"),
		},
		{
			"invalid code",
			[]byte("RTSP/2.0 string OK\r\n"),
		},
		{
			"empty message",
			[]byte("RTSP/2.0 string \r\n"),
		},
	} {
		t.Run(ca.name, func(t *testing.T) {
			var res Response
			err := res.Read(bufio.NewReader(bytes.NewBuffer(ca.byts)))
			require.Error(t, err)
		})
	}
}

func TestResponseWriteAutoFillStatus(t *testing.T) {
	res := &Response{
		StatusCode: StatusMethodNotAllowed,
		Header: Header{
			"CSeq":    HeaderValue{"2"},
			"Session": HeaderValue{"645252166"},
			"WWW-Authenticate": HeaderValue{
				"Digest realm=\"4419b63f5e51\", nonce=\"8b84a3b789283a8bea8da7fa7d41f08b\", stale=\"FALSE\"",
				"Basic realm=\"4419b63f5e51\"",
			},
			"Date": HeaderValue{"Sat, Aug 16 2014 02:22:28 GMT"},
		},
	}
	byts := []byte("RTSP/1.0 405 Method Not Allowed\r\n" +
		"CSeq: 2\r\n" +
		"Date: Sat, Aug 16 2014 02:22:28 GMT\r\n" +
		"Session: 645252166\r\n" +
		"WWW-Authenticate: Digest realm=\"4419b63f5e51\", nonce=\"8b84a3b789283a8bea8da7fa7d41f08b\", stale=\"FALSE\"\r\n" +
		"WWW-Authenticate: Basic realm=\"4419b63f5e51\"\r\n" +
		"\r\n",
	)

	var buf bytes.Buffer
	bw := bufio.NewWriter(&buf)
	err := res.Write(bw)
	require.NoError(t, err)
	require.Equal(t, byts, buf.Bytes())
}

func TestReadIgnoreFrames(t *testing.T) {
	byts := []byte{0x24, 0x6, 0x0, 0x4, 0x1, 0x2, 0x3, 0x4}
	byts = append(byts, []byte("RTSP/1.0 200 OK\r\n"+
		"CSeq: 1\r\n"+
		"Public: DESCRIBE, SETUP, TEARDOWN, PLAY, PAUSE\r\n"+
		"\r\n")...)

	rb := bufio.NewReader(bytes.NewBuffer(byts))
	buf := make([]byte, 10)
	var res Response
	err := res.ReadIgnoreFrames(rb, buf)
	require.NoError(t, err)
}
