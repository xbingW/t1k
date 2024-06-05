package t1k

import (
	"errors"
	"io"
	"net/http"
)

type ResultFlag string

const (
	ResultFlagAllowed ResultFlag = "."
	ResultFlagBlocked ResultFlag = "?"
)

func (d ResultFlag) Byte() byte {
	return d[0]
}

type DetectorResponse struct {
	Head        byte
	Body        []byte
	Delay       []byte
	ExtraHeader []byte
	ExtraBody   []byte
	Context     []byte
	Cookie      []byte
	WebLog      []byte
	BotQuery    []byte
	BotBody     []byte
	Forward     []byte
}

func (r *DetectorResponse) Allowed() bool {
	return r.Head == ResultFlagAllowed.Byte()
}

type HttpDetector struct {
	extra *HttpExtra
	req   Request
	resp  Response
}

func NewHttpDetector(req *http.Request, extra *HttpExtra) *HttpDetector {
	return &HttpDetector{
		req:   NewHttpRequest(req, extra),
		extra: extra,
	}
}

func (d *HttpDetector) SetResponse(resp *http.Response) *HttpDetector {
	d.resp = NewHttpResponse(d.req, resp, d.extra)
	return d
}

func (d *HttpDetector) DetectRequest(socket io.ReadWriter) (*DetectorResponse, error) {
	raw, err := d.req.Serialize()
	if err != nil {
		return nil, err
	}
	_, err = socket.Write(raw)
	if err != nil {
		return nil, err
	}
	return d.ReadResponse(socket)
}

func (d *HttpDetector) DetectResponse(socket io.ReadWriter) (*DetectorResponse, error) {
	raw, err := d.resp.Serialize()
	if err != nil {
		return nil, err
	}
	_, err = socket.Write(raw)
	if err != nil {
		return nil, err
	}
	return d.ReadResponse(socket)
}

func (d *HttpDetector) ReadResponse(r io.Reader) (*DetectorResponse, error) {
	res := &DetectorResponse{}
	for {
		p, err := ReadPacket(r)
		if err != nil {
			return nil, err
		}
		switch p.Tag().Strip() {
		case TAG_HEADER:
			if len(p.PayLoad()) != 1 {
				return nil, errors.New("len(T1K_HEADER) != 1")
			}
			res.Head = p.PayLoad()[0]
		case TAG_DELAY:
			res.Delay = p.PayLoad()
		case TAG_BODY:
			res.Body = p.PayLoad()
		case TAG_EXTRA_HEADER:
			res.ExtraHeader = p.PayLoad()
		case TAG_EXTRA_BODY:
			res.ExtraBody = p.PayLoad()
		case TAG_CONTEXT:
			res.Context = p.PayLoad()
		case TAG_COOKIE:
			res.Cookie = p.PayLoad()
		case TAG_WEB_LOG:
			res.WebLog = p.PayLoad()
		case TAG_BOT_QUERY:
			res.BotQuery = p.PayLoad()
		case TAG_BOT_BODY:
			res.BotBody = p.PayLoad()
		case TAG_FORWARD:
			res.Forward = p.PayLoad()
		}
		if p.Last() {
			break
		}
	}
	return res, nil
}
