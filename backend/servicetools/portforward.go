package servicetools

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// tunnel represents a single running in-process TCP forwarder: a listener on
// the local (Tailscale) side plus a cancel func used to tear down every
// connection spawned from it.
type tunnel struct {
	listener net.Listener
	cancel   context.CancelFunc
}

var (
	mu      sync.Mutex
	tunnels = make(map[int]*tunnel)
)

// StartTunnel starts an in-process TCP tunnel identified by id: it listens on
// localIP:localPort and forwards every accepted connection to
// remoteIP:remotePort, copying bytes in both directions until either side
// closes. This replaces shelling out to an external "./tcp" binary - the
// forwarding logic now lives entirely inside this process.
func StartTunnel(parent context.Context, id int, localIP, localPort, remoteIP, remotePort string) error {
	mu.Lock()
	if _, exists := tunnels[id]; exists {
		mu.Unlock()
		return fmt.Errorf("tunnel for service %d is already running", id)
	}
	mu.Unlock()

	addr := net.JoinHostPort(localIP, localPort)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to bind %s: %w", addr, err)
	}

	ctx, cancel := context.WithCancel(parent)
	t := &tunnel{listener: ln, cancel: cancel}

	mu.Lock()
	tunnels[id] = t
	mu.Unlock()

	remoteAddr := net.JoinHostPort(remoteIP, remotePort)

	// Closing the listener when ctx is cancelled unblocks Accept().
	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	go acceptLoop(ctx, id, ln, remoteAddr)

	log.Printf("tunnel %d: forwarding %s -> %s", id, addr, remoteAddr)
	return nil
}

func acceptLoop(ctx context.Context, id int, ln net.Listener, remoteAddr string) {
	defer func() {
		mu.Lock()
		delete(tunnels, id)
		mu.Unlock()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				log.Printf("tunnel %d: accept error: %v", id, err)
				return
			}
		}
		go handleConn(ctx, id, conn, remoteAddr)
	}
}

func handleConn(ctx context.Context, id int, local net.Conn, remoteAddr string) {
	defer local.Close()

	dialer := net.Dialer{Timeout: 5 * time.Second}
	remote, err := dialer.DialContext(ctx, "tcp", remoteAddr)
	if err != nil {
		log.Printf("tunnel %d: failed to dial remote %s: %v", id, remoteAddr, err)
		return
	}
	defer remote.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go pipe(&wg, remote, local)
	go pipe(&wg, local, remote)
	wg.Wait()
}

// pipe copies src -> dst until EOF/error, then half-closes dst so the other
// goroutine's copy also unblocks.
func pipe(wg *sync.WaitGroup, dst, src net.Conn) {
	defer wg.Done()
	_, _ = io.Copy(dst, src)
	if tcp, ok := dst.(*net.TCPConn); ok {
		_ = tcp.CloseWrite()
	} else {
		_ = dst.Close()
	}
}

// StopTunnel stops the running tunnel identified by id, closing its listener
// and every connection descended from it.
func StopTunnel(id int) error {
	mu.Lock()
	t, ok := tunnels[id]
	mu.Unlock()
	if !ok {
		return fmt.Errorf("no active tunnel for service %d", id)
	}
	t.cancel()
	return t.listener.Close()
}

// IsRunning reports whether a tunnel for the given service id is currently active.
func IsRunning(id int) bool {
	mu.Lock()
	defer mu.Unlock()
	_, ok := tunnels[id]
	return ok
}
