/*
Package config provides types and a function for getting grush configuration.

Configuration is read the file grush.ini. An possible content of such file could be
	[default]
	port=8080
	queueSize=100000
	consumers=1000
	storeType=redis

	[dir]
	responsePath=RESP
	requestPath=REQ

	[redis]
	maxIdle=10000
	maxActive=10000
	server=localhost:6379
	idleTimeout=1h

The storeType choose an store and implies that there is a section in the configuration
file for the type chosen
*/
package config

import (
	"fmt"
	"time"

	"github.com/msbranco/goconfig"
)

//General configuration data
type Config struct {
	//Port to listen to
	Port string
	//Maximun number of request enqueued, waiting for being processed
	QueueSize int
	//Maximun number of concurrent requests being processed
	Consumers int
	//Type of store for Petitions and Replies. Currently redis or dir
	StoreType string
	//Config for dir store if applies
	Dir *DirConfig
	//Config for redis store if applies
	Redis *RedisConfig
}

//Specific configuration data when redis store is selected.
type RedisConfig struct {
	MaxIdle     int
	MaxActive   int
	Server      string
	IdleTimeout time.Duration
}

//Specific configuration data when redis dir is selected.
type DirConfig struct {
	ResponsePath string
	RequestPath  string
}

//ReadConfig reads configuration from file with name filename. The file must include a section for the selected store type.
//The store configuration will be loaded into the returned Config object.
func ReadConfig(filename string) (*Config, error) {
	var (
		redisC *RedisConfig
		dirC   *DirConfig
	)
	cfg, err := goconfig.ReadConfigFile(filename)
	if err != nil {
		return nil, err
	}
	port, err := cfg.GetString("default", "port")
	if err != nil {
		return nil, fmt.Errorf("readConfig: %v", err)
	}
	//TODO: check if it is within port range
	queueSize, err := cfg.GetInt64("default", "queueSize")
	if err != nil {
		return nil, fmt.Errorf("readConfig: %v", err)
	}
	consumers, err := cfg.GetInt64("default", "consumers")
	if err != nil {
		return nil, fmt.Errorf("readConfig: %v", err)
	}
	storeType, err := cfg.GetString("default", "storeType")
	if err != nil {
		return nil, fmt.Errorf("readConfig: %v", err)
	}
	switch storeType {
	case "redis":
		redisC, err = newRedisConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("readConfig: %v", err)
		}
	case "dir":
		dirC, err = newDirConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("readConfig: %v", err)
		}
	default:
		return nil, fmt.Errorf("readConfig: Unsupported store type")
	}

	return &Config{
		Port:      port,
		QueueSize: int(queueSize), // Check downsizing, positive, blah, blah
		Consumers: int(consumers), // Check downsizing
		StoreType: storeType,
		Dir:       dirC,
		Redis:     redisC,
	}, nil
}

//newRedisConfig returns the RedisConfig data in an already poen configuration file.
//It must be present or an error will be returned
func newRedisConfig(cfg *goconfig.ConfigFile) (*RedisConfig, error) {
	maxIdle, err := cfg.GetInt64("redis", "maxIdle")
	if err != nil {
		return nil, fmt.Errorf("readConfig: %v", err)
	}
	maxActive, err := cfg.GetInt64("redis", "maxActive")
	if err != nil {
		return nil, fmt.Errorf("readConfig: %v", err)
	}
	server, err := cfg.GetString("redis", "server")
	if err != nil {
		return nil, fmt.Errorf("readConfig: %v", err)
	}
	idleStr, err := cfg.GetString("redis", "idleTimeout")
	if err != nil {
		return nil, fmt.Errorf("readConfig: %v", err)
	}
	idleTimeout, err := time.ParseDuration(idleStr)
	if err != nil {
		return nil, fmt.Errorf("readConfig: %v", err)
	}
	return &RedisConfig{
		MaxIdle:     int(maxIdle),   //check
		MaxActive:   int(maxActive), //check
		Server:      server,
		IdleTimeout: idleTimeout}, nil
}

//newDirConfig returns the DirConfig data in an already poen configuration file.
//It must be present or an error will be returned
func newDirConfig(cfg *goconfig.ConfigFile) (*DirConfig, error) {
	responsePath, err := cfg.GetString("dir", "responsePath")
	if err != nil {
		return nil, fmt.Errorf("readConfig: %v", err)
	}
	requestPath, err := cfg.GetString("dir", "requestPath")
	if err != nil {
		return nil, fmt.Errorf("readConfig: %v", err)
	}
	return &DirConfig{
		ResponsePath: responsePath,
		RequestPath:  requestPath,
	}, nil
}
