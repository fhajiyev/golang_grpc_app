package core

import (
	"github.com/getsentry/raven-go"
	"github.com/newrelic/go-agent"
	"github.com/spf13/viper"
)

// settings builds all the application dependencies
type settings struct {
	config   *viper.Viper
	newrelic newrelic.Application
	sentry   *raven.Client
}

// initializes all the dependencies
func (b *settings) init(configPath string, configStruct interface{}) (err error) {
	if err = b.initConfig(configPath, configStruct); err != nil {
		return
	}
	if err = b.initNewrelic(); err != nil {
		return
	}
	if err = b.initSentry(); err != nil {
		return
	}
	return
}

// initializes configuration
func (b *settings) initConfig(configPath string, configStruct interface{}) (err error) {
	config := viper.New()
	defaultEnv := "local"
	config.SetDefault("SERVER_ENV", defaultEnv)
	config.SetDefault("DEBUG", false)
	config.SetDefault("ENABLE_NEWRELIC", false)
	config.SetDefault("LEADER", false)
	config.AutomaticEnv()
	if configPath != "" {
		config.AddConfigPath(configPath)
	} else {
		config.AddConfigPath("./config/")
	}

	config.SetConfigName("config." + config.GetString("SERVER_ENV"))
	err = config.ReadInConfig()
	if err != nil && config.GetString("SERVER_ENV") == defaultEnv {
		err = nil // In case of local config missing, continue
	}
	if err != nil {
		return
	}
	b.config = config
	if configStruct != nil {
		err = config.Unmarshal(configStruct)
	}
	return
}

// initializes Newrelic
func (b *settings) initNewrelic() (err error) {
	leader := b.config.GetBool("ENABLE_NEWRELIC") || b.config.GetBool("LEADER")
	if !leader || !b.config.IsSet("newrelic") {
		return
	}
	config := b.config.GetStringMapString("newrelic")
	newrelicConfig := newrelic.NewConfig(config["app"], config["key"])
	b.newrelic, err = newrelic.NewApplication(newrelicConfig)
	return
}

// initializes Sentry
func (b *settings) initSentry() (err error) {
	if !b.config.IsSet("sentry") {
		return
	}
	if err = raven.SetDSN(b.config.GetString("sentry.dsn")); err != nil {
		return
	}
	b.sentry = raven.DefaultClient
	return
}
