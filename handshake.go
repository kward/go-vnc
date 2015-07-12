// Implementation of RFC 6143 §7.1 Handshake Messages.

package vnc

import (
	"fmt"
	"log"
)

const pvLen = 12 // ProtocolVersion message length.

func parseProtocolVersion(pv []byte) (uint, uint, error) {
	var major, minor uint

	if len(pv) < pvLen {
		return 0, 0, fmt.Errorf("ProtocolVersion message too short (%v < %v)", len(pv), pvLen)
	}

	l, err := fmt.Sscanf(string(pv), "RFB %d.%d\n", &major, &minor)
	if l != 2 {
		return 0, 0, fmt.Errorf("error parsing ProtocolVersion.")
	}
	if err != nil {
		return 0, 0, err
	}

	return major, minor, nil
}

const (
	// Client ProtocolVersions.
	PROTO_VERS_UNSUP = "UNSUPPORTED"
	PROTO_VERS_3_3   = "RFB 003.003\n"
	PROTO_VERS_3_8   = "RFB 003.008\n"
)

// protocolVersionHandshake implements §7.1.1 ProtocolVersion Handshake.
func (c *ClientConn) protocolVersionHandshake() error {
	if c.debug {
		log.Print("protocolVersionHandshake()")
	}

	var protocolVersion [pvLen]byte

	// Read the ProtocolVersion message sent by the server.
	if err := c.receive(&protocolVersion); err != nil {
		return err
	}
	if c.debug {
		log.Printf("protocolVersion: %v", string(protocolVersion[:]))
	}

	major, minor, err := parseProtocolVersion(protocolVersion[:])
	if err != nil {
		return err
	}
	pv := PROTO_VERS_UNSUP
	if major == 3 {
		if minor >= 8 {
			pv = PROTO_VERS_3_8
		} else if minor >= 3 {
			pv = PROTO_VERS_3_3
		}
	}
	if pv == PROTO_VERS_UNSUP {
		return NewVNCError(fmt.Sprintf("ProtocolVersion handshake failed; unsupported version '%v'", string(protocolVersion[:])))
	}
	c.protocolVersion = pv

	// Respond with the version we will support
	if c.debug {
		log.Printf("supported protocolVersion: %v", pv)
	}
	if err = c.send([]byte(pv)); err != nil {
		return err
	}

	return nil
}

// securityHandshake implements §7.1.2 Security Handshake.
func (c *ClientConn) securityHandshake() error {
	if c.debug {
		log.Print("securityHandshake()")
	}

	switch c.protocolVersion {
	case PROTO_VERS_3_3:
		if err := c.securityHandshake33(); err != nil {
			return err
		}
	case PROTO_VERS_3_8:
		if err := c.securityHandshake38(); err != nil {
			return err
		}
	default:
		return NewVNCError(fmt.Sprintf("Security handshake failed; unsupported protocol"))
	}
	return nil
}

func (c *ClientConn) securityHandshake33() error {
	if c.debug {
		log.Print("securityHandshake33()")
	}

	var secType uint32
	if err := c.receive(&secType); err != nil {
		return err
	}

	var auth ClientAuth
	switch secType {
	case secTypeInvalid: // Connection failed.
		reason, err := c.readErrorReason()
		if err != nil {
			return err
		}
		return NewVNCError(fmt.Sprintf("Security handshake failed; connection failed: %s", reason))
	case secTypeNone:
		auth = &ClientAuthNone{}
	case secTypeVNCAuth:
		auth = &ClientAuthVNC{c.config.Password}
	default:
		return NewVNCError(fmt.Sprintf("Security handshake failed; invalid security type: %v", secType))
	}
	if err := auth.Handshake(c); err != nil {
		return err
	}

	return nil
}

func (c *ClientConn) securityHandshake38() error {
	if c.debug {
		log.Print("securityHandshake38()")
	}

	// Determine server supported security types.
	var numSecurityTypes uint8
	if err := c.receive(&numSecurityTypes); err != nil {
		return err
	}
	if numSecurityTypes == 0 {
		reason, err := c.readErrorReason()
		if err != nil {
			return err
		}
		return NewVNCError(fmt.Sprintf("Security handshake failed; no security types: %v", reason))
	}
	securityTypes := make([]uint8, numSecurityTypes)
	if err := c.receive(&securityTypes); err != nil {
		return err
	}

	// Choose client security type.
	// TODO(kward): try "better" security types first.
	var auth ClientAuth
FindAuth:
	for _, securityType := range securityTypes {
		for _, a := range c.config.Auth {
			if a.SecurityType() == securityType {
				// We use the first matching supported authentication.
				auth = a
				break FindAuth
			}
		}
	}
	if auth == nil {
		return NewVNCError(fmt.Sprintf("Security handshake failed; no suitable auth schemes found; server supports: %#v", securityTypes))
	}
	if err := c.send(auth.SecurityType()); err != nil {
		return err
	}

	if err := auth.Handshake(c); err != nil {
		return err
	}
	return nil
}

// securityResultHandshake implements §7.1.3 SecurityResult Handshake.
func (c *ClientConn) securityResultHandshake() error {
	if c.debug {
		log.Print("securityResultHandshake()")
	}

	if c.protocolVersion == PROTO_VERS_3_3 {
		// Not required for 3.3.
		return nil
	}

	var securityResult uint32
	if err := c.receive(&securityResult); err != nil {
		return err
	}
	if securityResult == 1 {
		reason, err := c.readErrorReason()
		if err != nil {
			return err
		}
		return NewVNCError(fmt.Sprintf("SecurityResult handshake failed: %s", reason))
	}
	return nil
}

// TODO(kward): need a context for timeout
func (c *ClientConn) readErrorReason() (string, error) {
	var reasonLen uint32
	if err := c.receive(&reasonLen); err != nil {
		return "", err
	}

	reason := make([]uint8, reasonLen)
	if err := c.receive(&reason); err != nil {
		return "", err
	}

	return string(reason), nil
}
