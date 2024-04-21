package cmd

import (
	"fmt"
	"os"

	"github.com/BrainBuzzer/vpn/config"
	redisClient "github.com/BrainBuzzer/vpn/internal/redis"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vpn",
	Short: "VPN application",
	Long:  `This is the Simple VPN implementation in Go`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.NewConfig()
		if err != nil {
			panic(err)
		}
		redis := redisClient.NewRedisClient(config.RedisConfig)
		err = redis.HealthCheck()
		if err != nil {
			panic(err)
		}

		fmt.Printf("VPN application is able to connect with redis\n")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(clientCmd)
}
