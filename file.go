package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewFileLogger(path, name string, size, backup, age, skip, level int) (*zap.Logger, error) {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	ec := zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   callerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	highSync := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(path, "error."+name+".log"),
		MaxSize:    size, // megabytes
		MaxBackups: backup,
		MaxAge:     age, // days
		LocalTime:  true,
		Compress:   true,
	})
	totalSync := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(path, name+".log"),
		MaxSize:    size, // megabytes
		MaxBackups: backup,
		MaxAge:     age, // days
		LocalTime:  true,
		Compress:   true,
	})

	high := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.WarnLevel
	})
	total := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.Level(level)
	})

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewConsoleEncoder(ec), highSync, high),
		zapcore.NewCore(zapcore.NewConsoleEncoder(ec), totalSync, total),
	)
	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(skip)), nil
}

func callerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("%s %s", caller.TrimmedPath(), trimFunc(runtime.FuncForPC(caller.PC).Name())))
}

func trimFunc(fn string) string {
	idx := strings.LastIndexByte(fn, '/')
	if idx == -1 {
		return fn
	}
	return fn[idx+1:]
}
