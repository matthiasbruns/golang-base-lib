package cmd

import (
	_ "embed"
	"log"

	"github.com/getsentry/sentry-go"
	sentryotel "github.com/getsentry/sentry-go/otel"
	"github.com/joho/godotenv"
	"github.com/matthiasbruns/golang-base-lib/env"
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
			Environment:      env.GetEnvWithFallback(env.Env, ""),
			BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
				zap.S().Debugf("Sentry event: %s", event.Message)
				return event
			},
		},
	); err != nil {
		log.Fatal("Could not initialize sentry", err)
		return false
	}

	return true
}

func LoadEnv(localEnv string) {
	// init env
	if env.IsLocal() {
		err := godotenv.Load(localEnv)
		if err != nil {
			log.Fatal("No config file 'local.env' found for non prod env", err)
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
			log.Fatal("can't initialize zap logger", err)
			return
		}
		logger = l
	} else {
		l, err := zap.NewDevelopment(zap.AddStacktrace(zap.ErrorLevel))
		if err != nil {
			log.Fatal("can't initialize zap logger", err)
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
					"environment": env.GetEnvWithFallback(env.Env, ""),
				},
			},
			zapsentry.NewSentryClientFromClient(sentry.CurrentHub().Client()),
		)

		if err != nil {
			log.Fatal("can't initialize zapsentry", err)
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
