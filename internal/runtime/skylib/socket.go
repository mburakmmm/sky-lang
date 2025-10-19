package skylib

import (
	"net"
	"syscall"
)

// Socket represents a low-level socket
type Socket struct {
	fd     int
	domain int
	sotype int
}

// SocketCreate creates a new socket
func SocketCreate(domain, sotype, protocol int) (*Socket, error) {
	fd, err := syscall.Socket(domain, sotype, protocol)
	if err != nil {
		return nil, err
	}

	return &Socket{
		fd:     fd,
		domain: domain,
		sotype: sotype,
	}, nil
}

// Bind binds socket to address
func (s *Socket) Bind(addr string) error {
	// Parse address
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}

	// Resolve address
	tcpAddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return err
	}

	// Convert to syscall.Sockaddr
	sa := &syscall.SockaddrInet4{
		Port: tcpAddr.Port,
	}
	copy(sa.Addr[:], tcpAddr.IP.To4())

	return syscall.Bind(s.fd, sa)
}

// Listen listens for connections
func (s *Socket) Listen(backlog int) error {
	return syscall.Listen(s.fd, backlog)
}

// Accept accepts a connection
func (s *Socket) Accept() (*Socket, error) {
	nfd, _, err := syscall.Accept(s.fd)
	if err != nil {
		return nil, err
	}

	return &Socket{
		fd:     nfd,
		domain: s.domain,
		sotype: s.sotype,
	}, nil
}

// Connect connects to address
func (s *Socket) Connect(addr string) error {
	// Parse address
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}

	// Resolve address
	tcpAddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return err
	}

	// Convert to syscall.Sockaddr
	sa := &syscall.SockaddrInet4{
		Port: tcpAddr.Port,
	}
	copy(sa.Addr[:], tcpAddr.IP.To4())

	return syscall.Connect(s.fd, sa)
}

// Send sends data
func (s *Socket) Send(data []byte) (int, error) {
	return syscall.Write(s.fd, data)
}

// Recv receives data
func (s *Socket) Recv(n int) ([]byte, error) {
	buf := make([]byte, n)
	bytesRead, err := syscall.Read(s.fd, buf)
	return buf[:bytesRead], err
}

// Close closes the socket
func (s *Socket) Close() error {
	return syscall.Close(s.fd)
}

// SetSockOpt sets socket option
func (s *Socket) SetSockOpt(level, opt, value int) error {
	return syscall.SetsockoptInt(s.fd, level, opt, value)
}

// GetSockOpt gets socket option
func (s *Socket) GetSockOpt(level, opt int) (int, error) {
	return syscall.GetsockoptInt(s.fd, level, opt)
}

// Socket constants
const (
	AF_INET  = syscall.AF_INET
	AF_INET6 = syscall.AF_INET6
	AF_UNIX  = syscall.AF_UNIX

	SOCK_STREAM = syscall.SOCK_STREAM
	SOCK_DGRAM  = syscall.SOCK_DGRAM

	SOL_SOCKET   = syscall.SOL_SOCKET
	SO_REUSEADDR = syscall.SO_REUSEADDR
	SO_KEEPALIVE = syscall.SO_KEEPALIVE
	SO_BROADCAST = syscall.SO_BROADCAST
)
