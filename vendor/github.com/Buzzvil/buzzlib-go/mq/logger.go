package mq

// logger interface makes any logger structure (log.Logger, logrus.Logger, ...) acceptable
type logger interface {
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type mqLoggerWrapper struct {
	logger logger
}

const loggerPrefix = "[MQ] "

func (l *mqLoggerWrapper) Infof(format string, args ...interface{}) {
	l.logger.Infof(loggerPrefix+format, args...)
}

func (l *mqLoggerWrapper) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(loggerPrefix+format, args...)
}

func (l *mqLoggerWrapper) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(loggerPrefix+format, args...)
}

func newMQLoggerWrapper(logger logger) logger {
	return &mqLoggerWrapper{logger}
}
