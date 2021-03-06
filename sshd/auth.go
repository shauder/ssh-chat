package sshd

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net"

	"golang.org/x/crypto/ssh"
)

// Auth is used to authenticate connections based on public keys.
type Auth interface {
	// Whether to allow connections without a public key.
	AllowAnonymous() bool
	// Given address and public key, return if the connection should be permitted.
	Check(net.Addr, ssh.PublicKey) (bool, error)
}

// MakeAuth makes an ssh.ServerConfig which performs authentication against an Auth implementation.
func MakeAuth(auth Auth) *ssh.ServerConfig {
	config := ssh.ServerConfig{
		NoClientAuth: false,
		// Auth-related things should be constant-time to avoid timing attacks.
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			ok, err := auth.Check(conn.RemoteAddr(), key)
			if !ok {
				return nil, err
			}
			perm := &ssh.Permissions{Extensions: map[string]string{
				"pubkey": string(key.Marshal()),
			}}
			return perm, nil
		},
		KeyboardInteractiveCallback: func(conn ssh.ConnMetadata, challenge ssh.KeyboardInteractiveChallenge) (*ssh.Permissions, error) {
			if !auth.AllowAnonymous() {
				return nil, errors.New("public key authentication required")
			}
			_, err := auth.Check(conn.RemoteAddr(), nil)
			return nil, err
		},
	}

	return &config
}

// MakeNoAuth makes a simple ssh.ServerConfig which allows all connections.
// Primarily used for testing.
func MakeNoAuth() *ssh.ServerConfig {
	config := ssh.ServerConfig{
		NoClientAuth: false,
		// Auth-related things should be constant-time to avoid timing attacks.
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			perm := &ssh.Permissions{Extensions: map[string]string{
				"pubkey": string(key.Marshal()),
			}}
			return perm, nil
		},
		KeyboardInteractiveCallback: func(conn ssh.ConnMetadata, challenge ssh.KeyboardInteractiveChallenge) (*ssh.Permissions, error) {
			return nil, nil
		},
	}

	return &config
}

// Fingerprint performs a SHA256 BASE64 fingerprint of the PublicKey, similar to OpenSSH.
// See: https://anongit.mindrot.org/openssh.git/commit/?id=56d1c83cdd1ac
func Fingerprint(k ssh.PublicKey) string {
	hash := sha256.Sum256(k.Marshal())
	return base64.StdEncoding.EncodeToString(hash[:])
}
