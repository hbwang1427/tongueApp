package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"

	"tonguediag/bussiness"
	"tonguediag/utils"
)

var config *utils.Config
var logger *zap.Logger
var confFile = flag.String("c", "config.yaml", "config file path")

func main() {
	flag.Parse()
	utils.SetConfigFile(*confFile)

	//设置当前目录
	os.Chdir(utils.GetWorkDirectory())

	config = utils.AppConfig()
	logger = utils.Logger(config)

	// if "1" == "1" {
	// 	importRegions()
	// 	return
	// }

	s := createHTTPServer(config)
	go s.ListenAndServe()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		logger.Fatal("Server Shutdown:", zap.Error(err))
	}
	if !config.IsDevelop {
		// catching ctx.Done(). timeout of 5 seconds.
		select {
		case <-ctx.Done():
			logger.Info("timeout of 5 seconds.")
		}
	}
	logger.Info("Server exiting")
}

func createHTTPServer(config *utils.Config) *http.Server {
	if !config.IsDevelop {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	// Add a ginzap middleware, which:
	//   - Logs all requests, like a combined access and error log.
	//   - Logs to stdout.
	//   - RFC3339 with UTC time format.
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))

	// Logs all panic to error log
	//   - stack means whether output the stack info.
	r.Use(ginzap.RecoveryWithZap(logger, true))

	//cross domain request config.
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		//AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "GET", "DELETE", "PUT", "PATCH"},
		AllowHeaders:     []string{"Origin", "X-Requested-With", "Content-Type", "Accept", config.Token.AuthName},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		// AllowOriginFunc: func(origin string) bool {
		// 	return origin == "*"
		// },
		MaxAge: 12 * time.Hour,
	}))

	setUpHandlers(r)
	bussiness.Init(config, r)

	var s *http.Server

	if len(config.AutoCert.Domains) > 0 {
		logger.Debug("create https server", zap.Strings("domains", config.AutoCert.Domains))
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(config.AutoCert.Domains...),
			Cache:      autocert.DirCache(config.AutoCert.CacheDir),
		}

		s = &http.Server{
			Addr:      ":https",
			TLSConfig: m.TLSConfig(),
			Handler:   r,
			//ReadTimeout:    10 * time.Second,
			//WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}

		//go s.ListenAndServeTLS("", "")
	} else {
		logger.Debug("create http server", zap.String("bind", config.HTTP.Addr))
		s = &http.Server{
			Addr:    config.HTTP.Addr,
			Handler: r,
			//ReadTimeout:    10 * time.Second,
			//WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}

		//go s.ListenAndServe()
	}

	return s
}

func setUpHandlers(r *gin.Engine) {
	//static files
	r.Static("/assets", "./assets")
	//r.StaticFS("/more_static", http.Dir("my_file_system"))
	r.StaticFile("/favicon.ico", "./assets/favicon.ico")

	// Ping handler
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	r.GET("/", func(c *gin.Context) {
		c.Writer.Write([]byte(""))
	})
}
