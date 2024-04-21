package client

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os/exec"

	redisClient "github.com/BrainBuzzer/vpn/internal/redis"
	"github.com/songgao/water"
)

type Client struct {
	redisClient redisClient.RedisClientInterface
}

type ClientInterface interface {
	Start()
}

func NewClient(redisClient redisClient.RedisClientInterface) ClientInterface {
	return &Client{
		redisClient: redisClient,
	}
}

func (c *Client) Start() {
	ctx := context.Background()
	// fetch the ip address of server
	ip, err := c.redisClient.Get(ctx, "server_ip")
	if err != nil {
		slog.Error(fmt.Errorf("Cannot get server ip from redis, are you sure there is a server node running?").Error())
		return
	}

	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})

	if err != nil {
		slog.Error(fmt.Errorf("Cannot create TUN device").Error())
		return
	}

	name := ifce.Name()

	if err := exec.Command("ifconfig", name, "10.11.12.2", "netmask", "255.255.255.0").Run(); err != nil {
		slog.Error(fmt.Errorf("Cannot configure TUN device").Error())
		return
	}

	// connect to the server
	conn, err := net.Dial("tcp", ip)
	if err != nil {
		slog.Error(fmt.Errorf("Cannot connect to server: %s", err).Error())
		return
	}

	defer conn.Close()

	slog.Info(fmt.Sprintf("Connected to server: %s", ip))
	slog.Info(fmt.Sprintf("Interface Name: %s", name))
	slog.Info("Interface configured")
	go handleIncomingConnection(ifce, conn)
	go handleOutgoingConnection(ifce, conn)

	select {}
}

// This handles all the incoming connections from the server
// Logic for this is mostly same as the server, but reversed
// to emulate a client
func handleIncomingConnection(ifce *water.Interface, conn net.Conn) {
	buf := make([]byte, 1500)
	for {
		n, err := ifce.Read(buf)
		if err != nil {
			slog.Error(fmt.Errorf("Cannot read from server: %s", err).Error())
			return
		}
		_, err = conn.Write(buf[:n])
		if err != nil {
			slog.Error(fmt.Errorf("Cannot write to server: %s", err).Error())
			return
		}
	}
}

// This handles all the outgoing connections to the server
func handleOutgoingConnection(ifce *water.Interface, conn net.Conn) {
	buf := make([]byte, 1500)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			slog.Error(fmt.Errorf("Cannot read from server: %s", err).Error())
			return
		}
		_, err = ifce.Write(buf[:n])
		if err != nil {
			slog.Error(fmt.Errorf("Cannot write to server: %s", err).Error())
			return
		}
	}
}
