package server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"net"
	"os"

	"portfolio-server/internal/data"
	"portfolio-server/internal/discord"
	"portfolio-server/internal/sshbox"

	"golang.org/x/crypto/ssh"
)

// ─────────────────────────────────────────────
//  HOST KEY MANAGEMENT
// ─────────────────────────────────────────────

const hostKeyFile = "ssh_host_key"

func loadOrGenerateHostKey() (ssh.Signer, error) {
	if data, err := os.ReadFile(hostKeyFile); err == nil {
		block, _ := pem.Decode(data)
		if block != nil {
			key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err == nil {
				return ssh.NewSignerFromKey(key)
			}
		}
	}

	key, err := rsa.GenerateKey(rand.Reader, 3072)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(hostKeyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if err := pem.Encode(f, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}); err != nil {
		return nil, err
	}

	log.Println("[ssh] generated new RSA host key →", hostKeyFile)
	return ssh.NewSignerFromKey(key)
}

// ─────────────────────────────────────────────
//  SSH SERVER
// ─────────────────────────────────────────────

func StartSSH(addr string, store *data.Store, dc *discord.Client) {
	signer, err := loadOrGenerateHostKey()
	if err != nil {
		log.Fatalf("[ssh] host key error: %v", err)
	}

	cfg := &ssh.ServerConfig{
		NoClientAuth: true,
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			return nil, nil // Accept literally any password
		},
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			return nil, nil // Accept literally any key
		},
	}
	cfg.AddHostKey(signer)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("[ssh] listen %s: %v", addr, err)
	}
	log.Printf("[ssh] listening on %s", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[ssh] accept error: %v", err)
			continue
		}
		go handleSSHConn(conn, cfg, store, dc)
	}
}

func handleSSHConn(netConn net.Conn, cfg *ssh.ServerConfig, store *data.Store, dc *discord.Client) {
	defer netConn.Close()
	sshConn, chans, reqs, err := ssh.NewServerConn(netConn, cfg)
	if err != nil {
		log.Printf("[ssh] handshake error: %v", err)
		return
	}
	defer sshConn.Close()
	log.Printf("[ssh] connection from %s", sshConn.RemoteAddr())

	go ssh.DiscardRequests(reqs)

	for newChan := range chans {
		if newChan.ChannelType() != "session" {
			_ = newChan.Reject(ssh.UnknownChannelType, "unsupported channel type")
			continue
		}
		ch, requests, err := newChan.Accept()
		if err != nil {
			return
		}

		go func(in <-chan *ssh.Request) {
			for req := range in {
				switch req.Type {
				case "pty-req", "shell", "exec", "window-change":
					if req.WantReply {
						_ = req.Reply(true, nil)
					}
				default:
					if req.WantReply {
						_ = req.Reply(false, nil)
					}
				}
			}
		}(requests)

		// Start the interactive portfolio shell
		sshbox.RunShell(ch, store, dc)
	}
}
