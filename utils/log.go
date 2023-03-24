package utils

import (
	nested "github.com/antonfisher/nested-logrus-formatter"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"time"
)

func NewLfsHook(rotationTime time.Duration, maxRemainNum uint, path, moduleName string) logrus.Hook {
	lfsHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: initRotateLogs(rotationTime, maxRemainNum, "debug", path, moduleName),
		logrus.InfoLevel:  initRotateLogs(rotationTime, maxRemainNum, "info", path, moduleName),
		logrus.WarnLevel:  initRotateLogs(rotationTime, maxRemainNum, "warn", path, moduleName),
		logrus.ErrorLevel: initRotateLogs(rotationTime, maxRemainNum, "error", path, moduleName),
	}, &nested.Formatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		HideKeys:        false,
		FieldsOrder:     []string{"pId", "filePath", "operationId"},
	})
	return lfsHook
}
func initRotateLogs(rotationTime time.Duration, maxRemainNum uint, level string, path, moduleName string) *rotatelogs.RotateLogs {
	if moduleName != "" {
		moduleName = moduleName + "."
	}

	//rotatelogs.WithLinkName(linkName), // 生成软链，指向最新日志文件
	//rotatelogs.WithMaxAge(-1), // 文件最大保存时间
	//rotatelogs.WithRotationCount(50), // 最多文件数 WithMaxAge 与WithRotationCount 二选一
	//rotatelogs.WithRotationTime(-1), // 日志切割时间间隔
	//rotatelogs.WithRotationSize(8*1024*1024), // 日志切割大小 WithRotateTime 与WithRotationSize 二选一

	writer, err := rotatelogs.New(
		path+"/"+moduleName+level+"."+"%Y-%m-%d-%H"+".log",
		rotatelogs.WithRotationTime(rotationTime),
		rotatelogs.WithRotationCount(maxRemainNum),
	)
	if err != nil {
		panic(any(err.Error()))
	} else {
		return writer
	}
}
