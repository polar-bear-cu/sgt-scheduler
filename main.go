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

	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		log.Fatal(err)
	}

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

	rmqClosed := rmqConn.NotifyClose(make(chan *amqp.Error, 1))
	go func() {
		if err, ok := <-rmqClosed; ok {
			log.Fatalf("rabbitmq connection lost: %v", err)
		}
	}()

	pub, err := publisher.NewRabbitMQPublisher(rmqConn)
	if err != nil {
		log.Fatal(err)
	}

	subs := clients.NewSubscriptionClient(subConn)
	users := clients.NewUserClient(userConn)
	reminder := usecases.NewBillingReminder(subs, users, pub, loc, cfg.PublicURL)
	rollover := usecases.NewBillingRollover(subs, loc)

	sched := scheduler.New(loc)
	if err := sched.AddJob(cfg.RolloverCron, time.Minute, func(ctx context.Context) {
		res, err := rollover.Run(ctx, time.Now())
		if err != nil {
			log.Printf("billing rollover failed: date=%s err=%v", res.Date, err)
			return
		}
		log.Printf("billing rollover: date=%s advanced=%d trials_converted=%d", res.Date, res.Advanced, res.TrialsConverted)
	}); err != nil {
		log.Fatal(err)
	}
	if err := sched.AddJob(cfg.ReminderCron, 2*time.Minute, func(ctx context.Context) {
		start := time.Now()
		res, err := reminder.Run(ctx, start)
		log.Printf("reminder check: date=%s due=%d published=%d failed=%d took=%s",
			res.Date, res.Due, res.Published, res.Failed, time.Since(start).Round(time.Millisecond))
		if err != nil {
			log.Printf("reminder check errors: %v", err)
		}
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
