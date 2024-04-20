package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"time"

	"go-redis/client"
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
	msgCh     chan []byte
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
		msgCh:     make(chan []byte),
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

func (s *Server) handleRawMsg(rawMsg []byte) error {
	cmd, err := parseCommand(string(rawMsg))
	if err != nil {
		return err
	}

	switch v := cmd.(type) {
	case SetCommand:
		slog.Info("set command", "key", v.key, "val", v.val)
	}
	return nil
}

func (s *Server) loop() {
	for {
		select {
		case rawMsg := <-s.msgCh:
			if err := s.handleRawMsg(rawMsg); err != nil {
				slog.Error("handleRawMsg error", "err", err)
			}
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
	peer := NewPeer(conn, s.msgCh)
	s.addPeerCh <- peer

	slog.Info("new peer connected", "addr", conn.RemoteAddr().String())

	if err := peer.readLoop(); err != nil {
		slog.Error("perr read error", "err", err, "remoteAddr", conn.RemoteAddr())
	}
}

func main() {
	go func() {
		server := NewServer(Config{})

		log.Fatal(server.Start())
	}()

	time.Sleep(time.Second)

	client := client.New("localhost:5001")

	if err := client.Set(context.Background(), "foo", "bar"); err != nil {
		log.Fatal(err)
	}

	select {} // gambi to server not hangout
}
