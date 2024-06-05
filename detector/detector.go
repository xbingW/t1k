package detector

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/xbingW/t1k/pkg/datetime"
	"github.com/xbingW/t1k/pkg/rand"
	"github.com/xbingW/t1k/pkg/t1k"
)

type Detector struct {
	socket io.ReadWriter
}

func NewDetector(addr string) (*Detector, error) {
	socket, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Detector{
		socket: socket,
	}, nil
}

func (d *Detector) DetectorRequestStr(req string) (*t1k.DetectorResponse, error) {
	httpReq, err := http.ReadRequest(bufio.NewReader(strings.NewReader(req)))
	if err != nil {
		return nil, fmt.Errorf("read request failed: %v", err)
	}
	return d.DetectorRequest(httpReq)
}

func (d *Detector) DetectorRequest(req *http.Request) (*t1k.DetectorResponse, error) {
	extra, err := d.GenerateExtra(req)
	if err != nil {
		return nil, fmt.Errorf("generate extra failed: %v", err)
	}
	dc := t1k.NewHttpDetector(req, extra)
	return dc.DetectRequest(d.socket)
}

func (d *Detector) DetectorResponseStr(req string, resp string) (*t1k.DetectorResponse, error) {
	httpReq, err := http.ReadRequest(bufio.NewReader(strings.NewReader(req)))
	if err != nil {
		return nil, fmt.Errorf("read request failed: %v", err)
	}
	httpResp, err := http.ReadResponse(bufio.NewReader(strings.NewReader(resp)), httpReq)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %v", err)
	}
	extra, err := d.GenerateExtra(httpReq)
	if err != nil {
		return nil, fmt.Errorf("generate extra failed: %v", err)
	}
	return t1k.NewHttpDetector(httpReq, extra).SetResponse(httpResp).DetectResponse(d.socket)
}

func (d *Detector) DetectorResponse(req *http.Request, resp *http.Response) (*t1k.DetectorResponse, error) {
	extra, err := d.GenerateExtra(req)
	if err != nil {
		return nil, fmt.Errorf("generate extra failed: %v", err)
	}
	return t1k.NewHttpDetector(req, extra).SetResponse(resp).DetectResponse(d.socket)
}

func (d *Detector) GenerateExtra(req *http.Request) (*t1k.HttpExtra, error) {
	clientHost, clientPort, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return nil, err
	}
	serverHost, serverPort := req.Host, "80"
	if hasPort(req.Host) {
		serverHost, serverPort, err = net.SplitHostPort(req.Host)
		if err != nil {
			return nil, err
		}
	}
	return &t1k.HttpExtra{
		UpstreamAddr:  "",
		RemoteAddr:    clientHost,
		RemotePort:    clientPort,
		LocalAddr:     serverHost,
		LocalPort:     serverPort,
		ServerName:    "",
		Schema:        req.URL.Scheme,
		ProxyName:     "",
		UUID:          rand.String(12),
		HasRspIfOK:    "y",
		HasRspIfBlock: "n",
		ReqBeginTime:  strconv.FormatInt(datetime.Now(), 10),
		ReqEndTime:    "",
		RspBeginTime:  strconv.FormatInt(datetime.Now(), 10),
		RepEndTime:    "",
	}, nil
}

// has port check if host has port
func hasPort(host string) bool {
	return strings.LastIndex(host, ":") > strings.LastIndex(host, "]")
}
