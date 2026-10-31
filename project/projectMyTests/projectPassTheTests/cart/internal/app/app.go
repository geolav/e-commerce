package app

import (
	"context"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/config"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/controller"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/migrations"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	lomsadapter "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/adapter/loms/grpc"
	productadapter "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/adapter/product/grpc"
	cartcontroller "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/controller/cart"
	cartrepo "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/repository/cart"
	cartusecase "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/cart"
	itemusecase "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/item"
	cartpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"
	pbloms "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	productpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1"
	stockspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/stocks/v1"
)

func Run(logger *zap.Logger, cfg *config.Config) {
	ctx := context.Background()
	pool, err := newPool(ctx, cfg, logger)
	if err != nil {
		logger.Fatal("failed to connect to postgres", zap.Error(err))
	}
	defer pool.Close()

	migrations.SetupPostgres(pool, logger)

	lomsConn, err := grpc.NewClient(
		cfg.Clients.LOMSGrpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Fatal("failed to connect to loms", zap.Error(err))
	}
	defer lomsConn.Close()

	productClient := productadapter.NewProductClient(
		productpb.NewProductServiceClient(lomsConn),
	)
	lomsClient := lomsadapter.NewLOMSClient(
		pbloms.NewLomsClient(lomsConn),
		stockspb.NewStocksClient(lomsConn),
	)

	cartRepository := cartrepo.NewPostgresRepo(pool)

	itemService := itemusecase.NewItemService(cartRepository, productClient, lomsClient)
	cartService := cartusecase.NewCartService(cartRepository, productClient, lomsClient)

	cartServer := cartcontroller.NewCartServer(itemService, cartService)

	ctrl := controller.New(cartServer)

	go runRest(ctx, logger, cfg)
	runGrpc(logger, cfg, ctrl)
}

func runRest(ctx context.Context, logger *zap.Logger, cfg *config.Config) {
	mux := runtime.NewServeMux()
	gatewayPort := ":" + cfg.GRPC.GatewayPort
	grpcPort := "127.0.0.1:" + cfg.GRPC.Port

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := cartpb.RegisterCartHandlerFromEndpoint(ctx, mux, grpcPort, opts); err != nil {
		logger.Fatal("failed to register cart grpc gateway", zap.Error(err))
	}

	handler := corsHandler(mux)
	if err := http.ListenAndServe(gatewayPort, handler); err != nil {
		logger.Error("gateway listen error", zap.Error(err))
	}
}

func runGrpc(logger *zap.Logger, cfg *config.Config, ctrl *controller.API) {
	grpcPort := ":" + cfg.GRPC.Port

	lis, err := net.Listen("tcp4", "0.0.0.0:"+cfg.GRPC.Port)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	cartpb.RegisterCartServer(grpcServer, ctrl.Cart)
	reflection.Register(grpcServer)

	logger.Info("gRPC server started", zap.String("port", grpcPort))
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("failed to serve", zap.Error(err))
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
