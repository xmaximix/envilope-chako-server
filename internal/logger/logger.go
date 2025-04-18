package logger

import "go.uber.org/zap"

func NewLogger() (*zap.SugaredLogger, error) {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "ts"
	base, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	return base.Sugar(), nil
}
