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
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/polar-bear-cu/sgt-scheduler/clients"
	"github.com/polar-bear-cu/sgt-scheduler/config"
	"github.com/polar-bear-cu/sgt-scheduler/publisher"
	"github.com/polar-bear-cu/sgt-scheduler/routes"
	"github.com/polar-bear-cu/sgt-scheduler/scheduler"
	"github.com/polar-bear-cu/sgt-scheduler/usecases"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	subConn, err := grpc.NewClient(cfg.SubscriptionAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = subConn.Close() }()

	userConn, err := grpc.NewClient(cfg.UserServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = userConn.Close() }()

	rmqConn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = rmqConn.Close() }()

	pub, err := publisher.NewRabbitMQPublisher(rmqConn)
	if err != nil {
		log.Fatal(err)
	}

	subs := clients.NewSubscriptionClient(subConn)
	users := clients.NewUserClient(userConn)
	uc := usecases.NewBillingReminder(subs, users, pub)

	sched := scheduler.New()
	if err := sched.AddJob(cfg.CronSpec, func(ctx context.Context) {
		sent, err := uc.Run(ctx, cfg.WithinHours)
		if err != nil {
			log.Printf("billing reminder run: sent=%d err=%v", sent, err)
			return
		}
		log.Printf("billing reminder run: sent=%d", sent)
	}); err != nil {
		log.Fatal(err)
	}
	sched.Start()

	r := gin.Default()
	routes.Register(r)
	srv := &http.Server{Addr: ":" + cfg.Port, Handler: r}

	go func() {
		log.Println("listening :" + cfg.Port)
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
