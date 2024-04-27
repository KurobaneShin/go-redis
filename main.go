package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"

	"github.com/tidwall/resp"
)

const defaultListenerAddress = ":5001"

type Config struct {
	ListenAdrress string
}

type Message struct {
	cmd  Command
	peer *Peer
}

type Server struct {
	Config
	peers     map[*Peer]bool
	ln        net.Listener
	addPeerCh chan *Peer
	delPeerCh chan *Peer
	quitCh    chan struct{}
	msgCh     chan Message

	kv *KV
}

func NewServer(cfg Config) *Server {
	if len(cfg.ListenAdrress) == 0 {
		cfg.ListenAdrress = defaultListenerAddress
	}
	return &Server{
		Config:    cfg,
		peers:     make(map[*Peer]bool),
		addPeerCh: make(chan *Peer),
		delPeerCh: make(chan *Peer),
		quitCh:    make(chan struct{}),
		msgCh:     make(chan Message),
		kv:        NewKV(),
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

func (s *Server) handleMessage(msg Message) error {
	switch v := msg.cmd.(type) {
	case ClientCommand:
		if err := resp.
			NewWriter(msg.peer.conn).
			WriteString("OK"); err != nil {
			return err
		}
	case SetCommand:
		if err := s.kv.Set(v.key, v.val); err != nil {
			return err
		}
		if err := resp.
			NewWriter(msg.peer.conn).
			WriteString("OK"); err != nil {
			return err
		}
	case GetCommand:
		val, ok := s.kv.Get(v.key)
		if !ok {
			return fmt.Errorf("key not found")
		}
		if err := resp.
			NewWriter(msg.peer.conn).
			WriteString(string(val)); err != nil {
			return err
		}
	case HelloCommand:
		spec := map[string]string{
			"server": "redis",
		}
		_, err := msg.peer.Send(respWriteMap(spec))
		if err != nil {
			return fmt.Errorf("peer send error: %s", err)
		}
	}
	return nil
}

func (s *Server) loop() {
	for {
		select {
		case msg := <-s.msgCh:
			if err := s.handleMessage(msg); err != nil {
				slog.Error("handleRawMsg error", "err", err)
			}
		case <-s.quitCh:
			return
		case peer := <-s.addPeerCh:
			slog.Info("new peer", "remoteAddr", peer.conn.RemoteAddr())
			s.peers[peer] = true
		case peer := <-s.delPeerCh:
			slog.Info("peer removed", "remoteAddr", peer.conn.RemoteAddr())
			delete(s.peers, peer)
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
	peer := NewPeer(conn, s.msgCh, s.delPeerCh)
	s.addPeerCh <- peer

	if err := peer.readLoop(); err != nil {
		slog.Error("perr read error", "err", err, "remoteAddr", conn.RemoteAddr())
	}
}

func main() {
	f := flag.String("listenAddr", defaultListenerAddress, "server listen address")
	flag.Parse()
	server := NewServer(Config{
		ListenAdrress: *f,
	})
	log.Fatal(server.Start())
}
