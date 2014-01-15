// grush.go
package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/crbrox/minigrush"
	"github.com/crbrox/store"
	"github.com/crbrox/store/dir"
	"github.com/crbrox/store/redis"
)

func main() {
	var (
		reqStore, respStore store.Interface
	)
	log.Println("Hello World!")
	config, err := grass.ReadConfig("grush.ini")
	if err != nil {
		log.Fatalln("-", err)
	}

	reqChan := make(chan *grass.Petition, config.QueueSize)
	switch config.StoreType {
	case "dir":
		reqStore = &dir.Store{config.Dir.RequestPath}
		respStore = &dir.Store{config.Dir.ResponsePath}
	case "redis":
		optionsStore := redis.StoreOptions{
			Prefix: "grush-req:", MaxIdle: config.Redis.MaxIdle, MaxActive: config.Redis.MaxActive, Server: config.Redis.Server, IdleTimeout: config.Redis.IdleTimeout,
		}
		reqStore = redis.NewStore(optionsStore)
		optionsStore.Prefix = "grush-resp:"
		respStore = redis.NewStore(optionsStore)
	default:
		log.Fatalf("- Unsupported store type %q\n", config.StoreType)
	}

	l := &grass.Listener{SendTo: reqChan, PetitionStore: reqStore}
	c := &grass.Consumer{GetFrom: reqChan, PetitionStore: reqStore, ReplyStore: respStore}
	r := &grass.Replyer{ReplyStore: respStore}
	rcvr := &grass.Recoverer{SendTo: reqChan, PetitionStore: reqStore}

	endConsumers := c.Start(config.Consumers)
	if err := rcvr.Recover(); err != nil {
		log.Fatalln("-", err)
	}
	http.Handle("/", l)
	http.Handle("/responses/", http.StripPrefix("/responses/", r))
	go func() {
		log.Fatalln("-", http.ListenAndServe(":"+config.Port, nil))
	}()
	onEnd(func() {
		log.Println("\nShutting down grass ...")
		l.Stop()
		c.Stop()
	})
	<-endConsumers
	log.Println("Bye World!")
}

func onEnd(f func()) {
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		f()
	}()
}
