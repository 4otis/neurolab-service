package app

import (
	"context"
	"net/http"
	"time"

	_ "github.com/4otis/geonotify-service/docs"
	"github.com/4otis/neurolab-service/config"
	"github.com/4otis/neurolab-service/internal/adapter/repo/docker"
	"github.com/4otis/neurolab-service/internal/adapter/repo/postgres"
	"github.com/4otis/neurolab-service/internal/cases"
	"github.com/4otis/neurolab-service/internal/entity"
	httphandler "github.com/4otis/neurolab-service/internal/handler"
	"github.com/4otis/neurolab-service/pkg/logger"
	"github.com/docker/docker/client"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

type App struct {
	config       config.Config
	logger       *zap.Logger
	httpServer   *http.Server
	dbPool       *pgxpool.Pool
	pipelineInCh chan entity.PipelineReq
	// pipelineOutCh chan entity.PipelineResp
	wp           *cases.WorkerPool
	dockerClient *client.APIClient
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
	if err := app.initPipelineProcessors(buf); err != nil {
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

func (a *App) initPipelineProcessors(bufSize int) error {
	pipelineInCh := make(chan entity.PipelineReq, bufSize)
	pipelineOutCh := make(chan entity.PipelineResp, bufSize)

	a.pipelineInCh = pipelineInCh

	labRepo := &postgres.LabRepo{}
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return err
	}
	dockerRepo := docker.NewDockerRepo(cli)

	autotestUseCase := cases.NewAutotestUseCase(
		a.logger,
		labRepo,
		dockerRepo,
		a.config.SolutionsDir,
		a.config.ScriptsDir,
	)

	a.wp = cases.NewWorkerPool(
		bufSize,
		pipelineInCh,
		pipelineOutCh,
		a.logger,
		autotestUseCase,
	)

	return nil
}

func (a *App) initUseCasesAndHandlers() error {

	uploadUseCase := cases.NewUploadUseCase(
		a.logger,
		a.pipelineInCh,
		a.config.SolutionsDir,
		a.config.ScriptsDir,
	)

	// llmClient := clients.NewLLMClient(
	// 	a.config.LLMBaseURL,
	// 	a.config.LLMToken,
	// 	a.config.LLMModel,
	// )

	// teacherLabUseCase := cases.NewTeacherLabUseCase(llmClient)

	llmRepo := gigachat.NewLLMRepo()
	llmUseCase := cases.NewLLMUseCase(
		a.logger,
		llmRepo
	)

	llmHandler := httphandler.NewLLMHandler(llmUseCase)

	httpStudentHandler := httphandler.NewStudentHandler(
		a.logger,
		uploadUseCase,
	)

	httpTeacherHandler := httphandler.NewTeacherHandler(
		a.logger,
		uploadUseCase,
		// teacherLabUseCase,
	)

	r := chi.NewRouter()

	r.Use(logger.Log(a.logger))
	r.Use(middleware.Timeout(30 * time.Second))

	r.Post("/test/generate", llmHandler.GenerateCourse)
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/homepage", func(r chi.Router) {

			r.Route("/student/{student_id}", func(r chi.Router) {

				r.Route("/course/{course_id}", func(r chi.Router) {

				})

				r.Route("/lab/{lab_id}", func(r chi.Router) {

					r.Post("/upload", httpStudentHandler.UploadLab)
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
						// r.Post("/generate", httpTeacherHandler.GenerateLab)
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

func (a *App) Start() error {
	ctx := context.Background()

	go func() {
		a.wp.Start(ctx)
	}()

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
