// grush.go
package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/crbrox/minigrush"
	"github.com/crbrox/minigrush/config"
	"github.com/crbrox/store"
	"github.com/crbrox/store/dir"
	"github.com/crbrox/store/redis"
)

func main() {
	var (
		reqStore, respStore store.Interface
	)
	log.Println("Hello World!")
	cfg, err := config.ReadConfig("grush.ini")
	if err != nil {
		log.Fatalln("-", err)
	}

	reqChan := make(chan *minigrush.Petition, cfg.QueueSize)
	switch cfg.StoreType {
	case "dir":
		reqStore = &dir.Store{cfg.Dir.RequestPath}
		respStore = &dir.Store{cfg.Dir.ResponsePath}
	case "redis":
		optionsStore := redis.StoreOptions{
			Prefix: "grush-req:", MaxIdle: cfg.Redis.MaxIdle, MaxActive: cfg.Redis.MaxActive, Server: cfg.Redis.Server, IdleTimeout: cfg.Redis.IdleTimeout,
		}
		reqStore = redis.NewStore(optionsStore)
		optionsStore.Prefix = "grush-resp:"
		respStore = redis.NewStore(optionsStore)
	default:
		log.Fatalf("- Unsupported store type %q\n", cfg.StoreType)
	}

	l := &minigrush.Listener{SendTo: reqChan, PetitionStore: reqStore}
	c := &minigrush.Consumer{GetFrom: reqChan, PetitionStore: reqStore, ReplyStore: respStore}
	r := &minigrush.Replyer{ReplyStore: respStore}
	rcvr := &minigrush.Recoverer{SendTo: reqChan, PetitionStore: reqStore}

	endConsumers := c.Start(cfg.Consumers)
	if err := rcvr.Recover(); err != nil {
		log.Fatalln("-", err)
	}
	http.Handle("/", l)
	http.Handle("/responses/", http.StripPrefix("/responses/", r))
	go func() {
		log.Fatalln("-", http.ListenAndServe(":"+cfg.Port, nil))
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
