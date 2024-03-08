package cmd

import (
	_ "embed"

	"github.com/getsentry/sentry-go"
	sentryotel "github.com/getsentry/sentry-go/otel"
	"github.com/happyann/golang-base-lib/env"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/TheZeroSlave/zapsentry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//go:generate sh -c "printf %s $(git rev-parse HEAD) > version.txt"
//go:embed version.txt
var Version string

func initSentry() bool {
	sentryDSN := env.GetEnvWithFallback("SENTRY_DSN", "")
	if sentryDSN == "" {
		return false
	}

	if err := sentry.Init(
		sentry.ClientOptions{
			Dsn:              sentryDSN,
			EnableTracing:    true,
			TracesSampleRate: 1.0,
			Release:          Version,
			Environment:      env.GetEnvWithFallback(env.StageEnv, ""),
			BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
				zap.S().Debugf("Sentry event: %s", event.Message)
				return event
			},
		},
	); err != nil {
		zap.L().Fatal("Could not initialize sentry", zap.Error(err))
		return false
	}

	return true
}

func LoadEnv() {
	// init env
	if env.IsLocal() {
		err := godotenv.Load("local.env")
		if err != nil {
			zap.L().Fatal("No config file 'local.env' found for non prod env", zap.Error(err))
		}
	}

	// check if we should init sentry
	useSentry := initSentry()

	var logger *zap.Logger

	if env.IsProd() {
		l, err := zap.NewProduction(
			zap.AddStacktrace(zap.ErrorLevel), zap.Fields(
				zap.Field{
					Key:    "Version",
					Type:   zapcore.StringType,
					String: Version,
				},
			),
		)
		if err != nil {
			l.Fatal("can't initialize zap logger", zap.Error(err))
			return
		}
		logger = l
	} else {
		l, err := zap.NewDevelopment(zap.AddStacktrace(zap.ErrorLevel))
		if err != nil {
			l.Fatal("can't initialize zap logger", zap.Error(err))
		}
		logger = l
	}

	// Setup zapsentry
	if useSentry {
		core, err := zapsentry.NewCore(
			zapsentry.Configuration{
				Level:             zapcore.ErrorLevel, // when to send message to sentry
				EnableBreadcrumbs: true,               // enable sending breadcrumbs to Sentry
				BreadcrumbLevel:   zapcore.InfoLevel,  // at what level should we sent breadcrumbs to sentry
				Tags: map[string]string{
					"Version":     Version,
					"Stage":       env.GetEnvWithFallback(env.StageEnv, ""),
					"environment": env.GetEnvWithFallback(env.StageEnv, ""),
				},
			},
			zapsentry.NewSentryClientFromClient(sentry.CurrentHub().Client()),
		)

		if err != nil {
			logger.Fatal("can't initialize zapsentry", zap.Error(err))
		}
		logger = zapsentry.AttachCoreToLogger(core, logger)
	}

	zap.ReplaceGlobals(logger)

	zap.S().Infof("Build Version: %s", Version)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sentryotel.NewSentrySpanProcessor()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(sentryotel.NewSentryPropagator())
}
