package app

import (
	"context"
	"net/http"

	"github.com/4otis/neurolab-service/config"
	"github.com/4otis/neurolab-service/pkg/logger"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type App struct {
	config     config.Config
	logger     *zap.Logger
	httpServer *http.Server
	dbPool     *pgxpool.Pool
}

func New(cfg *config.Config) (*App, error) {
	zapLogger, err := logger.NewDevelopment(cfg.LogLevel)
	if err != nil {
		return nil, err
	}

	app := &App{
		config: *cfg,
		logger: zapLogger,
	}

	if err := app.initDB(); err != nil {
		return nil, err
	}

	if err := app.initUseCasesAndHandlers(); err != nil {
		return nil, err
	}

	return app, nil
}

func (a *App) Run() error {
	return nil
}

func (a *App) Stop() {

}

func (a *App) initDB() error {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, a.config.DBURL)
	if err != nil {
		return err
	}
	a.dbPool = pool

	if err := pool.Ping(ctx); err != nil {
		return err
	}

	a.logger.Info("Database connected successfully")
	return nil
}

func (a *App) initUseCasesAndHandlers() error {
	// incidentRepo := postgres.NewIncidentRepo(a.dbPool)

	// locationUseCase := cases.NewLocationUseCase(
	// 	incidentRepo,
	// 	checkRepo,
	// 	webhookRepo,
	// 	a.redisClient,
	// 	a.logger,
	// 	a.config.CacheTTLMinutes,
	// )

	// httpIncidentHandler := httphandler.NewIncidentHandler(
	// 	a.logger,
	// 	incidentUseCase,
	// )

	r := chi.NewRouter()

	// r.Use(logger.Log(a.logger))
	// r.Use(middleware.Timeout(30 * time.Second))

	// r.Post("/api/v1/location/check", httpLocationHandler.LocationCheck)

	// r.Route("/api/v1/incidents", func(r chi.Router) {

	// 	r.Post("/", httpIncidentHandler.IncidentCreate)
	// })

	// r.Get("/swagger/*", httpSwagger.WrapHandler)

	a.httpServer = &http.Server{
		Addr:    ":" + a.config.HTTPPort,
		Handler: r,
	}

	return nil
}
