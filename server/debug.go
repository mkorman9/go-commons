package server

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/gookit/config/v2"
	"github.com/mkorman9/go-commons/info"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"net/http"
)

type healthcheckResponse struct {
	Status         string `json:"status"`
	AppName        string `json:"app,omitempty"`
	AppVersion     string `json:"version,omitempty"`
	DeploymentName string `json:"deploymentName"`
	StartupTime    string `json:"startupTime"`
	BuildCommit    string `json:"buildCommit,omitempty"`
	BuildTime      string `json:"buildTime,omitempty"`
}

type debugAPI struct {
	metricsPath     string
	pprofPath       string
	healthcheckPath string
	response        *healthcheckResponse
}

func DebugAPI(engine *gin.Engine, appInfo info.AppInfo) {
	debugAPI := &debugAPI{
		metricsPath:     "/debug/metrics",
		pprofPath:       "/debug/pprof",
		healthcheckPath: "/debug/health",
		response: &healthcheckResponse{
			Status:         "healthy",
			AppName:        appInfo.Name,
			AppVersion:     appInfo.Version,
			DeploymentName: appInfo.DeploymentName,
			StartupTime:    appInfo.StartupTime,
			BuildCommit:    appInfo.BuildCommit,
			BuildTime:      appInfo.BuildTime,
		},
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

	metrics := ginprometheus.NewPrometheus("gin")
	metrics.MetricsPath = debugAPI.metricsPath
	metrics.Use(engine)

	pprof.Register(engine, debugAPI.pprofPath)

	engine.GET(debugAPI.healthcheckPath, debugAPI.healthCheck)
}

func (api *debugAPI) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, api.response)
}
