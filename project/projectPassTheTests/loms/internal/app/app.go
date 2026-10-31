package app

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/config"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/migrations"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	lomscontroller "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/loms"
	productcontroller "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/product"
	stockscontroller "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/stocks"
	outboxworker "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/outbox"
	orderrepo "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/order"
	outboxrepo "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/outbox"
	productrepo "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/product"
	lomsusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usercase/loms"
	pbloms "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	productpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1"
	stockspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/stocks/v1"
)

const sleepingTime = 200

func Run(logger *zap.Logger, cfg *config.Config) {
	ctx := context.Background()
	pool, err := newPool(ctx, cfg, logger)
	if err != nil {
		logger.Fatal("failed to connect to postgres", zap.Error(err))
	}
	defer pool.Close()

	migrations.SetupPostgres(pool, logger)

	orderRepository := orderrepo.NewPostgresRepo(pool)
	productRepository := productrepo.NewPostgresRepo(pool)
	outboxRepository := outboxrepo.NewPostgresRepo(pool)

	lomsService := lomsusecase.NewLomsService(orderRepository, productRepository, outboxRepository, pool)

	lomsServer := lomscontroller.NewLomsServer(lomsService)
	productServer := productcontroller.NewProductServer(productRepository)
	stocksServer := stockscontroller.NewStocksServer(productRepository)

	control := controller.New(lomsServer, productServer, stocksServer)

	worker := outboxworker.NewWorker(outboxRepository, outboxworker.Config{
		Workers:       cfg.Outbox.Workers,
		BatchSize:     cfg.Outbox.BatchSize,
		FetchPeriod:   cfg.Outbox.FetchPeriod,
		InProgressTTL: cfg.Outbox.InProgressTTL,
		KafkaBrokers:  cfg.Kafka.Brokers,
		KafkaTopic:    cfg.Kafka.Topic,
	}, logger)

	go runGrpc(logger, cfg, control)
	time.Sleep(sleepingTime * time.Millisecond)
	worker.Run(ctx)
	runRest(ctx, logger, cfg)
}

func runRest(ctx context.Context, logger *zap.Logger, cfg *config.Config) {
	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
			logger.Error("gateway error", zap.Error(err))
			runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
		}),
	)

	gatewayPort := ":" + cfg.GRPC.GatewayPort
	grpcPort := "127.0.0.1:" + cfg.GRPC.Port

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := pbloms.RegisterLomsHandlerFromEndpoint(ctx, mux, grpcPort, opts); err != nil {
		logger.Error("REGISTRATION FAILED for loms", zap.Error(err))
		return // не падаем, просто выходим
	}
	if err := productpb.RegisterProductServiceHandlerFromEndpoint(ctx, mux, grpcPort, opts); err != nil {
		logger.Error("REGISTRATION FAILED for product", zap.Error(err))
		return
	}
	if err := stockspb.RegisterStocksHandlerFromEndpoint(ctx, mux, grpcPort, opts); err != nil {
		logger.Error("REGISTRATION FAILED for stocks", zap.Error(err))
		return
	}

	handler := corsHandler(mux)
	if err := http.ListenAndServe(gatewayPort, handler); err != nil {
		logger.Error("gateway listen error", zap.Error(err))
	}
}

func runGrpc(logger *zap.Logger, cfg *config.Config, ctrl *controller.API) {
	grpcPort := ":" + cfg.GRPC.Port
	listen, err := net.Listen("tcp4", "0.0.0.0:"+cfg.GRPC.Port)
	if err != nil {
		logger.Fatal("grpc listen error", zap.Error(err))
	}

	grpsServer := grpc.NewServer()
	pbloms.RegisterLomsServer(grpsServer, ctrl.Loms)
	productpb.RegisterProductServiceServer(grpsServer, ctrl.Product)
	stockspb.RegisterStocksServer(grpsServer, ctrl.Stocks)
	reflection.Register(grpsServer) // mb not need

	logger.Info("gRPC server started", zap.String("port", grpcPort))
	if err := grpsServer.Serve(listen); err != nil { //  тут программа висит пока сервер слушает
		logger.Error("failed to serve on gRPC", zap.Error(err))
	}
}

func corsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "http://localhost:5173"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
