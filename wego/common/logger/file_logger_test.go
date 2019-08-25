package logger

import (
	"testing"
	"time"
)

var (
	fileLogs FileLogger
)

func Log(writer func(string)) {
	for i := 0; i < 10000; i++ {
		writer("I am a Info")
	}
}

var (
	logWriter *FileLogWriter
)

func setinfo() {
	for {
		logWriter.Info("I am a Info")
	}
}
func seterror() {
	for {
		logWriter.Error("I am a error")
	}
}
func setcritical() {
	for {
		logWriter.Critical("I am a error")
	}
}

func TestLoggerWriter(t *testing.T) {
	// logWriter = logger.NewDefaultFileLogWriter("./", "logTest", 1000, true)
	logWriter, _ = NewLineFileLogWriter("./", "logTest", 10000, 1000, true, 3)
	_ = logWriter.Init()
	go setinfo()
	go seterror()
	go setcritical()
	time.Sleep(10 * time.Second)
}

func TestFileLogger(t *testing.T) {
	fileLogs = NewFileLogger()
	defer fileLogs.Close()
	_ = fileLogs.AddDailyLogger("Daily", "./", "DailyLog", 100, true, 10)
	// fileLogs.AddLineLogger("Line", "./", "LineLog", 100000)
	// fileLogs.AddSizeLogger("Size", "./", "SizeLog", 100000)
	// fileLogs.AddLogger("Default", "./", "DefaultLog")

	fileLog := fileLogs.GetWriter("Daily")
	logFun := fileLogs.GetLogFun(ERROR, "Daily")

	logFun("I am logfun")
	fileLog.Info("I am logger")
	time.Sleep(10 * time.Second)
}
