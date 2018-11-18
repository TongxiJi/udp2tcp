package udp2tcp

import (
	"net"
	"time"
	"log"
	"sync"
)

const TCP_BUFFER = 64 * 1024

var tcpServerMapper sync.Map

type tcpServerTunnel struct {
	key        string
	clientConn net.Conn
	serverConn net.PacketConn
	sendBuffer []byte
	recvBuffer []byte
	//sendBuffChan chan []byte
}

func (tunnel *tcpServerTunnel) destroy(err error) {
	log.Println("destroy tcp tunnel by err:", err)
	tcpServerMapper.Delete(tunnel.key)
	if tunnel.clientConn != nil {
		tunnel.clientConn.SetDeadline(time.Now())
		tunnel.clientConn.Close()
	}
	if tunnel.serverConn != nil {
		tunnel.serverConn.SetDeadline(time.Now())
		tunnel.serverConn.Close()
	}
	//if tcpTunnel.sendBuffChan != nil {
	//	close(tcpTunnel.sendBuffChan)
	//}
}

func StartServer(listen string, appServer string, timeOut time.Duration) (err error) {
	relayServer, err := net.Listen("tcp", listen)
	if err != nil {
		return
	}

	appUdpServer, _ := net.ResolveUDPAddr("udp", appServer)

	for {
		relayClientConn, err := relayServer.Accept()
		if err != nil {
			continue
		}
		go func() {
			key := relayClientConn.RemoteAddr().String()
			relayServerConn, _ := net.DialUDP("udp", nil, appUdpServer)
			tcpServerTunnel := &tcpServerTunnel{
				key: key,
				clientConn:relayClientConn,
				serverConn:relayServerConn,
				sendBuffer:make([]byte, TCP_BUFFER),
				recvBuffer:make([]byte, TCP_BUFFER),
			}
			tcpServerMapper.Store(key, tcpServerTunnel)
			go func() {
				for {
					relayClientConn.SetReadDeadline(time.Now().Add(timeOut))
					n, err := relayClientConn.Read(tcpServerTunnel.sendBuffer)
					if err != nil {
						tcpServerTunnel.destroy(err)
						return
					}
					//log.Println(tcpServerTunnel.sendBuffer[:n])
					relayServerConn.SetWriteDeadline(time.Now().Add(timeOut))
					_, err = relayServerConn.Write(tcpServerTunnel.sendBuffer[:n])
					if err != nil {
						tcpServerTunnel.destroy(err)
					}
				}
			}()

			go func() {
				for {
					relayServerConn.SetReadDeadline(time.Now().Add(timeOut))
					n, err := relayServerConn.Read(tcpServerTunnel.recvBuffer)
					if err != nil {
						tcpServerTunnel.destroy(err)
						return
					}
					relayClientConn.SetWriteDeadline(time.Now().Add(timeOut))
					_, err = relayClientConn.Write(tcpServerTunnel.recvBuffer[:n])
					if err != nil {
						tcpServerTunnel.destroy(err)
						return
					}
				}
			}()
		}()
	}
}
