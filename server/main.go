package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mohitkumar/orchy/server/agent"
	"github.com/mohitkumar/orchy/server/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type cfg struct {
	config.Config
}
type cli struct {
	cfg cfg
}

func setupFlags(cmd *cobra.Command) error {
	cmd.Flags().String("config-file", "", "Path to config file.")
	cmd.Flags().String("redis-addr", "localhost:6379", "comma separated list of redis host:port")
	cmd.Flags().String("namespace", "orchy", "namespace used in storage")
	cmd.Flags().Int("http-port", 8080, "htt port for rest endpoints")
	cmd.Flags().Int("grpc-port", 8099, "grpc port for worker connection")
	cmd.Flags().String("storage-impl", "redis", "implementation of underline storage")
	cmd.Flags().String("queue-impl", "redis", "implementation of underline queue ")
	cmd.Flags().String("encoder-decoder", "JSON", "encoder decoder used to serialzie data")
	cmd.Flags().Int("executor-capacity", 512, "action executor capacity")
	return viper.BindPFlags(cmd.Flags())
}

func (c *cli) setupConfig(cmd *cobra.Command, args []string) error {
	var err error

	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}
	viper.SetConfigFile(configFile)

	if err = viper.ReadInConfig(); err != nil {
		// it's ok if config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	c.cfg.RedisConfig.Addrs = strings.Split(viper.GetString("redis-addr"), ",")
	c.cfg.RedisConfig.Namespace = viper.GetString("namespace")
	c.cfg.HttpPort = viper.GetInt("http-port")
	c.cfg.GrpcPort = viper.GetInt("grpc-port")
	c.cfg.StorageType = config.StorageType(viper.GetString("storage-impl"))
	c.cfg.QueueType = config.QueueType(viper.GetString("queue-impl"))
	c.cfg.EncoderDecoderType = config.EncoderDecoderType(viper.GetString("encoder-decoder"))
	c.cfg.ActionExecutorCapacity = viper.GetInt("executor-capacity")
	return nil
}

func (c *cli) run(cmd *cobra.Command, args []string) error {
	var err error
	agent, err := agent.New(c.cfg.Config)
	if err != nil {
		panic(err)
	}
	err = agent.Start()
	if err != nil {
		panic(err)
	}
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	<-sigc
	return agent.Shutdown()
}

func main() {
	cli := &cli{}

	cmd := &cobra.Command{
		Use:     "orchy",
		PreRunE: cli.setupConfig,
		RunE:    cli.run,
	}

	if err := setupFlags(cmd); err != nil {
		log.Fatal(err)
	}

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
