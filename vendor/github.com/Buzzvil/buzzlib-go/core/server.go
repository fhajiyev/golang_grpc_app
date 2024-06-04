package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	pluginEcho "github.com/Buzzvil/buzzlib-go/plugin/echo"
	"github.com/go-playground/validator"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// Server provides methods for controlling application lifecycle
type Server interface {
	Init(configPath string, configStruct interface{})
	Start()
	Exit()
	Engine() *Engine
}

// Service provides methods for interacting with a microservice instance
type Service interface {
	Init() error
	RegisterRoute(serviceDriver *Engine)
	Health() error
	Clean() error
}

// Context is a router context
type Context = echo.Context

// Engine is a http service drive
type Engine = echo.Echo

// Router group
type RouterGroup = echo.Group

// Middleware function
type MiddlewareFunc = echo.MiddlewareFunc

// Handler function
type HandlerFunc = echo.HandlerFunc

// Http error
type HttpError = echo.HTTPError

// Fields logger fields
type Fields = logrus.Fields

type ErrorListener interface {
	OnError(err error, c Context)
}

var (
	// Logger is the default application logger
	Logger = getLogger(false)
	// Config is the application configuration
	Config *viper.Viper
	// Loggers contains applications loggers
	Loggers = map[string]*logrus.Logger{}
)

type server struct {
	services       []Service
	settings       *settings
	engine         *Engine
	httpServer     *http.Server
	errorListeners []ErrorListener
}

type reqValidator struct {
	validator *validator.Validate
}

func (cv *reqValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func (s *server) Engine() *Engine {
	if s.engine != nil {
		return s.engine
	}
	s.engine = NewEngine(s.settings)

	return s.engine
}

func NewEngine(s *settings) *Engine {
	engine := echo.New()
	engine.Validator = &reqValidator{validator: validator.New()}

	// Because dd trace middleware calls error handler that log error messages,
	// does not pass error to echo to prevent duplicated error handler call.
	// refer: https://github.com/labstack/echo/blob/v3.3.10/echo.go#L594
	engine.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			_ = next(c)
			return nil
		}
	})

	if s != nil && s.config.GetString("logger.level") == "debug" {
		engine.Debug = true
		engine.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "method=${method}, uri=${uri}, status=${status}\n",
		}))
	}

	engine.Use(middleware.Recover())
	engine.Use(middleware.CORS())
	engine.Use(pluginEcho.DatadogTrace())
	engine.Use(pluginEcho.NewMetric())

	// Newrelic
	if s != nil && s.newrelic != nil {
		engine.Use(pluginEcho.NewRelicWithApplication(s.newrelic))
		Logger.Info("Newrelic is enabled!")
	}
	return engine
}

// Init initializes the dependencies
func (s *server) Init(configPath string, configStruct interface{}) {
	s.settings = &settings{}
	err := s.settings.init(configPath, configStruct)
	if err != nil {
		panic(err)
	}
	Config = s.settings.config
	// Loggers
	s.initLoggers()
	engine := s.Engine()
	s.initErrorHandlers()
	for _, service := range s.services {
		// Init services
		err := service.Init()
		if err != nil {
			Logger.Error("Failed to initialize service: ", err)
			panic(err)
		}
		// Routes
		service.RegisterRoute(engine)
	}
	s.registerMiscRoute(engine)
}

func (s *server) registerMiscRoute(engine *Engine) {
	// global routes
	engine.GET("/health", s.healthController)
	engine.GET("/", s.homeController)
}

func (s *server) homeController(c Context) error {
	appName := s.settings.config.GetString("name")
	version := s.settings.config.GetString("version")
	return c.JSON(http.StatusOK, map[string]interface{}{"app": appName, "version": version})
}

func (s *server) healthController(c Context) error {
	for _, service := range s.services {
		err := service.Health()
		if err != nil {
			Logger.Error("Service health error: ", err)
			return err
		}
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"server": "healthy"})
}

// Start will start the server with service
func (s *server) Start() {
	if s.settings == nil {
		s.Init("", nil)
	}
	Config.SetDefault("server.port", 8080)
	port := Config.GetInt("server.port")
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: s.engine,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs

		s.Exit()
		close(idleConnsClosed)
	}()

	if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		Logger.Errorf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed
}

// Exit will gracefully shutdown the server and close services.
func (s *server) Exit() {
	Logger.Info("Start to shutdown the server ...")
	if err := s.httpServer.Shutdown(context.Background()); err != nil {
		Logger.Errorf("HTTP server Shutdown: %v", err)
	}

	for _, service := range s.services {
		err := service.Clean()
		if err != nil {
			Logger.Error("Failed to clean service: ", err)
		}
	}
	Logger.Info("The server is down")
}

// initializes loggers
func (s *server) initLoggers() {
	if Config.IsSet("logger") {
		config := Config.GetStringMapString("logger")
		initLogger(Logger, config) // default logger
		if s.settings.sentry != nil && config["sentry"] == "true" {
			registerSentryHook(Logger, s.settings.sentry)
		}
	}
	if Config.IsSet("loggers") {
		// Multiple loggers
		loggers := Config.GetStringMap("loggers")
		for k := range loggers {
			config := Config.GetStringMapString("loggers." + k)
			Loggers[k] = getLogger(true)
			initLogger(Loggers[k], config)
			if s.settings.sentry != nil && config["sentry"] == "true" {
				registerSentryHook(Logger, s.settings.sentry)
			}
		}
	}
}

func (s *server) initErrorHandlers() {
	s.errorListeners = []ErrorListener{
		&defaultErrorListener{s.engine},
	}

	if s.settings.sentry != nil {
		Logger.Info("Sentry is enabled!")
		s.errorListeners = append(s.errorListeners, pluginEcho.NewSentryErrorHandler(s.settings.sentry))
	}

	s.engine.HTTPErrorHandler = func(e error, i echo.Context) {
		for _, el := range s.errorListeners {
			el.OnError(e, i)
		}
	}
}

type defaultErrorListener struct {
	engine *Engine
}

func (del *defaultErrorListener) OnError(err error, c Context) {
	del.engine.DefaultHTTPErrorHandler(err, c)

	m := map[string]interface{}{
		"level":      "error",
		"method":     c.Request().Method,
		"uri":        c.Path(),
		"status":     c.Response().Status,
		"user_agent": c.Request().UserAgent(),
		"error":      err.Error(),
	}

	span, ok := tracer.SpanFromContext(c.Request().Context())
	if ok {
		m["dd"] = map[string]interface{}{
			"span_id":  span.Context().SpanID(),
			"trace_id": span.Context().TraceID(),
		}
	}

	values, err := c.FormParams()
	if err == nil {
		m["params"] = values
	} else {
		m["params"] = c.QueryParams()
	}

	marshaled, err := json.Marshal(m)
	if err != nil {
		Logger.Warnf("error handler err: %v", err)
		return
	}
	fmt.Printf("%s\n", string(marshaled)) // TODO replace structured log
}

// NewServer initializes a new instance of server with all services
func NewServer(services ...Service) Server {
	server := &server{services: services}
	return Server(server)
}
