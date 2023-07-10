/*
 * @Descripttion:
 * @version:
 * @Date: 2023-07-03 20:48:07
 * @LastEditTime: 2023-07-03 20:49:47
 */
package components

import (
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var _ log.Logger = (*ZapLogger)(nil)

// Zap 结构体
type ZapLogger struct {
	log  *zap.Logger
	Sync func() error
}

// 创建一个 ZapLogger 实例
func NewZapLogger(encoder zapcore.EncoderConfig, level zap.AtomicLevel, opts ...zap.Option) *ZapLogger {
	writeSyncer := getLogWriter()
	// 设置 zapcore
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoder),
		zapcore.NewMultiWriteSyncer(
			writeSyncer,
		), level)
	//  new 一个 *zap.Logger
	zapLogger := zap.New(core, opts...)
	return &ZapLogger{log: zapLogger, Sync: zapLogger.Sync}
}

// Log 方法实现了 kratos/log/log.go 中的 Logger interface
func (l *ZapLogger) Log(level log.Level, keyvals ...interface{}) error {
	if len(keyvals) == 0 || len(keyvals)%2 != 0 {
		l.log.Warn(fmt.Sprint("Keyvalues must appear in pairs: ", keyvals))
		return nil
	}
	// 按照 KV 传入的时候,使用的 zap.Field
	var data []zap.Field
	for i := 0; i < len(keyvals); i += 2 {
		data = append(data, zap.Any(fmt.Sprint(keyvals[i]), fmt.Sprint(keyvals[i+1])))
	}
	switch level {
	case log.LevelDebug:
		l.log.Debug("", data...)
	case log.LevelInfo:
		l.log.Info("", data...)
	case log.LevelWarn:
		l.log.Warn("", data...)
	case log.LevelError:
		l.log.Error("", data...)
	}
	return nil
}

// 日志自动切割，采用 lumberjack 实现的
func getLogWriter() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   "./gpt-meeting-service.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}
