package config

import (
	"fmt"
	"net"

	"github.com/mohitkumar/orchy/analytics"
)

type StorageType string

type QueueType string

const STORAGE_TYPE_REDIS StorageType = "redis"
const STORAGE_TYPE_INMEM StorageType = "memory"
const STORAGE_TYPE_DYNAMO StorageType = "dynamo"

const QUEUE_TYPE_REDIS QueueType = "redis"
const QUEUE_TYPE_SQS QueueType = "sqs"

type EncoderDecoderType string

const JSON_ENCODER_DECODER EncoderDecoderType = "JSON"
const PROTO_ENCODER_DECODER EncoderDecoderType = "PROTO"

type Config struct {
	RedisConfig        RedisStorageConfig
	InMemoryConfig     InmemStorageConfig
	RedisQueueConfig   RedisQueueConfig
	InMemQueueConfig   InMemQueueConfig
	HttpPort           int
	GrpcPort           int
	StorageType        StorageType
	QueueType          QueueType
	EncoderDecoderType EncoderDecoderType
	ClusterConfig      ClusterConfig
	BatchSize          int
	AnalyticsConfig    analytics.DataCollectorConfig
}

type ClusterConfig struct {
	NodeName       string
	BindAddr       string
	Tags           map[string]string
	StartJoinAddrs []string
	PartitionCount int
}

func (c Config) RPCAddr() (string, error) {
	host, _, err := net.SplitHostPort(c.ClusterConfig.BindAddr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", host, c.GrpcPort), nil
}

type RedisStorageConfig struct {
	Addrs     []string
	Namespace string
}

type RedisQueueConfig struct {
	Addrs     []string
	Namespace string
}

type InmemStorageConfig struct {
}
type InMemQueueConfig struct {
}
