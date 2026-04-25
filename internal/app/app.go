package app

import (
	"context"
	"net/http"
	"time"

	_ "github.com/4otis/geonotify-service/docs"
	"github.com/4otis/neurolab-service/config"
	"github.com/4otis/neurolab-service/internal/cases"
	"github.com/4otis/neurolab-service/internal/clients"
	"github.com/4otis/neurolab-service/internal/entity"
	httphandler "github.com/4otis/neurolab-service/internal/handler"
	"github.com/4otis/neurolab-service/pkg/logger"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

type App struct {
	config     config.Config
	logger     *zap.Logger
	httpServer *http.Server
	dbPool     *pgxpool.Pool
	pipelineCh chan entity.PipelineReq
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

	// if err := app.initDB(); err != nil {
	// 	return nil, err
	// }

	buf := 1000 // pipelineCh buffer size
	if err := app.initPipelineProcessor(buf); err != nil {
		return nil, err
	}

	if err := app.initUseCasesAndHandlers(); err != nil {
		return nil, err
	}

	return app, nil
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

func (a *App) initPipelineProcessor(bufSize int) error {
	a.pipelineCh = make(chan entity.PipelineReq, bufSize)

	// TODO: инициализация воркер пула
	return nil
}

func (a *App) initUseCasesAndHandlers() error {

	uploadUseCase := cases.NewUploadUseCase(
		a.logger,
		a.pipelineCh,
		a.config.SolutionsDir,
		a.config.ScriptsDir,
	)

	llmClient := clients.NewLLMClient(
		a.config.LLMBaseURL,
		a.config.LLMToken,
		a.config.LLMModel,
	)

	teacherLabUseCase := cases.NewTeacherLabUseCase(llmClient)

	httpStudentHandler := httphandler.NewStudentHandler(
		a.logger,
		uploadUseCase,
	)

	httpTeacherHandler := httphandler.NewTeacherHandler(
		a.logger,
		uploadUseCase,
		teacherLabUseCase,
	)

	r := chi.NewRouter()

	r.Use(logger.Log(a.logger))
	r.Use(middleware.Timeout(30 * time.Second))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/homepage", func(r chi.Router) {

			r.Route("/student/{student_id}", func(r chi.Router) {

				r.Route("/course/{course_id}", func(r chi.Router) {

					r.Route("/lab/{lab_id}", func(r chi.Router) {

						r.Post("/upload", httpStudentHandler.UploadLab)
					})
				})
			})

			r.Route("/teacher/{teacher_id}", func(r chi.Router) {

				r.Route("/course/{course_id}", func(r chi.Router) {

					r.Route("/lab/{lab_id}", func(r chi.Router) {

						// TODO: замени на соответствующие методы TeacherHandler
						r.Patch("/save", httpStudentHandler.UploadLab)
						r.Get("/generate", httpStudentHandler.UploadLab)
						r.Get("/scripts", httpStudentHandler.UploadLab)
						r.Post("/upload", httpStudentHandler.UploadLab)
						r.Post("/generate", httpTeacherHandler.GenerateLab)
					})
				})
			})

		})

		r.Route("/scripts", func(r chi.Router) {
			r.Post("/upload", httpTeacherHandler.UploadScript)
		})
	})

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	a.httpServer = &http.Server{
		Addr:    ":" + a.config.HTTPPort,
		Handler: r,
	}

	return nil
}

func (a *App) Run() error {
	go func() {
		a.logger.Info("Starting HTTP server",
			zap.String("port", a.config.HTTPPort),
			zap.String("env", a.config.LogLevel))

		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("HTTP server error", zap.Error(err))
		}
	}()

	return nil
}

func (a *App) Stop() {
	a.logger.Info("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	if a.dbPool != nil {
		a.dbPool.Close()
		a.logger.Info("Database connection closed")
	}

	a.logger.Sync()
	a.logger.Info("Servers stopped gracefully")
}
