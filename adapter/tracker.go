package adapter

import (
	"net"
	"time"

	"github.com/Dreamacro/clash/constant"
)

type tracker struct {
	conn     constant.Conn      `json:"-"`
	metadata *constant.Metadata `json:"metadata"`
	proxy    *ProxyAdapter
}

func (p *tracker) Read(b []byte) (n int, err error) {
	n, err = p.conn.Read(b)
	p.proxy.download.Add(uint64(n))
	return n, err
}

func (p *tracker) Write(b []byte) (n int, err error) {
	n, err = p.conn.Write(b)
	p.proxy.upload.Add(uint64(n))
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

func newTicker(conn constant.Conn, metadata *constant.Metadata, proxy *ProxyAdapter) *tracker {
	return &tracker{
		conn:     conn,
		metadata: metadata,
		proxy:    proxy,
	}
}
