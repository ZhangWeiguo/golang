package logger

// 文件日志记录器
// 先Init，再AddLogger
type FileLogger map[string]*FileLogWriter

func NewFileLogger() FileLogger {
	return make(FileLogger)
}

func (l FileLogger) AddLogger(key string, path string, fileName string, bufferLength int, console bool) error {
	fileLogWriter, err := NewDefaultFileLogWriter(path, fileName, bufferLength, console)
	l[key] = fileLogWriter
	return err
}

func (l FileLogger) AddHourlyLogger(key string, path string, fileName string, bufferLength int, console bool, maxBackup int) error {
	fileLogWriter, err := NewHourlyFileLogWriter(path, fileName, bufferLength, console, maxBackup)
	l[key] = fileLogWriter
	return err
}

func (l FileLogger) AddDailyLogger(key string, path string, fileName string, bufferLength int, console bool, maxBackup int) error {
	fileLogWriter, err := NewDailyFileLogWriter(path, fileName, bufferLength, console, maxBackup)
	l[key] = fileLogWriter
	return err
}

func (l FileLogger) AddSizeLogger(key string, path string, fileName string, maxSize int64, bufferLength int, console bool, maxBackup int) error {
	fileLogWriter, err := NewSizeFileLogWriter(path, fileName, maxSize, bufferLength, console, maxBackup)
	l[key] = fileLogWriter
	return err
}

func (l FileLogger) AddLineLogger(key string, path string, fileName string, maxLine int64, bufferLength int, console bool, maxBackup int) error {
	fileLogWriter, err := NewLineFileLogWriter(path, fileName, maxLine, bufferLength, console, maxBackup)
	l[key] = fileLogWriter
	return err
}

func (l FileLogger) Close() {
	for name, filter := range l {
		filter.Close()
		delete(l, name)
	}
}

// 获取一个FileLogWriter对象
func (l FileLogger) GetWriter(key string) *FileLogWriter {
	return l[key]
}

// 获取FileLogWriter对象的函数
func (l FileLogger) GetLogFun(level Level, key string) func(string) {
	return l[key].GetLogFun(level)
}
