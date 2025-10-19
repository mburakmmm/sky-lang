package skylib

import (
	"net"
	"time"
)

// NetConn wraps net.Conn
type NetConn struct {
	conn net.Conn
}

// TCPConnect connects to TCP address
func NetTCPConnect(addr string) (*NetConn, error) {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &NetConn{conn: conn}, nil
}

// TCPListen listens on TCP address
func NetTCPListen(addr string) (*NetListener, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &NetListener{listener: listener}, nil
}

// NetListener wraps net.Listener
type NetListener struct {
	listener net.Listener
}

// Accept accepts a connection
func (l *NetListener) Accept() (*NetConn, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	return &NetConn{conn: conn}, nil
}

// Close closes the listener
func (l *NetListener) Close() error {
	return l.listener.Close()
}

// Read reads from connection
func (c *NetConn) Read(n int) ([]byte, error) {
	buf := make([]byte, n)
	bytesRead, err := c.conn.Read(buf)
	return buf[:bytesRead], err
}

// Write writes to connection
func (c *NetConn) Write(data []byte) (int, error) {
	return c.conn.Write(data)
}

// Close closes the connection
func (c *NetConn) Close() error {
	return c.conn.Close()
}

// SetTimeout sets read/write timeout
func (c *NetConn) SetTimeout(ms int) error {
	deadline := time.Now().Add(time.Duration(ms) * time.Millisecond)
	return c.conn.SetDeadline(deadline)
}

// UDPConn wraps UDP connection
type UDPConn struct {
	conn *net.UDPConn
}

// NewUDPConn creates UDP connection
func NetUDPListen(addr string) (*UDPConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	return &UDPConn{conn: conn}, nil
}

// SendTo sends data to address
func (u *UDPConn) SendTo(data []byte, addr string) (int, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return 0, err
	}
	return u.conn.WriteToUDP(data, udpAddr)
}

// RecvFrom receives data
func (u *UDPConn) RecvFrom(n int) ([]byte, string, error) {
	buf := make([]byte, n)
	bytesRead, addr, err := u.conn.ReadFromUDP(buf)
	if err != nil {
		return nil, "", err
	}
	return buf[:bytesRead], addr.String(), nil
}

// Resolve resolves hostname
func NetResolve(host string) ([]string, error) {
	addrs, err := net.LookupHost(host)
	return addrs, err
}
