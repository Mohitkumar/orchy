package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mohitkumar/orchy/server/agent"
	"github.com/mohitkumar/orchy/server/analytics"
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
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Flags().String("config-file", "", "Path to config file.")
	cmd.Flags().String("redis-addr", "localhost:6379", "comma separated list of redis host:port")
	cmd.Flags().String("namespace", "orchy", "namespace used in storage")
	cmd.Flags().Int("http-port", 8080, "http port for rest endpoints")
	cmd.Flags().Int("grpc-port", 8099, "grpc port for worker connection")
	cmd.Flags().String("storage-impl", "redis", "implementation of underline storage")
	cmd.Flags().String("queue-impl", "redis", "implementation of underline queue ")
	cmd.Flags().String("data-serializer", "JSON", "encoder decoder used to serialzie data")
	cmd.Flags().Int("partitions", 7, "number of partition")
	cmd.Flags().Int("batch-size", 100, "batch size for internal queue")
	cmd.Flags().String("bind-addr", "127.0.0.1:8400", "address for cluster events")
	cmd.Flags().StringSlice("cluster-address", nil, "cluster address to join.")
	cmd.Flags().String("node-name", hostname, "name of the node")
	cmd.Flags().String("anlytics-output", "LOG_FILE", "output source where anlytics will be pushed")
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
	c.cfg.EncoderDecoderType = config.EncoderDecoderType(viper.GetString("data-serializer"))
	c.cfg.BatchSize = viper.GetInt("batch-size")
	c.cfg.ClusterConfig = config.ClusterConfig{
		NodeName:       viper.GetString("node-name"),
		BindAddr:       viper.GetString("bind-addr"),
		StartJoinAddrs: viper.GetStringSlice("cluster-address"),
		PartitionCount: viper.GetInt("partitions"),
	}
	rpcAddr, err := c.cfg.RPCAddr()
	if err != nil {
		return err
	}
	c.cfg.ClusterConfig.Tags = map[string]string{
		"rpc_addr": rpcAddr,
	}
	analyticsOut := viper.GetString("anlytics-output")
	switch analyticsOut {
	case "LOG_FILE":
		c.cfg.AnalyticsConfig.CollectorType = analytics.LOG_FILE_DATA_COLLECTOR
		c.cfg.AnalyticsConfig.FileName = "analytics.log"
	}

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
