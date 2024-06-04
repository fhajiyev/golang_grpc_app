# Scaffold for Go Service
This project provides the basic structure for [echo](https://echo.labstack.com/) go service.

# Prerequisite
Go 1.9 or later

# How to create service
- Create a `config/` dir inside your repo and add neccssary config files (`config.[ENV].json`) for every env (i.e local, dev, prod, etc..):
  - Make neccessary changes by refering `config/config.sample.json`
- Create and attach your service instance to the server.
  - You service should implement the `server.Service` interface 
```go
package main

import (
  "github.com/Buzzvil/buzzlib-go/core"
  "github.com/Buzzvil/your-project/yourservice"
)

func main() {
	yourservice := yourservice.NewService()
	server := core.NewServer(yourservice)
	server.Start()
}
```

# Setup
- Set `SERVER_ENV=env` to load the env configration file (if it is not set, `config.local.json` will be loaded)
- Set `DEBUG=false` to disable debug mode
- Set `ENABLE_NEWRELIC=true` to enable newrelic (Note: you also need to set the config)

# Config
- Application configrations are defined under `./config`. Depending on `SERVER_ENV`, the config file will be loaded. Note: all server env config varaibles are overwritten by varaibles on config file.
- Use `config.local.json` to put all sensetive configs and avoid pushing it to source control.
- Config uses [Viper](https://github.com/spf13/viper) internally. So, the interface is the same with Viper.
> Pro tip: you can pass a config path and the pointer of a config object using `server.Init(configPath, configObject)` before calling `Start`
```go
import "github.com/Buzzvil/buzzlib-go/core"
...
value := core.Config.GetBool("configKey")
```
## Logger
```go
import "github.com/Buzzvil/buzzlib-go/core"
...
core.Logger.Infof("hello %s", "world")
// prints: INFO[2018-02-26 12:20:46] hello world
```
- Set the config variable `logger.level` for to adjust the log level (i.e. `info`, `warn`, etc..)
- Set the config variable `logger.file` to file path when you want to log to file.
- Set the config variable `logger.formatter` to change the log format (e.g. `json`)
- Set the config variable `logger.sentry` to `true` to send error, fatal and panic logs to sentry (see below to configure sentry)
> Pro tip: you can have multiple loggers by setting the config `loggers.[LoggerName].*` then use it like `core.Loggers["LoggerName"].Info(..)`

## newrelic
- Set the config variable `newrelic.app` and `newrelic.key`
- To enable newrelic, set `ENABLE_NEWRELIC=true` either using config file or server env

## sentry
- Set the config variable `sentry.dsn` to enable sentry

# Errors
- Please avoid using `panics`. Always return Error to router context.
- The error middleware will handle all router errors and by sending error response and logging
```go
import "http"
import "github.com/Buzzvil/buzzlib-go/core"

func sampleController(c *core.Context) error {
	result, err := sampleTask()
	if err != nil {
		// Error middleware will send a response with 500 
		// with JSON Body: {"error": err.Error()} if it is debug mode
		// It also make an error log with the stack trace
		return err
	}
	return c.String(http.StatusOK, result)
}

```
- By default, the response code for error is `500`. You can change the response code by setting `HttpStatusCode` property of the Error.

# TODO
- unit test