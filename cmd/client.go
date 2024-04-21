package cmd

import (
	"fmt"
	"log/slog"

	"github.com/BrainBuzzer/vpn/config"
	"github.com/BrainBuzzer/vpn/internal/client"
	redisClient "github.com/BrainBuzzer/vpn/internal/redis"
	"github.com/spf13/cobra"
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Client for the VPN",
	Long:  `This is the client for the VPN application`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.NewConfig()
		if err != nil {
			slog.Error(fmt.Errorf("Failed to read config: %v", err).Error())
			return
		}

		redisClient := redisClient.NewRedisClient(config.RedisConfig)
		client := client.NewClient(redisClient)
		client.Start()
	},
}
