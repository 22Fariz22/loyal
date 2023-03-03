package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/22Fariz22/loyal/internal/config"

	"github.com/22Fariz22/loyal/internal/auth"
	authHttp "github.com/22Fariz22/loyal/internal/auth/delivery/http"
	authRepo "github.com/22Fariz22/loyal/internal/auth/repository/postgres"
	"github.com/22Fariz22/loyal/internal/auth/usecase"
	"github.com/22Fariz22/loyal/internal/history"
	historyHttp "github.com/22Fariz22/loyal/internal/history/delivery/http"
	historyRepo "github.com/22Fariz22/loyal/internal/history/repository/postgres"
	historyUsecase "github.com/22Fariz22/loyal/internal/history/usecase"
	"github.com/22Fariz22/loyal/internal/order"
	orderHttp "github.com/22Fariz22/loyal/internal/order/delivery/http"
	orderRepo "github.com/22Fariz22/loyal/internal/order/repository/postgres"
	orderUsecase "github.com/22Fariz22/loyal/internal/order/usecase"
	"github.com/22Fariz22/loyal/internal/worker"
	workerRepo "github.com/22Fariz22/loyal/internal/worker/repository/postgres"
	workerUsecase "github.com/22Fariz22/loyal/internal/worker/usecase"
	"github.com/22Fariz22/loyal/pkg/logger"
	"github.com/22Fariz22/loyal/pkg/postgres"
	"github.com/gin-gonic/gin"
)

type App struct {
	cfg        *config.Config
	httpServer *http.Server
	authUC     auth.UseCase
	orderUC    order.UseCase
	historyUC  history.UseCase
	workerUC   worker.UseCase
}

func NewApp(cfg *config.Config) *App {
	// Repository
	db, err := postgres.New(cfg.DatabaseURI, postgres.MaxPoolSize(2))
	if err != nil {
		log.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}

	//defer db.Close()

	userRepo := authRepo.NewUserRepository(db)
	orderRepo := orderRepo.NewOrderRepository(db)
	historyRepo := historyRepo.NewHistoryRepository(db)
	workerRepo := workerRepo.NewWorkerRepository(db)

	return &App{
		cfg: cfg,
		authUC: usecase.NewAuthUseCase(
			userRepo,
			"hash_salt",
			[]byte("signing_key"),
			time.Duration(86400),
		),
		orderUC:   orderUsecase.NewOrderUseCase(orderRepo),
		historyUC: historyUsecase.NewHistoryUseCase(historyRepo),
		workerUC:  workerUsecase.NewWorkerUseCase(workerRepo),
	}
}

func (a *App) Run() error {
	l := logger.New("debug")

	// Init gin handler
	router := gin.Default()
	router.Use(
		gin.Recovery(),
		gin.Logger(),
	)

	// Set up http handlers
	// SignUp/SignIn endpoints
	authHttp.RegisterHTTPEndpoints(router, a.authUC, l)

	// API endpoints
	authMiddleware := authHttp.NewAuthMiddleware(a.authUC, l)
	api := router.Group("/", authMiddleware)

	orderHttp.RegisterHTTPEndpointsOrder(api, a.orderUC, l)
	historyHttp.RegisterHTTPEndpoints(api, a.historyUC, l)

	// HTTP Server
	a.httpServer = &http.Server{
		Addr:           a.cfg.RunAddress, //":" + "8088", //8080
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go worker.CollectNewOrders(a.workerUC, l, a.cfg) //запуск по тикеру

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil {
			l.Fatal("Failed to listen and serve: %+v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Interrupt)

	<-quit

	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()

	return a.httpServer.Shutdown(ctx)
}
