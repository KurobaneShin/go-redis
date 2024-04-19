package main

import (
	"log"
	"log/slog"
	"net"
)

const defaultListenerAddress = ":5001"

type Config struct {
	ListenAdrress string
}

type Server struct {
	Config
	peers     map[*Peer]bool
	ln        net.Listener
	addPeerCh chan *Peer
	quitCh    chan struct{}
}

func NewServer(cfg Config) *Server {
	if len(cfg.ListenAdrress) == 0 {
		cfg.ListenAdrress = defaultListenerAddress
	}
	return &Server{
		Config:    cfg,
		peers:     make(map[*Peer]bool),
		addPeerCh: make(chan *Peer),
		quitCh:    make(chan struct{}),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAdrress)
	if err != nil {
		return nil
	}

	s.ln = ln
	go s.loop()

	slog.Info("server running", "ListenAdrress", s.ListenAdrress)

	return s.acceptLoop()
}

func (s *Server) loop() {
	for {
		select {
		case <-s.quitCh:
			return
		case peer := <-s.addPeerCh:
			s.peers[peer] = true
		}
	}
}

func (s *Server) acceptLoop() error {
	for {

		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("accept error", "err", err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	peer := NewPeer(conn)
	s.addPeerCh <- peer

	peer.readLoop()
}

func main() {
	server := NewServer(Config{})

	log.Fatal(server.Start())
}