package logger

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Level int

const (
	FINEST   = Level(0)
	FINE     = Level(1)
	DEBUG    = Level(2)
	TRACE    = Level(3)
	INFO     = Level(4)
	WARNING  = Level(5)
	ERROR    = Level(6)
	CRITICAL = Level(7)
)

var (
	levels = map[Level]string{
		FINEST:   "FINEST",
		FINE:     "FINE",
		DEBUG:    "DEBUG",
		TRACE:    "TRACE",
		INFO:     "INFO",
		WARNING:  "WARNING",
		ERROR:    "ERROR",
		CRITICAL: "CRITICAL",
	}
)

type Locate struct {
	FileName string
	Func     string
	Line     int
}

type LogRecord struct {
	LogLevel Level     // The log level
	Location Locate    // The Log's location
	Created  time.Time // The time at which the log message was created (nanoseconds)
	Message  string    // The log message
}

// 单个日志记录器
// 日志Rotate规则优先级 hourly > daily > size = lines
type FileLogWriter struct {
	rec  chan *LogRecord
	read chan bool
	// The opened file
	Path       string
	FilePrefix string
	file       *os.File
	// print in console
	Console bool
	// buffer length
	BufferLength int
	// Keep old logfiles (.001, .002, etc)
	MaxBackup int
	// Rotate at linecount
	MaxLine        int64
	maxLineCurLine int64
	// Rotate at size
	Maxsize        int64
	maxsizeCurSize int64
	// Current file nums
	currentFileNums int64
	// Rotate daily/hourly
	Daily   bool
	Hourly  bool
	logTime time.Time
	// Define log level and rotate type
	MinLevel   Level
	rotateType string
	// Lock
	rotateMutex sync.RWMutex
}

// a logWriter without rotate
func NewDefaultFileLogWriter(path string, filename string, bufferLength int, console bool) (*FileLogWriter, error) {
	var fileLogWriter FileLogWriter
	fileLogWriter = FileLogWriter{
		Path:         path,
		FilePrefix:   filename,
		Console:      console,
		Hourly:       false,
		Daily:        false,
		Maxsize:      0,
		MaxLine:      0,
		BufferLength: bufferLength,
		MaxBackup:    168,
		MinLevel:     FINEST,
	}
	err := fileLogWriter.Init()
	return &fileLogWriter, err
}

// a hourly logWriter
func NewHourlyFileLogWriter(path string, filename string, bufferLength int, console bool, maxBackup int) (*FileLogWriter, error) {
	var fileLogWriter FileLogWriter
	fileLogWriter = FileLogWriter{
		Path:         path,
		FilePrefix:   filename,
		Console:      console,
		Hourly:       false,
		Daily:        false,
		Maxsize:      0,
		MaxLine:      0,
		BufferLength: bufferLength,
		MaxBackup:    168,
		MinLevel:     FINEST,
	}
	fileLogWriter.SetHourly()
	fileLogWriter.SetMaxBackup(maxBackup)
	err := fileLogWriter.Init()
	return &fileLogWriter, err
}

// a daily logWriter
func NewDailyFileLogWriter(path string, filename string, bufferLength int, console bool, maxBackup int) (*FileLogWriter, error) {
	var fileLogWriter FileLogWriter
	fileLogWriter = FileLogWriter{
		Path:         path,
		FilePrefix:   filename,
		Console:      console,
		Hourly:       false,
		Daily:        false,
		Maxsize:      0,
		MaxLine:      0,
		BufferLength: bufferLength,
		MaxBackup:    168,
		MinLevel:     FINEST,
	}
	fileLogWriter.SetDaily()
	fileLogWriter.SetMaxBackup(maxBackup)
	err := fileLogWriter.Init()
	return &fileLogWriter, err
}

// a maxsize logWriter
func NewSizeFileLogWriter(path string, filename string, maxsize int64, bufferLength int, console bool, maxBackup int) (*FileLogWriter, error) {
	var fileLogWriter FileLogWriter
	fileLogWriter = FileLogWriter{
		Path:         path,
		FilePrefix:   filename,
		Console:      console,
		Hourly:       false,
		Daily:        false,
		Maxsize:      0,
		MaxLine:      0,
		BufferLength: bufferLength,
		MaxBackup:    168,
		MinLevel:     FINEST,
	}
	fileLogWriter.SetMaxSize(maxsize)
	fileLogWriter.SetMaxBackup(maxBackup)
	err := fileLogWriter.Init()
	return &fileLogWriter, err
}

// a maxLine logWriter
func NewLineFileLogWriter(path string, filename string, maxLine int64, bufferLength int, console bool, maxBackup int) (*FileLogWriter, error) {
	var fileLogWriter FileLogWriter
	fileLogWriter = FileLogWriter{
		Path:         path,
		FilePrefix:   filename,
		Console:      console,
		Hourly:       false,
		Daily:        false,
		Maxsize:      0,
		MaxLine:      0,
		BufferLength: bufferLength,
		MaxBackup:    168,
		MinLevel:     FINEST,
	}
	fileLogWriter.SetMaxLine(maxLine)
	fileLogWriter.SetMaxBackup(maxBackup)
	err := fileLogWriter.Init()
	return &fileLogWriter, err
}

func (w *FileLogWriter) Init() error {
	switch {
	case w.Hourly:
		w.rotateType = "Hourly"
	case w.Daily:
		w.rotateType = "Daily"
	case w.Maxsize > 0:
		w.rotateType = "MaxSize"
	case w.MaxLine > 0:
		w.rotateType = "MaxLine"
	default:
		w.rotateType = "None"
	}
	w.logTime = getFileCreateTime(w.Path, w.FilePrefix)
	w.file = initFile(w.Path, w.FilePrefix)
	w.maxLineCurLine = getFileLine(w.Path, w.FilePrefix)
	w.maxsizeCurSize = getFileSize(w.Path, w.FilePrefix)
	w.read = make(chan bool)
	w.rec = make(chan *LogRecord, w.BufferLength)
	w.rotateMutex = sync.RWMutex{}
	go w.write()
	if w.rotateType != "None" {
		if w.MaxBackup <= 0 {
			w.MaxBackup = 100000000000
		}
		go w.delete()
	}
	return nil

}

func (w *FileLogWriter) GetLogFun(level Level) func(string) {
	return func(msg string) {
		pc, file, line, ok := runtime.Caller(1)
		if ok == true {
			f := runtime.FuncForPC(pc)
			_, currentFile := filepath.Split(file)
			fun := f.Name()
			rec := LogRecord{
				Created:  time.Now(),
				LogLevel: level,
				Message:  msg,
				Location: Locate{
					FileName: currentFile,
					Line:     line,
					Func:     fun,
				},
			}
			w.rec <- &rec
		}
	}
}

func (w *FileLogWriter) Info(msg string) {
	w.addLog(INFO, msg)
}
func (w *FileLogWriter) Finest(msg string) {
	w.addLog(FINEST, msg)
}
func (w *FileLogWriter) Fine(msg string) {
	w.addLog(FINE, msg)
}
func (w *FileLogWriter) Debug(msg string) {
	w.addLog(DEBUG, msg)
}
func (w *FileLogWriter) Trace(msg string) {
	w.addLog(TRACE, msg)
}
func (w *FileLogWriter) Warning(msg string) {
	w.addLog(WARNING, msg)
}
func (w *FileLogWriter) Critical(msg string) {
	w.addLog(CRITICAL, msg)
}
func (w *FileLogWriter) Error(msg string) {
	w.addLog(ERROR, msg)
}

func (w *FileLogWriter) addLog(level Level, msg string) {
	pc, file, line, ok := runtime.Caller(2)
	if ok == true {
		f := runtime.FuncForPC(pc)
		_, currentFile := filepath.Split(file)
		fun := f.Name()
		rec := LogRecord{
			Created:  time.Now(),
			LogLevel: level,
			Message:  msg,
			Location: Locate{
				FileName: currentFile,
				Line:     line,
				Func:     fun,
			},
		}
		w.rec <- &rec
	}
}

// ToDo: delete log file in time
func (w *FileLogWriter) delete() {
	if w.rotateType == "None" {
		return
	}
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		fileList, err := ioutil.ReadDir(w.Path)
		removeList := make([]os.FileInfo, 0, len(fileList))
		if err == nil {

			for _, file := range fileList {
				if strings.Contains(file.Name(), w.FilePrefix) {
					removeList = append(removeList, file)
				}
			}
			if len(removeList) <= w.MaxBackup+1 {
				continue
			}
			removeNum := len(removeList) - w.MaxBackup - 1
			for i := 0; i < len(removeList); i++ {
				for j := i + 1; j < len(removeList); j++ {

					if removeList[j-1].ModTime().Nanosecond() >= removeList[j].ModTime().Nanosecond() {
						tempInfo := removeList[j]
						removeList[j] = removeList[j-1]
						removeList[j-1] = tempInfo
					}
				}
			}
			for i := 0; i < removeNum; i++ {
				err = os.Remove(filepath.Join(w.Path, removeList[i].Name()))
				if err == nil {
					w.Info("Remove OutDate LogFile Succ: " + removeList[i].Name())
				} else {
					w.Info("Remove OutDate LogFile Fail: " + err.Error())
				}
			}
		}
	}
}

func (w *FileLogWriter) Close() {
	w.Flush()
	// close(w.rec)
}

func (w *FileLogWriter) write() {
	for {
		w.rotate()
		record := <-w.rec
		log := LogRecord2String(record)
		if record.LogLevel >= w.MinLevel {
			_, _ = fmt.Fprintln(w.file, log)
			atomic.AddInt64(&w.maxLineCurLine, 1)
			atomic.AddInt64(&w.maxsizeCurSize, int64(len([]byte(log))))
		}
		if w.Console {
			fmt.Println(log)
		}
	}
}

func (w *FileLogWriter) Flush() {
	_ = w.file.Sync()
}

func (w *FileLogWriter) rotate() {
	switch w.rotateType {
	case "Hourly":
		w.rotateHourly()
	case "Daily":
		w.rotateDaily()
	case "MaxSize":
		w.rotateMaxSize()
	case "MaxLine":
		w.rotateMaxLine()
	default:
		return
	}
}

func (w *FileLogWriter) rotateMaxSize() {
	if w.maxsizeCurSize = getFileSize(w.Path, w.FilePrefix); w.maxsizeCurSize >= w.Maxsize {
		oldFName := filepath.Join(w.Path, w.FilePrefix)
		newFName := filepath.Join(w.Path, fmt.Sprintf("%s.%s.%s", w.FilePrefix,
			w.logTime.Format("20060102150405"), strconv.Itoa(int(w.currentFileNums))))
		w.changeFile(oldFName, newFName)
	}
}

func (w *FileLogWriter) rotateMaxLine() {
	if w.maxLineCurLine >= w.MaxLine {
		oldFName := filepath.Join(w.Path, w.FilePrefix)
		newFName := filepath.Join(w.Path, fmt.Sprintf("%s.%s.%s", w.FilePrefix,
			w.logTime.Format("20060102150405"), strconv.Itoa(int(w.currentFileNums))))
		w.changeFile(oldFName, newFName)
	}
}

func (w *FileLogWriter) rotateDaily() {
	currentDay := time.Now().Day()
	if w.logTime.Day() != currentDay {
		oldFName := filepath.Join(w.Path, w.FilePrefix)
		newFName := filepath.Join(w.Path, w.FilePrefix+w.logTime.Format(".20060102"))
		w.changeFile(oldFName, newFName)
	}
}

func (w *FileLogWriter) rotateHourly() {
	currentHour := time.Now().Hour()
	currentDay := time.Now().Day()
	if w.logTime.Hour() != currentHour || w.logTime.Day() != currentDay {
		oldFName := filepath.Join(w.Path, w.FilePrefix)
		newFName := filepath.Join(w.Path, w.FilePrefix+w.logTime.Format(".2006010215"))
		w.changeFile(oldFName, newFName)
	}
}

func (w *FileLogWriter) changeFile(oldFName string, newFName string) {
	w.rotateMutex.Lock()
	defer w.rotateMutex.Unlock()
	_ = w.file.Close()
	_ = os.Rename(oldFName, newFName)
	w.file = initFile(w.Path, w.FilePrefix)
	atomic.StoreInt64(&w.maxLineCurLine, 0)
	atomic.StoreInt64(&w.maxsizeCurSize, 0)
	atomic.AddInt64(&w.currentFileNums, 1)
	w.logTime = time.Now()
}

func (w *FileLogWriter) SetMaxBackup(maxBackup int) {
	w.MaxBackup = maxBackup
}

func (w *FileLogWriter) SetConsole(console bool) {
	w.Console = console
}
func (w *FileLogWriter) SetDaily() {
	w.Daily = true
	w.Hourly = false
}
func (w *FileLogWriter) SetHourly() {
	w.Hourly = true
	w.Daily = false
}
func (w *FileLogWriter) SetMaxSize(maxSize int64) {
	if maxSize > 0 {
		w.Daily = false
		w.Hourly = false
		w.MaxLine = 0
		w.Maxsize = maxSize
	} else {
		panic("maxSize must > 0")
	}
}
func (w *FileLogWriter) SetMaxLine(maxLine int64) {
	if maxLine > 0 {
		w.Daily = false
		w.Hourly = false
		w.Maxsize = 0
		w.MaxLine = maxLine
	} else {
		panic("maxLine must > 0")
	}
}

func (w *FileLogWriter) SetBufferLength(bufferLength int) {
	w.BufferLength = bufferLength
}

func (w *FileLogWriter) SetLogLevel(level Level) {
	w.MinLevel = level
}

// LogRecord 2 string
func LogRecord2String(logRecord *LogRecord) string {
	sTime := logRecord.Created.Format("2006/01/02 15:04:05 MST")
	sLocation := fmt.Sprintf("%s:%s:%d", logRecord.Location.FileName, logRecord.Location.Func, logRecord.Location.Line)
	sLevel := logRecord.LogLevel
	return fmt.Sprintf("[%s] [%s] (%s) : %s", sTime, levels[sLevel], sLocation, logRecord.Message)
}

func initFile(path string, fname string) *os.File {
	filename := filepath.Join(path, fname)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0)
	if err != nil {
		panic("Open LogFile Failed")
	}
	return file
}

func getFileSize(path string, fName string) int64 {
	fileName := filepath.Join(path, fName)
	fileInfo, _ := os.Stat(fileName)
	return fileInfo.Size()
}

func getFileCreateTime(path string, fName string) time.Time {
	fileName := filepath.Join(path, fName)
	fileInfo, err := os.Stat(fileName)
	if err == nil {
		return fileInfo.ModTime()
	}
	return time.Now()
}

func getFileLine(path string, fName string) int64 {
	filename := filepath.Join(path, fName)
	file, err := os.Open(filename)
	defer file.Close()
	if err == nil {
		k := 0
		reader := bufio.NewReader(file)
		for {
			_, _, err = reader.ReadLine()
			if err != io.EOF {
				k++
			} else {
				break
			}
		}
		return int64(k)
	}
	return 0
}
