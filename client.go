package udp2tcp

import (
	"net"
	"sync"
	"time"
	"log"
	"errors"
)

const UDP_BUFFER = 64 * 1024

var tcpClientMapper sync.Map

type tcpClientTunnel struct {
	key          string
	conn         net.Conn
	sendBuffer   []byte
	recvBuffer   []byte
	sendBuffChan chan []byte
}

func (tunnel *tcpClientTunnel) destroy(err error) {
	log.Println("destroy tcp tunnel by err:", err)
	tcpClientMapper.Delete(tunnel.key)
	if tunnel.conn != nil {
		tunnel.conn.SetDeadline(time.Now())
		tunnel.conn.Close()
	}
	//if tcpTunnel.sendBuffChan != nil {
	//	close(tcpTunnel.sendBuffChan)
	//}
}

func StartClient(listen string, server string, timeOut time.Duration) (err error) {
	relayClient, err := net.ListenPacket("udp", listen)
	if err != nil {
		return
	}
	buff := make([]byte, UDP_BUFFER)
	for {
		n, raddr, err := relayClient.ReadFrom(buff)
		if err != nil {
			continue
		}
		key := raddr.String()

		var tcpTunnel *tcpClientTunnel
		if v, ok := tcpClientMapper.Load(key); ok {
			tcpTunnel = v.(*tcpClientTunnel)
		} else {
			tcpTunnel = &tcpClientTunnel{
				key : key,
				sendBuffer: make([]byte, UDP_BUFFER),
				recvBuffer: make([]byte, UDP_BUFFER),
				sendBuffChan:make(chan []byte, 1),
			}
			tcpClientMapper.Store(key, tcpTunnel)
			go func(key string) {
				conn, err := net.DialTimeout("tcp", server, time.Second * 3)
				if err != nil {
					tcpTunnel.destroy(err)
					return
				}
				tcpTunnel.conn = conn
				go func() {
					for {
						buff, ok := <-tcpTunnel.sendBuffChan
						if !ok {
							tcpTunnel.destroy(errors.New("failed to read buffer channel"))
							return
						}
						conn.SetWriteDeadline(time.Now().Add(timeOut))
						if _, err := conn.Write(buff); err != nil {
							tcpTunnel.destroy(err)
							return
						}
					}
				}()
				go func() {
					for {
						conn.SetReadDeadline(time.Now().Add(timeOut))
						n, err := conn.Read(tcpTunnel.recvBuffer)
						if err != nil {
							tcpTunnel.destroy(err)
							return
						}
						cAddr, _ := net.ResolveUDPAddr("udp", tcpTunnel.key)
						relayClient.SetWriteDeadline(time.Now().Add(timeOut))
						relayClient.WriteTo(tcpTunnel.recvBuffer[:n], cAddr)
					}
				}()
			}(key)
		}
		//go func(tcpBuffLen int) {
		if len(tcpTunnel.sendBuffChan) == 0 {
			copy(tcpTunnel.sendBuffer, buff[:n])
			tcpTunnel.sendBuffChan <- tcpTunnel.sendBuffer[:n]
		} else {
			//log.Println("sendBuffChan is full", len(tcpTunnel.sendBuffChan))
		}
		//}(n)
	}
}
