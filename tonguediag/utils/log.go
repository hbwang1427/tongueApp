package utils

import (
	"math"
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logger         *zap.Logger
	loggerInitOnce sync.Once
)

//Logger returns logger
func Logger(config *Config) *zap.Logger {
	if logger != nil {
		return logger
	}
	if config == nil {
		config = AppConfig()
	}

	loggerInitOnce.Do(func() {
		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "linenum",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
			EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 UTC 时间格式
			EncodeDuration: zapcore.SecondsDurationEncoder, //
			EncodeCaller:   zapcore.FullCallerEncoder,      // 全路径编码器
			EncodeName:     zapcore.FullNameEncoder,
		}

		// 设置日志级别
		level := zap.DebugLevel
		switch strings.ToLower(config.Logger.Level) {
		case "info":
			level = zap.InfoLevel
		case "warn":
			level = zap.WarnLevel
		case "error":
			level = zap.ErrorLevel
		case "panic":
			level = zap.PanicLevel
		case "fatal":
			level = zap.FatalLevel
		}

		atomicLevel := zap.NewAtomicLevelAt(level)

		var allCore []zapcore.Core

		//append file writer
		logFile := config.Logger.File
		if logFile == "" && !config.IsDevelop {
			os.MkdirAll("./logs/", 0777)
			logFile = "./logs/myapp.log"
		}
		if logFile != "" {
			hook := lumberjack.Logger{
				Filename:   logFile,                                           // 日志文件路径
				MaxSize:    128,                                               // 每个日志文件保存的最大尺寸 单位：M
				MaxBackups: 30,                                                // 日志文件最多保存多少个备份
				MaxAge:     int(math.Ceil(config.Logger.MaxAge.Hours() / 24)), // 文件最多保存多少天
				Compress:   config.Logger.Compress,                            // 是否压缩
			}
			allCore = append(allCore, zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig),
				zapcore.AddSync(&hook), atomicLevel))
		}

		if config.IsDevelop {
			//add console encoder in develop mode.
			consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
			consoleDebugging := zapcore.Lock(os.Stdout)
			allCore = append(allCore, zapcore.NewCore(consoleEncoder, consoleDebugging, atomicLevel))
		}
		core := zapcore.NewTee(allCore...)
		logger = zap.New(core)

		if config.IsDevelop {
			logger = logger.WithOptions(zap.AddCaller(), zap.Development())
		}

		//logger.Info("log 初始化成功")
		// logger.Info("无法获取网址",
		// 	zap.String("url", "http://www.baidu.com"),
		// 	zap.Int("attempt", 3),
		// 	zap.Duration("backoff", time.Second))
	})

	return logger
}
