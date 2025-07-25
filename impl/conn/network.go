package conn

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/Relixik/minecraft-server/apis/buff"
	"github.com/Relixik/minecraft-server/apis/logs"
	"github.com/Relixik/minecraft-server/impl/base"
	"github.com/Relixik/minecraft-server/impl/data/system"
)

type network struct {
	host string
	port int

	logger  *logs.Logging
	packets base.Packets

	join chan base.PlayerAndConnection
	quit chan base.PlayerAndConnection

	report chan system.Message
	
	// Memory leak prevention
	ctx        context.Context
	cancel     context.CancelFunc
	listener   *net.TCPListener
	connections sync.Map  // Active connections tracking
	wg         sync.WaitGroup
}

func NewNetwork(host string, port int, packet base.Packets, report chan system.Message, join chan base.PlayerAndConnection, quit chan base.PlayerAndConnection) base.Network {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &network{
		host: host,
		port: port,

		join: join,
		quit: quit,

		report: report,

		logger:  logs.NewLogging("network", logs.EveryLevel...),
		packets: packet,
		
		ctx:    ctx,
		cancel: cancel,
	}
}

func (n *network) Load() {
	if err := n.startListening(); err != nil {
		n.report <- system.Make(system.FAIL, err)
		return
	}
}

func (n *network) Kill() {
	n.logger.Info("Shutting down network...")
	
	// Cancel context to signal all goroutines to stop
	n.cancel()
	
	// Close the listener to stop accepting new connections
	if n.listener != nil {
		n.listener.Close()
	}
	
	// Close all active connections
	n.connections.Range(func(key, value interface{}) bool {
		if conn, ok := value.(base.Connection); ok {
			conn.Stop()
		}
		return true
	})
	
	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		n.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		n.logger.Info("All network goroutines stopped gracefully")
	case <-time.After(5 * time.Second):
		n.logger.Warn("Timeout waiting for network goroutines to stop")
	}
}

func (n *network) startListening() error {
	ser, err := net.ResolveTCPAddr("tcp", n.host+":"+strconv.Itoa(n.port))
	if err != nil {
		return fmt.Errorf("address resolution failed [%v]", err)
	}

	tcp, err := net.ListenTCP("tcp", ser)
	if err != nil {
		return fmt.Errorf("failed to bind [%v]", err)
	}
	
	n.listener = tcp
	n.logger.InfoF("listening on %s:%d", n.host, n.port)

	// Start the accept loop in a separate goroutine with proper cleanup
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		defer tcp.Close()
		
		for {
			select {
			case <-n.ctx.Done():
				n.logger.Info("Accept loop stopping due to context cancellation")
				return
			default:
				// Set a short timeout for Accept to allow context checking
				tcp.SetDeadline(time.Now().Add(1 * time.Second))
				
				con, err := tcp.AcceptTCP()
				if err != nil {
					// Check if it's a timeout (normal for shutdown)
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						continue // Timeout is normal, check context again
					}
					// Real error or shutdown
					select {
					case <-n.ctx.Done():
						return // Shutting down
					default:
						n.report <- system.Make(system.FAIL, err)
						return
					}
				}

				_ = con.SetNoDelay(true)
				_ = con.SetKeepAlive(true)
				
				// Create connection wrapper
				conn := NewConnection(con)
				
				// Track the connection
				n.connections.Store(conn, conn)

				// Handle connection in separate goroutine
				n.wg.Add(1)
				go func(conn base.Connection) {
					defer n.wg.Done()
					defer n.connections.Delete(conn)
					handleConnect(n, conn)
				}(conn)
			}
		}
	}()

	return nil
}

func handleConnect(network *network, conn base.Connection) {
	network.logger.DataF("New Connection from &6%v", conn.Address())

	var inf []byte

	for {
		select {
		case <-network.ctx.Done():
			network.logger.DataF("Connection handler stopping for %v due to context cancellation", conn.Address())
			conn.Stop()
			return
		default:
			inf = make([]byte, 1024)
			sze, err := conn.Pull(inf)

			if err != nil && err.Error() == "EOF" {
				network.quit <- base.PlayerAndConnection{
					Player:     nil,
					Connection: conn,
				}
				return
			}

			if err != nil || sze == 0 {
				_ = conn.Stop()

				network.quit <- base.PlayerAndConnection{
					Player:     nil,
					Connection: conn,
				}
				return
			}

			buf := NewBufferWith(conn.Decrypt(inf[:sze]))

			// decompression
			// decryption

			if buf.UAS()[0] == 0xFE { // LEGACY PING
				continue
			}

			packetLen := buf.PullVrI()

			bufI := NewBufferWith(buf.UAS()[buf.InI() : buf.InI()+packetLen])
			bufO := NewBuffer()

			handleReceive(network, conn, bufI, bufO)

			if bufO.Len() > 1 {
				temp := NewBuffer()
				temp.PushVrI(bufO.Len())

				comp := NewBuffer()
				comp.PushUAS(conn.Deflate(bufO.UAS()), false)

				temp.PushUAS(comp.UAS(), false)

				_, err := conn.Push(conn.Encrypt(temp.UAS()))

				if err != nil {
					network.logger.Fail("Failed to push client bound packet: %v", err)
				}
			}
		}
	}
}

func handleReceive(network *network, conn base.Connection, bufI buff.Buffer, bufO buff.Buffer) {
	uuid := bufI.PullVrI()

	packetI := network.packets.GetPacketI(uuid, conn.GetState())
	if packetI == nil {
		network.logger.DataF("unable to decode %v packet with uuid: %d", conn.GetState(), uuid)
		return
	}

	network.logger.DataF("GET packet: %d | %v | %v", packetI.UUID(), reflect.TypeOf(packetI), conn.GetState())

	// populate incoming packet
	packetI.Pull(bufI, conn)

	network.packets.PubAs(packetI)
	network.packets.PubAs(packetI, conn)
}
