package conn

import (
	"fmt"
	"net"
)

// MHFConn wraps a CryptConn and provides convenience methods for MHF connections.
type MHFConn struct {
	*CryptConn
	RawConn net.Conn
}

// DialWithInit connects to addr and sends the 8 NULL byte initialization
// required by sign and entrance servers.
func DialWithInit(addr string) (*MHFConn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}

	// Sign and entrance servers expect 8 NULL bytes to initialize the connection.
	_, err = conn.Write(make([]byte, 8))
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("write init bytes to %s: %w", addr, err)
	}

	return &MHFConn{
		CryptConn: NewCryptConn(conn),
		RawConn:   conn,
	}, nil
}

// DialDirect connects to addr without sending initialization bytes.
// Used for channel server connections.
func DialDirect(addr string) (*MHFConn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}

	return &MHFConn{
		CryptConn: NewCryptConn(conn),
		RawConn:   conn,
	}, nil
}

// Close closes the underlying connection.
func (c *MHFConn) Close() error {
	return c.RawConn.Close()
}
