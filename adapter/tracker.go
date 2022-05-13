package adapter

import (
	"net"
	"time"

	"github.com/Dreamacro/clash/constant"
)

type Tracker interface {
	IncrDownload(size uint64)
	IncrUpload(size uint64)

	Download() uint64
	Upload() uint64
}

type tracker struct {
	conn     constant.Conn      `json:"-"`
	metadata *constant.Metadata `json:"metadata"`

	tracker Tracker
	t       Tracker
}

func (p *tracker) Read(b []byte) (n int, err error) {
	n, err = p.conn.Read(b)
	p.t.IncrDownload(uint64(n))
	return n, err
}

func (p *tracker) Write(b []byte) (n int, err error) {
	n, err = p.conn.Write(b)
	p.t.IncrUpload(uint64(n))
	return n, err
}

func (p *tracker) Close() error {
	return p.conn.Close()
}

func (p *tracker) LocalAddr() net.Addr {
	return p.conn.LocalAddr()
}

func (p *tracker) RemoteAddr() net.Addr {
	return p.conn.RemoteAddr()
}

func (p *tracker) SetDeadline(t time.Time) error {
	return p.conn.SetDeadline(t)
}

func (p *tracker) SetReadDeadline(t time.Time) error {
	return p.conn.SetReadDeadline(t)
}

func (p *tracker) SetWriteDeadline(t time.Time) error {
	return p.conn.SetWriteDeadline(t)
}

func newTicker(conn constant.Conn, metadata *constant.Metadata, t Tracker) *tracker {
	return &tracker{
		conn:     conn,
		metadata: metadata,
		t:        t,
	}
}
