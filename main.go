package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/polar-bear-cu/sgt-scheduler/routes"
	"github.com/polar-bear-cu/sgt-scheduler/scheduler"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	sched := scheduler.New()
	// TODO: replace with the real billing-reminder job once clients and publisher land
	if err := sched.AddJob("@every 1h", func(ctx context.Context) {
		log.Println("tick")
	}); err != nil {
		log.Fatal(err)
	}
	sched.Start()

	r := gin.Default()
	routes.Register(r)
	srv := &http.Server{Addr: ":8090", Handler: r}

	go func() {
		log.Println("listening :8090")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	stop()
	log.Println("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown: %v", err)
	}
	<-sched.Stop().Done()
}
