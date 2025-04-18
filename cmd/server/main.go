package main

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/xmaximix/envilope-chako-server/internal/auth/transport"
	"github.com/xmaximix/envilope-chako-server/internal/config"
	"github.com/xmaximix/envilope-chako-server/internal/db"
	"github.com/xmaximix/envilope-chako-server/internal/logger"
	errs "github.com/xmaximix/envilope-chako-server/pkg/error"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile("config.yaml")
	_ = viper.ReadInConfig()

	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	viper.BindEnv("server.port")
	viper.BindEnv("server.read_timeout")
	viper.BindEnv("server.write_timeout")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(errs.Wrap("reading config file", err))
		}
	}

	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(errs.Wrap("parsing configuration", err))
	}

	cfg.Auth.AccessTokenTTL = viper.GetDuration("auth.access_ttl")
	cfg.Auth.RefreshTokenTTL = viper.GetDuration("auth.refresh_ttl")

	logg, err := logger.NewLogger()
	if err != nil {
		panic(errs.Wrap("initializing logger", err))
	}

	dbConn, err := db.NewPostgres(cfg.Database)
	if err != nil {
		logg.Fatal(errs.Wrap("connecting to PostgreSQL", err))
	}
	defer func(dbConn *sqlx.DB) {
		err := dbConn.Close()
		if err != nil {
			logg.Error(errs.Wrap("closing database connection", err))
		}
	}(dbConn)

	if err := db.MigrateUp(dbConn.DB); err != nil {
		logg.Fatal(errs.Wrap("applying database migrations", err))
	}

	router := gin.New()
	router.Use(gin.Recovery(), transport.ErrorMiddleware(logg))

	router.GET("/healthz", func(c *gin.Context) {
		if err := dbConn.PingContext(c); err != nil {
			c.Status(http.StatusServiceUnavailable)
		} else {
			c.Status(http.StatusOK)
		}
	})

	transport.RegisterAuthRoutes(router, dbConn, cfg.Auth, logg)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logg.Infof("listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logg.Fatalf("ListenAndServe: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logg.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logg.Errorf("server shutdown: %v", err)
	}
}
