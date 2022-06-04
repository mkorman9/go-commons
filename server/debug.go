package server

import (
	"github.com/gookit/config/v2"
	"net/http"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

type debugAPI struct {
	metricsPath        string
	pprofPath          string
	healthcheckPath    string
	healthCheckHandler HealthCheckHandlerFunc
	startupTime        string
	appName            string
	appVersion         string
	deploymentName     string
}

type healthcheckResponse struct {
	Status         string `json:"status"`
	AppName        string `json:"app,omitempty"`
	AppVersion     string `json:"version,omitempty"`
	DeploymentName string `json:"deploymentName"`
	StartupTime    string `json:"startupTime"`
}

type HealthCheckHandlerFunc = func() bool

type DebugAPIOpt = func(*debugAPI)

func AppName(appName string) DebugAPIOpt {
	return func(debugAPI *debugAPI) {
		debugAPI.appName = appName
	}
}

func AppVersion(appVersion string) DebugAPIOpt {
	return func(debugAPI *debugAPI) {
		debugAPI.appVersion = appVersion
	}
}

func DeploymentName(deploymentName string) DebugAPIOpt {
	return func(debugAPI *debugAPI) {
		debugAPI.deploymentName = deploymentName
	}
}

func DebugAPI(engine *gin.Engine, opts ...DebugAPIOpt) {
	debugAPI := &debugAPI{
		metricsPath:     "/debug/metrics",
		pprofPath:       "/debug/pprof",
		healthcheckPath: "/debug/health",
		startupTime:     time.Now().UTC().Format(time.RFC3339),
	}

	metricsPath := config.String("debug.metrics.path")
	if metricsPath != "" {
		debugAPI.metricsPath = metricsPath
	}

	pprofPath := config.String("debug.pprof.path")
	if pprofPath != "" {
		debugAPI.pprofPath = pprofPath
	}

	healthcheckPath := config.String("debug.healthcheck.path")
	if healthcheckPath != "" {
		debugAPI.healthcheckPath = healthcheckPath
	}

	for _, opt := range opts {
		opt(debugAPI)
	}

	metrics := ginprometheus.NewPrometheus("gin")
	metrics.MetricsPath = debugAPI.metricsPath
	metrics.Use(engine)

	pprof.Register(engine, debugAPI.pprofPath)

	engine.GET(debugAPI.healthcheckPath, debugAPI.healthCheck)
}

func (api *debugAPI) healthCheck(c *gin.Context) {
	if api.healthCheckHandler != nil {
		if ok := api.healthCheckHandler(); !ok {
			c.JSON(http.StatusInternalServerError, api.responseBody("unhealthy"))
			return
		}
	}

	c.JSON(http.StatusOK, api.responseBody("healthy"))
}

func (api *debugAPI) responseBody(status string) *healthcheckResponse {
	return &healthcheckResponse{
		Status:         status,
		AppName:        api.appName,
		AppVersion:     api.appVersion,
		DeploymentName: api.deploymentName,
		StartupTime:    api.startupTime,
	}
}
