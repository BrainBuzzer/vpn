package cmd

import (
	"github.com/BrainBuzzer/vpn/config"
	redisClient "github.com/BrainBuzzer/vpn/internal/redis"
	"github.com/BrainBuzzer/vpn/internal/server"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Server for the VPN",
	Long:  `This is the server for the VPN application`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.NewConfig()
		if err != nil {
			panic(err)
		}

		redis := redisClient.NewRedisClient(config.RedisConfig)
		server := server.NewServer(redis)
		server.Start()
	},
}
