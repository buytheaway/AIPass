package logger

import "go.uber.org/zap"

func New(env string) *zap.Logger {
	if env == "local" || env == "dev" {
		log, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		return log
	}
	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	return log
}
