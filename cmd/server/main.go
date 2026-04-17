// Trading Journal HTTP Server
//
// @title           Trading Journal API
// @version         1.0
// @description     REST API for the Trading Journal – track cauciones (fixed-term investments), capital movements, and compare investment instruments.
//
// @contact.name    Trading Journal
//
// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         http
package main

import (
	"alearmas/tradingJournal/internal/adapter/filejson"
	"alearmas/tradingJournal/internal/adapter/ppi"
	sqladapter "alearmas/tradingJournal/internal/adapter/sqlite"
	"alearmas/tradingJournal/internal/handler"
	"alearmas/tradingJournal/internal/port"
	"alearmas/tradingJournal/internal/service"
	"io"
	"log"
	"os"
	"strconv"

	_ "alearmas/tradingJournal/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	dataPath := getenv("JOURNAL_DATA", "data/cauciones.json")
	store := getenv("JOURNAL_STORE", "json")

	var repo port.CaucionRepository
	var mrepo port.MovimientoRepository

	switch store {
	case "json":
		repo = filejson.NewFileJSONRepository(dataPath)
		mpath := getenv("JOURNAL_MOVEMENTS", "data/movimientos.json")
		mrepo = filejson.NewMovimientoFileJSONRepository(mpath)
	case "sqlite":
		dbPath := getenv("JOURNAL_DB", "data/journal.db")
		sqlRepo, err := sqladapter.NewSQLiteRepository(dbPath)
		if err != nil {
			log.Fatalf("error opening sqlite: %v", err)
		}
		repo = sqlRepo

		msqlRepo, err := sqladapter.NewMovimientoSQLiteRepository(dbPath)
		if err != nil {
			log.Fatalf("error opening sqlite for movimientos: %v", err)
		}
		mrepo = msqlRepo
	default:
		log.Fatal("invalid JOURNAL_STORE, use json|sqlite")
	}

	if closer, ok := repo.(io.Closer); ok {
		defer func() { _ = closer.Close() }()
	}
	if closer, ok := mrepo.(io.Closer); ok {
		defer func() { _ = closer.Close() }()
	}

	caucionSvc := service.NewCaucionService(repo)
	movimientoSvc := service.NewMovimientoService(mrepo)

	caucionH := handler.NewCaucionHandler(caucionSvc)
	movimientoH := handler.NewMovimientoHandler(movimientoSvc)
	capitalH := handler.NewCapitalHandler(caucionSvc, movimientoSvc)

	// PPI integration (optional – only active when PPI_PUBLIC_KEY is set)
	var syncH *handler.SyncHandler
	var ppiDebugH *handler.PPIDebugHandler
	if apiKey := getenv("PPI_PUBLIC_KEY", ""); apiKey != "" {
		sandbox, _ := strconv.ParseBool(getenv("PPI_SANDBOX", "true"))
		ppiClient := ppi.NewClient(ppi.Config{
			Sandbox:          sandbox,
			AuthorizedClient: getenv("AUTHORIZED_CLIENT", ""),
			ClientKey:        getenv("CLIENT_KEY", ""),
			APIKey:           apiKey,
			APISecret:        getenv("PPI_PRIVATE_KEY", ""),
		})
		accountNumber := getenv("PPI_ACCOUNT", "")
		if accountNumber == "" {
			log.Println("WARNING: PPI_ACCOUNT not set – sync endpoint will fail at runtime")
		}
		syncSvc := service.NewSyncService(ppiClient, accountNumber, movimientoSvc, caucionSvc)
		syncH = handler.NewSyncHandler(syncSvc)
		ppiDebugH = handler.NewPPIDebugHandler(ppiClient, accountNumber)
		log.Printf("PPI integration enabled (sandbox=%v)", sandbox)
	} else {
		log.Println("PPI integration disabled (PPI_PUBLIC_KEY not set)")
	}

	r := gin.Default()

	api := r.Group("/api/v1")
	{
		// Cauciones
		api.POST("/cauciones", caucionH.Create)
		api.GET("/cauciones", caucionH.List)
		api.GET("/cauciones/summary", caucionH.Summary)
		api.GET("/cauciones/report", caucionH.MonthlyReport)
		api.GET("/cauciones/export", caucionH.ExportCSV)

		// Movimientos
		api.POST("/movimientos/deposit", movimientoH.Deposit)
		api.POST("/movimientos/withdraw", movimientoH.Withdraw)
		api.GET("/movimientos", movimientoH.List)

		// Capital
		api.GET("/capital/balance", capitalH.Balance)

		// Compare
		api.GET("/compare", handler.Compare)

		// PPI (only registered when PPI is configured)
		if syncH != nil {
			api.POST("/sync", syncH.Sync)
			api.GET("/ppi/movements", ppiDebugH.RawMovements)
			api.GET("/ppi/orders", ppiDebugH.RawOrders)
		}
	}

	// Swagger UI at /swagger/index.html
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	addr := getenv("JOURNAL_ADDR", ":8080")
	log.Printf("Server starting on %s", addr)
	log.Printf("Swagger UI: http://localhost%s/swagger/index.html", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}
