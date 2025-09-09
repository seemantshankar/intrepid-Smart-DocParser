package container

import (
	"contract-analysis-service/configs"
	"contract-analysis-service/internal/handlers"
	"contract-analysis-service/internal/pkg/cache"
	"contract-analysis-service/internal/pkg/database"
	"contract-analysis-service/internal/pkg/external"
	"contract-analysis-service/internal/pkg/logger"
	"contract-analysis-service/internal/pkg/metrics"
	"contract-analysis-service/internal/pkg/storage"
	"contract-analysis-service/internal/pkg/tracing"
	"contract-analysis-service/internal/repositories"
	"contract-analysis-service/internal/repositories/sqlite"
	"contract-analysis-service/internal/services/document"
	"contract-analysis-service/internal/services/analysis"
	"contract-analysis-service/internal/services/knowledge"
	"contract-analysis-service/internal/services/llm"
	llmclient "contract-analysis-service/internal/services/llm/client"
	"contract-analysis-service/internal/services/ocr"
	"contract-analysis-service/internal/services/validation"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

// Container holds all application dependencies
type Container struct {
	Config     *configs.Config
	Logger     *zap.Logger
	DB          *gorm.DB
	Tracer      *tracing.TracerProvider
	RedisClient *redis.Client
	OCRTMetrics *metrics.OCRMetrics

	// Repositories
	ContractRepo       repositories.ContractRepository
	KnowledgeRepo      repositories.KnowledgeEntryRepository
	ValidationRepo     repositories.ValidationRepository
	ValidationAuditRepo repositories.ValidationAuditRepository
	ValidationFeedbackRepo repositories.ValidationFeedbackRepository

	// Services
	LLMService        llm.Service
	OCRService        ocr.Service
	DocumentService   document.Service
	ValidationService validation.Service
	AnalysisService   analysis.Service
	KnowledgeService  knowledge.Service
}

// NewContainer creates and initializes a new Container
func NewContainer(cfg *configs.Config) *Container {
	// Initialize logger
	logger := logger.NewLogger(cfg.Logger)

	// Initialize database
	db, err := database.NewDB(cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}

	// Initialize Redis
	redisClient, err := cache.NewRedisClient(cfg.Redis)
	if err != nil {
		logger.Fatal("failed to connect to Redis", zap.Error(err))
	}

	// Initialize metrics
	ocrMetrics := metrics.NewOCRMetrics()

	// Set log mode if enabled
	if cfg.Database.GetLogMode() {
		db = db.Debug()
	}

	// Initialize tracing only if Jaeger URL is provided
	var tp *tracing.TracerProvider
	if cfg.Jaeger.URL != "" {
		tp, err = tracing.InitTracer(tracing.Config{
			ServiceName:  cfg.ServiceName,
			Environment:  cfg.Environment,
			JaegerURL:    cfg.Jaeger.URL,
			SamplingRate: cfg.Jaeger.SamplingRate,
		}, logger)
		if err != nil {
			logger.Fatal("failed to initialize tracing", zap.Error(err))
		}
	} else {
		logger.Info("tracing disabled, no Jaeger URL provided")
	}

		// Initialize services
	llmService := llm.NewLLMService(logger)
	llmclient.AddOpenRouterClientToService(llmService, cfg)

	resilientClient := external.NewHTTPClient(
		cfg.LLM.OpenRouter.BaseURL,
		"OCR",
		external.RetryConfig{
			MaxRetries:      3, // Example values
			InitialInterval: 2 * time.Second,
			MaxInterval:     10 * time.Second,
		},
		cfg.LLM.OpenRouter.Timeout,
	)
	ocrValidator := ocr.NewValidator()
	baseOCRService := ocr.NewOCRService(resilientClient, cfg.OCR.APIKey, cfg.OCR.FallbackModels, ocrValidator, ocrMetrics)
	ocrService := ocr.NewCachedOCRService(baseOCRService, redisClient, logger, 1*time.Hour) // Cache for 1 hour

	// Initialize storage
	fileStorage, err := storage.NewLocalStorage("./uploads")
	if err != nil {
		logger.Fatal("failed to create file storage", zap.Error(err))
	}

	// Initialize repositories
	contractRepo := sqlite.NewContractRepository(db)
	knowledgeRepo := sqlite.NewKnowledgeEntryRepository(db)
	validationRepo := sqlite.NewValidationRepository(db)
	validationAuditRepo := sqlite.NewValidationAuditRepository(db)
	validationFeedbackRepo := sqlite.NewValidationFeedbackRepository(db)

	validationService := validation.NewValidationService(llmService, logger, validationRepo, validationAuditRepo, validationFeedbackRepo)
	analysisService := analysis.NewService(llmService)
	documentService := document.NewDocumentService(contractRepo, fileStorage, analysisService, logger)
	knowledgeService := knowledge.NewKnowledgeService(llmService, logger, knowledgeRepo, redisClient)

	return &Container{
		Config:                 cfg,
		Logger:                 logger,
		DB:                     db,
		Tracer:                 tp,
		RedisClient:            redisClient,
		OCRTMetrics:            ocrMetrics,
		ContractRepo:           contractRepo,
		KnowledgeRepo:          knowledgeRepo,
		ValidationRepo:         validationRepo,
		ValidationAuditRepo:    validationAuditRepo,
		ValidationFeedbackRepo: validationFeedbackRepo,
		LLMService:             llmService,
		OCRService:             ocrService,
		DocumentService:        documentService,
		ValidationService:      validationService,
		AnalysisService:        analysisService,
		KnowledgeService:       knowledgeService,
	}
}

// NewHealthHandler creates a new health check handler
func (c *Container) NewHealthHandler() *handlers.HealthHandler {
	return handlers.NewHealthHandler(c.DB, c.RedisClient)
}

// NewDocumentHandler creates a new document handler
func (c *Container) NewDocumentHandler() *handlers.DocumentHandler {
	return handlers.NewDocumentHandler(c.DocumentService, c.ValidationService, c.Logger)
}

// NewAnalysisHandler creates a new analysis handler
func (c *Container) NewAnalysisHandler() *handlers.AnalysisHandler {
	return handlers.NewAnalysisHandler(c.Logger, c.OCRService, c.AnalysisService)
}
