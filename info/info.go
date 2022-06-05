package info

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

var startupTime time.Time

func init() {
	startupTime = time.Now().UTC()
}

type AppInfo struct {
	Name           string
	Version        string
	DeploymentName string
	StartupTime    string
	BuildCommit    string
	BuildTime      string
}

func Build(appName, appVersion string) AppInfo {
	deploymentName := os.Getenv("DEPLOYMENT_NAME")
	if deploymentName == "" {
		deploymentName = "default"
	}

	var buildCommit string
	var buildTime string
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				buildCommit = setting.Value
			case "vcs.time":
				buildTime = setting.Value
			}
		}
	}

	return AppInfo{
		Name:           appName,
		Version:        appVersion,
		DeploymentName: deploymentName,
		StartupTime:    startupTime.Format(time.RFC3339),
		BuildCommit:    buildCommit,
		BuildTime:      buildTime,
	}
}

func (info AppInfo) String() string {
	fields := []string{
		fmt.Sprintf("name=%s", info.Name),
		fmt.Sprintf("version=%s", info.Version),
		fmt.Sprintf("deployment=%s", info.DeploymentName),
		fmt.Sprintf("startup=%s", info.StartupTime),
	}

	if info.BuildCommit != "" {
		fields = append(fields, fmt.Sprintf("commit=%s", info.BuildCommit))
	}

	if info.BuildTime != "" {
		fields = append(fields, fmt.Sprintf("buildTime=%s", info.BuildTime))
	}

	return strings.Join(fields, " ")
}
