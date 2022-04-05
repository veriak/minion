package api

import (	
	"crypto/tls"
		
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/ansrivas/fiberprometheus/v2"		
	
	"github.com/rs/zerolog/log"
	
	"github.com/veriak/minion/config"
	"github.com/veriak/minion/internal"
)

type appServer struct {	
	cfg     *config.Config
	app	*fiber.App	
}

var (
	appSrv		*appServer
	zlogger	= log.With().Str("service", "Minion").Logger()	
)

func NewAppServer(cfg *config.Config) *appServer {

	thumbnailer.Start()
		
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})	

	if cfg.Observability.Prometheus.Enabled {
		prometheus := fiberprometheus.New("Minion")
		prometheus.RegisterAt(app, "/metrics")
		app.Use(prometheus.Middleware)
	}

	app.Use(recover.New())
	//app.Use(logger.New())
	app.Use(logger.New(logger.Config{
		Format: "{\"level\":\"info\",\"service\":\"Minion\",\"error\":\"${status}\",\"time\":\"${time}\",\"message\":\"${locals:requestid} ${latency} ${method} ${path}\"}\n",
	}))	

	api := app.Group("/api", middleware)
	v1 := api.Group("/v1", middleware)
	v1.Get("/health", healthCheck)
	v1.Get("/info", info)
			
	appSrv = &appServer{
		cfg:	cfg,
		app:	app,		
	}
	
	return appSrv	
}

func (appSrv *appServer) ListenAndServe() chan error {
	errCh := make(chan error)
	go func() {    	
		if appSrv.cfg.Server.Cert != "" && appSrv.cfg.Server.Key != "" {
			zlogger.Info().Msgf("Started listening addr https://" + appSrv.cfg.Server.Addr)
			cer, err := tls.LoadX509KeyPair(appSrv.cfg.Server.Cert, appSrv.cfg.Server.Key)
			if err != nil {
				panic(err)
			}

			config := &tls.Config{Certificates: []tls.Certificate{cer}}

			ln, err := tls.Listen("tcp", appSrv.cfg.Server.Addr, config)
			if err != nil {
				panic(err)
			}
			
			errCh <- appSrv.app.Listener(ln)			
		} else {
			zlogger.Info().Msgf("Started listening addr http://" + appSrv.cfg.Server.Addr)
			errCh <- appSrv.app.Listen(appSrv.cfg.Server.Addr)
		}
	}()
	return errCh
}
