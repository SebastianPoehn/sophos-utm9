// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Transport interface is used by connections to transport data to and from
// the confd
type Transport interface {
	Connect(url *url.URL) error
	IsConnected() bool
	io.Closer
	http.RoundTripper
}

// TCPTransport implements a tcp+http RoundTripper for confd connections
type tcpTransport struct {
	Timeout     time.Duration // Timeout specifies the conn read/write timeout
	LastRequest time.Time     // LastRequest last time a request was done
	conn        *net.TCPConn
	mu          sync.RWMutex
}

// Connect to the passed url
func (t *tcpTransport) Connect(url *url.URL) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	conn, err := net.Dial("tcp", url.Host)
	if err != nil {
		return err
	}
	t.conn = conn.(*net.TCPConn)
	return nil
}

// RoundTrip executes a request/response round trip
func (t *tcpTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	// send to remote side and recieve response
	if t.conn == nil {
		return nil, fmt.Errorf("Called confd.tcpTransport.RoundTrip " +
			"without being connected!")
	}
	err = t.conn.SetDeadline(time.Now().Add(t.Timeout))
	if err != nil {
		goto err
	}

	err = req.Write(t.conn)
	if err != nil {
		goto err
	}

	// read response
	resp, err = http.ReadResponse(bufio.NewReader(t.conn), nil)
	t.LastRequest = time.Now()
	if err != nil {
		goto err
	}

	return
err:
	// close the connection on transport errors, so that we require a reconnect
	_ = t.close() // ignore errors
	return
}

// IsConnected returns
func (t *tcpTransport) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.conn != nil
}

// Close the transport
func (t *tcpTransport) Close() (err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.conn == nil {
		return // we already disconnected
	}
	return t.close()
}

// close the connection
func (t *tcpTransport) close() (err error) {
	err = t.conn.Close()
	t.conn = nil
	return
}
