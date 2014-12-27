package web

import (
	"fmt"
	"github.com/euforia/spinal-cord/config"
	"github.com/euforia/spinal-cord/logging"
	"github.com/euforia/spinal-cord/reactor"
	"net/http"
)

type CoreWebService struct {
	listenAddr  string
	handlersMgr *reactor.HandlersManager
	router      *RESTRouter
	logger      *logging.Logger
}

func NewCoreWebService(cfg *config.SpinalCordConfig, logger *logging.Logger) *CoreWebService {
	if cfg.Core.Web.Webroot != "" {
		http.Handle("/", http.FileServer(http.Dir(cfg.Core.Web.Webroot)))
		logger.Warning.Printf("Webroot: %s\n", cfg.Core.Web.Webroot)
	} else {
		logger.Warning.Printf("Web UI disabled! No webroot specified.")
	}

	return &CoreWebService{
		listenAddr:  fmt.Sprintf(":%d", cfg.Core.Web.Port),
		handlersMgr: reactor.NewHandlersManager(cfg.Core.HandlersDir, logger),
		router:      NewRESTRouter("/api/ns", "*", logger), /* prefix, default acl, logger */
		logger:      logger,
	}
}

func (c *CoreWebService) registerEndpoints() {

	c.router.Register("/", NewNamespaceHandle(c.handlersMgr))
	c.router.Register("/namespace", NewEventTypeHandle(c.handlersMgr))
	c.router.Register("/namespace/eventType", NewEventTypeHandlersHandle(c.handlersMgr))
	c.router.Register("/namespace/eventType/handler", NewEventHandlerHandle(c.handlersMgr))
	http.Handle("/api/ns/", c.router)
}

func (c *CoreWebService) Start() {
	c.registerEndpoints()
	c.logger.Warning.Printf("Spawning core HTTP service on: %s\n", c.listenAddr)
	go func() {
		c.logger.Error.Fatal(http.ListenAndServe(c.listenAddr, nil))
	}()
}
