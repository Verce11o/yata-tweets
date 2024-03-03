package app

import (
	"fmt"
	pb "github.com/Verce11o/yata-protos/gen/go/tweets"
	"github.com/Verce11o/yata-tweets/config"
	tweetGrpc "github.com/Verce11o/yata-tweets/internal/handler/grpc"
	"github.com/Verce11o/yata-tweets/internal/lib/logger"
	"github.com/Verce11o/yata-tweets/internal/lib/notification/rabbitmq"
	"github.com/Verce11o/yata-tweets/internal/metrics/trace"
	"github.com/Verce11o/yata-tweets/internal/repository/minio"
	"github.com/Verce11o/yata-tweets/internal/repository/postgres"
	"github.com/Verce11o/yata-tweets/internal/repository/redis"
	"github.com/Verce11o/yata-tweets/internal/service"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func Run() {
	log := logger.NewLogger()
	cfg := config.LoadConfig()

	tracer := trace.InitTracer("yata-tweets")

	// Init repos
	db := postgres.NewPostgres(cfg)
	repo := postgres.NewTweetPostgres(db, tracer.Tracer)

	rdb := redis.NewRedis(cfg)
	redisRepo := redis.NewTweetsRedis(rdb, tracer.Tracer)

	minioClient := minio.NewMinio(cfg)
	minioRepo := minio.NewTweetMinio(minioClient, tracer.Tracer)

	// Init broker

	s := grpc.NewServer(grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor(
		otelgrpc.WithTracerProvider(tracer.Provider),
		otelgrpc.WithPropagators(propagation.TraceContext{}),
	)))

	amqpConn := rabbitmq.NewAmqpConnection(cfg.RabbitMQ)
	tweetPublisher := rabbitmq.NewTweetPublisher(amqpConn, log, tracer.Tracer, cfg.RabbitMQ)
	tweetService := service.NewTweetService(log, tracer.Tracer, tweetPublisher, repo, redisRepo, minioRepo)

	pb.RegisterTweetsServer(s, tweetGrpc.NewTweetGRPC(log, tracer.Tracer, tweetService))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.App.Port))

	if err != nil {
		log.Info("failed to listen: %v", err)
	}

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Infof("error while listen server: %s", err)
		}
	}()

	log.Info(fmt.Sprintf("server listening at %s", lis.Addr().String()))

	defer log.Sync()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.GracefulStop()

	if err := db.Close(); err != nil {
		log.Infof("error while close db: %s", err)
	}

}
