package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ride4Low/contracts/env"
	"github.com/ride4Low/contracts/events"
	"github.com/ride4Low/contracts/pkg/otel"
	"github.com/ride4Low/contracts/pkg/rabbitmq"
	"github.com/ride4Low/payment-service/internal/application"
	"github.com/ride4Low/payment-service/internal/infrastructure/messaging"
	"github.com/ride4Low/payment-service/internal/infrastructure/payment/stripe"
	"github.com/ride4Low/payment-service/internal/interface/consumer"
)

var (
	rabbitMQURI      = env.GetString("RABBITMQ_URI", "amqp://guest:guest@localhost:5672/")
	stripeSecretKey  = env.GetString("STRIPE_SECRET_KEY", "")
	stripeSuccessURL = env.GetString("STRIPE_SUCCESS_URL", "")
	stripeCancelURL  = env.GetString("STRIPE_CANCEL_URL", "")
	jaegerEndpoint   = env.GetString("JAEGER_ENDPOINT", "jaeger:4317")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	otelCfg := otel.DefaultConfig("payment-service")
	otelCfg.JaegerEndpoint = jaegerEndpoint

	otelProvider, err := otel.Setup(ctx, otelCfg)
	if err != nil {
		log.Fatalf("failed to setup otel: %v", err)
	}
	defer func() {
		if err := otelProvider.Shutdown(context.Background()); err != nil {
			log.Printf("failed to shutdown otel: %v", err)
		}
	}()

	rmq, err := rabbitmq.NewRabbitMQ(rabbitMQURI)
	if err != nil {
		log.Fatal("failed to connect to RabbitMQ: ", err)
	}
	defer rmq.Close()

	// Infrastructure layer: Create Stripe payment provider (adapter)
	stripeProvider := stripe.NewProvider(stripe.PaymentConfig{
		StripeSecretKey: stripeSecretKey,
		SuccessURL:      stripeSuccessURL,
		CancelURL:       stripeCancelURL,
	})

	// Infrastructure layer: Create RabbitMQ event publisher (adapter)
	rmqPublisher := rabbitmq.NewPublisher(rmq)
	eventPublisher := messaging.NewRabbitMQPublisher(rmqPublisher)

	// Application layer: Create payment service with provider and publisher
	paymentSvc := application.NewPaymentService(stripeProvider, eventPublisher)

	// Interface layer: Create event handler with payment service
	eventHandler := consumer.NewEventHandler(paymentSvc)

	// Start consuming messages
	msgConsumer := rabbitmq.NewConsumer(rmq, eventHandler)
	go msgConsumer.Consume(ctx, events.PaymentTripResponseQueue)

	<-ctx.Done()
	log.Println("shutting down consumer")
}
