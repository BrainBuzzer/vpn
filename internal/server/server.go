package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os/exec"

	redisClient "github.com/BrainBuzzer/vpn/internal/redis"
	"github.com/songgao/water"
)

type Server struct {
	redisClient redisClient.RedisClientInterface
}

type ServerInterface interface {
	Start()
}

func NewServer(redisClient redisClient.RedisClientInterface) ServerInterface {
	return &Server{
		redisClient: redisClient,
	}
}

func (s *Server) Start() {
	ctx := context.Background()
	config := water.Config{
		DeviceType: water.TUN,
	}

	ifce, err := water.New(config)
	if err != nil {
		slog.Error(fmt.Errorf("Failed to create TUN device: %v", err).Error())
		return
	}

	name := ifce.Name()
	slog.Info(fmt.Sprintf("Interface Name: %s", name))

	if err := exec.Command("ifconfig", name, "10.11.12.1", "netmask", "255.255.255.0").Run(); err != nil {
		slog.Error(fmt.Errorf("Failed to configure TUN device: %v", err).Error())
		return
	}

	slog.Info("Interface configured")

	// start a vpn server on port 33512
	ln, err := net.Listen("tcp", ":33512")
	if err != nil {
		slog.Error(fmt.Errorf("Failed to listen on port 33512: %v", err).Error())
		return
	}

	defer ln.Close()

	// get eth0 interface
	eth0, err := net.InterfaceByName("eth0")
	if err != nil {
		slog.Error(fmt.Errorf("Failed to get eth0 interface: %v", err).Error())
		return
	}

	// get eth0 interface address
	addrs, err := eth0.Addrs()
	if err != nil {
		slog.Error(fmt.Errorf("Failed to get eth0 interface address: %v", err).Error())
		return
	}

	// get the first address
	addr := addrs[0].String()

	// get the ip address
	ip, _, err := net.ParseCIDR(addr)

	// store the ip address in redis
	err = s.redisClient.Set(ctx, "server_ip", fmt.Sprintf("%s:33512", ip.String()), 0)

	slog.Info("Server started")

	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Error(fmt.Errorf("Failed to accept connection: %v", err).Error())
		}

		go s.handleIncomingConnection(ifce, conn)
		go s.handleOutgoingConnection(ifce, conn)
	}

}

// All incoming connections are handled by this function
func (s *Server) handleIncomingConnection(tun *water.Interface, conn net.Conn) {
	buf := make([]byte, 1500)
	// the reason for a continuous loop is because we want to keep reading from the connection
	for {
		// Read all the bytes from incoming connection request
		n, err := conn.Read(buf)
		if err != nil {
			// this particular scenario occurs on first net.Dial call from client, so we need to just ignore it
			if err.Error() == "EOF" {
				slog.Info("Connection closed")
			} else {
				slog.Error(fmt.Errorf("Failed to read from connection: %v", err).Error())
				return
			}
		}

		// before sending the request back to tun device, we will insert a few headers to the request
		// this is to ensure that the request is coming from the server
		// and not from the client
		if _, err := tun.Write([]byte("HTTP/1.1 101 Switching Protocols\r\n" +
			"Connection: Upgrade\r\n" +
			"Upgrade: VPNServer\r\n" +
			"\r\n")); err != nil {
			slog.Error(fmt.Errorf("Failed to write custom headers to TUN device: %v", err).Error())
			return
		}

		// This will receive and write any bytes from the connection to the TUN device
		if n > 0 {
			if _, err := tun.Write(buf[:n]); err != nil {
				slog.Error(fmt.Errorf("Failed to write to TUN device: %v", err).Error())
				return
			}
		}
	}
}

// All outgoing connections are handled by this function
func (s *Server) handleOutgoingConnection(tun *water.Interface, conn net.Conn) {
	buf := make([]byte, 1500)

	// the reason for a continuous loop is because we want to keep reading from the connection
	// and write to the TUN device
	for {
		// Read all the bytes from the TUN device
		n, err := tun.Read(buf)
		if err != nil {
			slog.Error(fmt.Errorf("Failed to read from TUN device: %v", err).Error())
			return
		}

		// print buffer here
		slog.Info(fmt.Sprintf("Sending Buffer: %s", buf[:n]))

		// This will receive and write any bytes from the TUN device to the connection
		if _, err := conn.Write(buf[:n]); err != nil {
			slog.Error(fmt.Errorf("Failed to write to connection: %v", err).Error())
			return
		}
	}
}
