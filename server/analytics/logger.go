package analytics

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogFileDataCollector struct {
	fileName string
	logger   *zap.Logger
}

func NewLogFileDataCollector(fileName string) (*LogFileDataCollector, error) {
	enccoderConfig := zap.NewProductionEncoderConfig()
	enccoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	enccoderConfig.StacktraceKey = "" // to hide stacktrace info
	fileEncoder := zapcore.NewJSONEncoder(enccoderConfig)
	logFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	writer := zapcore.AddSync(logFile)
	defaultLogLevel := zapcore.InfoLevel
	core := zapcore.NewTee(zapcore.NewCore(fileEncoder, writer, defaultLogLevel))
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return &LogFileDataCollector{
		fileName: fileName,
		logger:   logger,
	}, nil
}

func (lc *LogFileDataCollector) RecordActionSuccess(wfName string, flowId string, actionName string, actionId int, data map[string]any) {
	lc.logger.Info("success", zap.String("name", wfName), zap.String("id", flowId), zap.String("action", actionName), zap.Int("actionId", actionId), zap.Any("data", data))
}
func (lc *LogFileDataCollector) RecordActionFailure(wfName string, flowId string, actionName string, actionId int, reson string) {
	lc.logger.Info("failure", zap.String("name", wfName), zap.String("id", flowId), zap.String("action", actionName), zap.Int("actionId", actionId), zap.String("reason", reson))
}
