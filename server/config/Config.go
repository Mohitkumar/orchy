package config

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
