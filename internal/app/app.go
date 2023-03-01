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
	http2 "github.com/22Fariz22/loyal/internal/auth/delivery/http"
	postgres2 "github.com/22Fariz22/loyal/internal/auth/repository/postgres"
	"github.com/22Fariz22/loyal/internal/auth/usecase"
	"github.com/22Fariz22/loyal/internal/history"
	delivery3 "github.com/22Fariz22/loyal/internal/history/delivery/http"
	postgres4 "github.com/22Fariz22/loyal/internal/history/repository/postgres"
	usecase3 "github.com/22Fariz22/loyal/internal/history/usecase"
	"github.com/22Fariz22/loyal/internal/order"
	delivery2 "github.com/22Fariz22/loyal/internal/order/delivery/http"
	postgres3 "github.com/22Fariz22/loyal/internal/order/repository/postgres"
	usecase2 "github.com/22Fariz22/loyal/internal/order/usecase"
	"github.com/22Fariz22/loyal/internal/worker"
	postgres5 "github.com/22Fariz22/loyal/internal/worker/repository/postgres"
	usecase4 "github.com/22Fariz22/loyal/internal/worker/usecase"
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

	//log.Println("viper.GetString('a'): ", viper.GetString("a"))
	//log.Println("viper.GetString('d'): ", viper.GetString("d"))
	//log.Println("viper.GetString('r'): ", viper.GetString("r"))
	//
	//log.Println("cfg DatabaseURI: ", cfg.DatabaseURI)

	// Repository
	db, err := postgres.New(cfg.DatabaseURI, postgres.MaxPoolSize(2))
	if err != nil {
		log.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}

	//defer db.Close()

	userRepo := postgres2.NewUserRepository(db)
	orderRepo := postgres3.NewOrderRepository(db)
	historyRepo := postgres4.NewHistoryRepository(db)
	workerRepo := postgres5.NewWorkerRepository(db)

	return &App{
		cfg: cfg,
		authUC: usecase.NewAuthUseCase(
			userRepo,
			"hash_salt",
			[]byte("signing_key"),
			time.Duration(86400),
		),
		orderUC:   usecase2.NewOrderUseCase(orderRepo),
		historyUC: usecase3.NewHistoryUseCase(historyRepo),
		workerUC:  usecase4.NewWorkerUseCase(workerRepo),
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
	http2.RegisterHTTPEndpoints(router, a.authUC, l)

	// API endpoints
	authMiddleware := http2.NewAuthMiddleware(a.authUC, l)
	api := router.Group("/", authMiddleware)

	delivery2.RegisterHTTPEndpointsOrder(api, a.orderUC, l)
	delivery3.RegisterHTTPEndpoints(api, a.historyUC, l)

	// HTTP Server
	a.httpServer = &http.Server{
		Addr:           a.cfg.RunAddress, //":" + "8088", //8080
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Println("app-a.httpServer:,", a.httpServer)

	go worker.CollectNewOrders(a.workerUC, l, a.cfg) //запуск по тикеру

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil {
			l.Fatal("Failed to listen and serve: %+v", err)
		}
	}()
	//if err := http.ListenAndServe(a.cfg.RunAddress, router); err != http.ErrServerClosed {
	//	log.Fatalf("HTTP server ListenAndServe Error: %v", err)
	//}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Interrupt)

	<-quit

	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()

	return a.httpServer.Shutdown(ctx)
}
