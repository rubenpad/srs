package server

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rubenpad/srs/internal/domain/service"
	"github.com/rubenpad/srs/internal/infrastructure/api"
	"github.com/rubenpad/srs/internal/infrastructure/server/handler/health"
	"github.com/rubenpad/srs/internal/infrastructure/server/handler/stock"
	"github.com/rubenpad/srs/internal/infrastructure/server/middleware/logging"
	"github.com/rubenpad/srs/internal/infrastructure/server/middleware/pagination"
	"github.com/rubenpad/srs/internal/infrastructure/server/middleware/search"
	"github.com/rubenpad/srs/internal/infrastructure/storage/cockroach"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type Server struct {
	httpAddress string
	engine      *gin.Engine

	shutdownTimeout time.Duration
}

func New(ctx context.Context, host string, port uint, shutdownTimeout time.Duration, connectionPool *pgxpool.Pool) (context.Context, Server) {
	gin.SetMode(gin.ReleaseMode)

	server := Server{
		engine:          gin.New(),
		httpAddress:     fmt.Sprintf("%s:%d", host, port),
		shutdownTimeout: shutdownTimeout,
	}

	server.registerRoutes(connectionPool)
	return serverContext(ctx), server
}

func (s *Server) registerRoutes(connectionPool *pgxpool.Pool) {
	s.engine.Use(
		gin.Recovery(),
		logging.Middleware(),
		pagination.Middleware(),
		search.Middleware(),
		otelgin.Middleware("srs"),
	)

	stockRatingRepository := cockroach.NewStockRatingRepository(connectionPool)
	stockRatingService := service.NewStockRatingService(stockRatingRepository, api.NewStockRatingApi())
	stockRatingController := stock.NewStockRatingController(stockRatingService)

	s.engine.GET("/api/health", health.HealthCheck)
	s.engine.GET("/api/stock-ratings", stockRatingController.GetStockRatings)
	s.engine.POST("/api/stock-ratings-data", stockRatingController.LoadStockRatingData)
	s.engine.GET("/api/stock-recommendations", stockRatingController.GetStockRecommendations)
	s.engine.GET("/api/stock-details/:ticker", stockRatingController.GetStockDetails)
}

func (s *Server) Run(ctx context.Context) error {
	slog.Info("Server running on", "httpAddress", s.httpAddress)

	server := &http.Server{
		Addr:    s.httpAddress,
		Handler: s.engine,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server shut down", err)
		}
	}()

	<-ctx.Done()
	ctxShutDown, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	return server.Shutdown(ctxShutDown)
}

func serverContext(ctx context.Context) context.Context {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		<-c
		cancel()
	}()

	return ctx
}
