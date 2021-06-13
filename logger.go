package p_log4go

import (
	"fmt"
	"github.com/thiinbit/plog4go/file"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
)

const (
	// One day
	nanoSecInOneDay = int64(time.Hour * 24)
	// One week
	nanoSecInOneWeek = nanoSecInOneDay * 7
)

// ======== ======== PLogger: TimeRotated writer ======== ========

// RotateInterval enum
type RotateInterval string

const (
	// Rotate interval
	Hourly RotateInterval = "Hourly"
	Daily  RotateInterval = "Daily"
	Weekly RotateInterval = "Weekly"
)

// eastUTCOffset east UTC offset in nanoSecs, use to name rotate file
var eastUTCOffset = func() int64 {
	// Calc  its offset in seconds east of UTC.
	_, secOffset := time.Now().In(time.Local).Zone()
	return int64(secOffset) * int64(time.Second)
}()

// TimedRotatingWriter
type TimedRotatingWriter struct {
	lock            sync.Mutex     // Write file lock
	filename        string         // File name
	fp              *os.File       // File pointer
	interval        RotateInterval // File rotating interval
	intervalNanoSec int64          // File rotating interval nanoSec
	format          string         // Rotated file name format
	rotate          int64          // Rotate file count
	rotateDateIndex int64          // Rotate flag
	eastOfUTCOffset int64          // east of UTC offset(nanoSecs)
}

// NewRotateWrite new writer
func NewTimedRotateWriter(filename string, interval RotateInterval, rotate int64) (*TimedRotatingWriter, error) {
	w := &TimedRotatingWriter{
		filename: filename,
		interval: interval,
		rotate:   rotate,
	}

	switch interval {
	case Hourly:
		w.intervalNanoSec = int64(time.Hour)
		w.format = "2006-01-02_15"
	case Daily:
		w.intervalNanoSec = nanoSecInOneDay
		w.format = "2006-01-02"
	case Weekly:
		w.intervalNanoSec = nanoSecInOneWeek
		w.format = "2006-01-02"
	}

	err := w.initialize()
	if err != nil {
		return nil, fmt.Errorf("error when init logger, %s", err)
	}

	return w, nil
}

// initialize
func (w *TimedRotatingWriter) initialize() error {
	if len(w.filename) <= 0 {
		return fmt.Errorf("file name not set when init rotate writer")
	}
	var err error
	fileInfo, err := os.Stat(w.filename)
	if err == nil {
		w.rotateDateIndex = (fileInfo.ModTime().UnixNano() + eastUTCOffset) / w.intervalNanoSec
	} else {
		w.rotateDateIndex = (time.Now().UnixNano() + eastUTCOffset) / w.intervalNanoSec
	}
	w.fp, err = os.OpenFile(w.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	return nil
}

// try rotate
// There may be concurrency problems when renaming files
func (w *TimedRotatingWriter) tryRotate() (err error) {
	// 0. check should exec rotate
	now := time.Now()
	nowDateIndex := (now.UnixNano() + eastUTCOffset) / w.intervalNanoSec
	if nowDateIndex == w.rotateDateIndex {
		return nil
	}
	// 1. close existing file if open
	if w.fp != nil {
		err = w.fp.Close()
		if err != nil {
			fmt.Printf("close exist file error when rotate, file: %s", w.fp.Name())
			return
		}
		w.fp = nil
	}
	// 2. rename dest file if it already exists
	fInfo, err := os.Stat(w.filename)
	if err == nil {
		modeTime := fInfo.ModTime()
		archiveTime := now.Add(-time.Duration(w.intervalNanoSec) * time.Nanosecond)
		if modeTime.Before(archiveTime) {
			archiveTime = modeTime
			// TODO: Check and delete if has more than rotate amount file
		}
		err = os.Rename(w.filename, w.filename+"."+archiveTime.Format(w.format))
		if err != nil {
			fmt.Printf("rename log file error when rotate, file: %s: err: %v", w.filename, err)
			return
		}
	}
	// 3. create a new file
	w.fp, err = os.OpenFile(w.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	// 4. update rotate index
	w.rotateDateIndex = nowDateIndex
	// 5. remove old file (more older file will delete when check mod time is before archive time)
	oldestArchiveTime := now.Add(-time.Duration(w.rotate*w.intervalNanoSec) * time.Nanosecond)
	os.Remove(w.filename + "." + oldestArchiveTime.Format(w.format))
	return
}

func (w *TimedRotatingWriter) Write(output []byte) (int, error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.tryRotate()
	return w.fp.Write(output)
}

var (
	// Even if it is not used, it must be put here, otherwise it will be recycled during GC.
	nullFile   *os.File // Null file /dev/null
	stdoutFile *os.File // stdout stderr
)

// stdOut and stdErr to where ? console or file
type StdOutTo int8

// stdOut and stdErr to where ? console or file
const (
	ToConsole StdOutTo = iota
	ToFile
	ToNull
)

// stdOut to where conf
type StdOutToConf struct {
	To    StdOutTo // Stdout to where (null?console?file?)
	ToDir string   // File dir if to file
}

// InitStdOnce
var initStdOnce sync.Once

// Proc stdin/stdout/stderr. Invoke once on application start if need proc
func InitStd(stdOutTo StdOutToConf) {
	initStdOnce.Do(func() {
		var err error

		if nullFile, err = os.OpenFile(os.DevNull, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err != nil {
			fmt.Printf("open /dev/null, err = [%v]", err)
		}

		// stdin to /dev/null
		if err = file.SyscallDup(int(nullFile.Fd()), int(os.Stdin.Fd())); err != nil {
			fmt.Printf("dup2 stdin to /dev/null, err = [%v]", err)
		}

		// Stdout/stdErr to /dev/null if set ToNull
		if stdOutTo.To == ToNull {
			if err = file.SyscallDup(int(nullFile.Fd()), int(os.Stdout.Fd())); err != nil {
				fmt.Printf("dup2 stdout to null, err = [%v]", err)
			}
			if err = file.SyscallDup(int(nullFile.Fd()), int(os.Stderr.Fd())); err != nil {
				fmt.Printf("dup2 stderr to null, err = [%v]", err)
			}
		}

		// Stdout/StdErr to console if set ToConsole
		if stdOutTo.To == ToConsole {
			// Default is to console
			return
		}

		// Stdout/StdErr to file if set ToFile
		if stdOutTo.To == ToFile {
			if len(stdOutTo.ToDir) < 1 {
				fmt.Printf("stdOutToFile path conf, err = wrongFilePath [%s]", stdOutTo.ToDir)
			}

			// Rotate old stdout file
			latestStdoutPath := path.Join(stdOutTo.ToDir, "stdout.log")
			oldStdoutPath := path.Join(stdOutTo.ToDir, "stdout.log."+time.Now().Format("20060102.150405"))
			os.Rename(latestStdoutPath, oldStdoutPath)
			if stdoutFile, err = os.OpenFile(latestStdoutPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err != nil {
				fmt.Printf("open stdout.log, err = [%v]", err)
			}

			// stdout to stdout.log
			if err = file.SyscallDup(int(stdoutFile.Fd()), int(os.Stdout.Fd())); err != nil {
				fmt.Printf("dup2 stdout to stdout.log, err = [%v]", err)
			}
			// stderr to stdout.log
			if err = file.SyscallDup(int(stdoutFile.Fd()), int(os.Stderr.Fd())); err != nil {
				fmt.Printf("dup2 stderr to stdout.log, err = [%v]", err)
			}
		}

	})
}

// ==== ==== ==== ==== PLogger: Get Logger ==== ==== ==== ====

// Log level DEBUG/INFO/WARN/ERROR
type LogLevel int8

// Log level iota
const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

type PLogger struct {
	loggerInst    *log.Logger
	logLevel      LogLevel // Loglevel DEBUG INFO WARN ERROR
	isTraceEnable bool     // Is trace enable
}

func GetLogger(fileName string, logLevel LogLevel, interval RotateInterval, rotate int64) (*PLogger, error) {

	return getLogger1(fileName, logLevel, interval, rotate, false)
}

func getLogger1(fileName string, logLevel LogLevel, interval RotateInterval, rotate int64, traceOn bool) (*PLogger, error) {

	fileDir := filepath.Dir(fileName)
	exist, err := pathExists(fileDir)
	if err != nil {
		return nil, fmt.Errorf("log file path err, %v", err)
	}

	if !exist {
		if err = os.MkdirAll(fileDir, 0755); err != nil {
			return nil, fmt.Errorf("mkdir logfile dir err, %v", err)
		}
	}
	var writer *TimedRotatingWriter
	writer, err = NewTimedRotateWriter(fileName, interval, rotate)
	if err != nil {
		return nil, fmt.Errorf("create RotateRiter err, %v", err)
	}
	return &PLogger{
		log.New(writer, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile),
		logLevel,
		traceOn,
	}, nil
}

//RawLogger return raw go logger
func (l *PLogger) RawLogger() *log.Logger {
	return l.loggerInst
}

// StartTrace
func (l *PLogger) StartTrace() {
	l.isTraceEnable = true
}

// StopTrace
func (l *PLogger) StopTrace() {
	l.isTraceEnable = false
}

// Trace Log
func (l *PLogger) Trace(format string, v ...interface{}) {
	if !l.isTraceEnable {
		return
	}
	l.RawLogger().Output(2, fmt.Sprintf("[TRACE] "+format, v...))
}

// debug Log
func (l *PLogger) Debug(format string, v ...interface{}) {
	if l.logLevel > DEBUG {
		return
	}
	l.RawLogger().Output(2, fmt.Sprintf("[DEBUG] "+format, v...))
}

// Info Log
func (l *PLogger) Info(format string, v ...interface{}) {
	if l.logLevel > INFO {
		return
	}
	l.RawLogger().Output(2, fmt.Sprintf("[INFO] "+format, v...))
}

// Warn Log
func (l *PLogger) Warn(format string, v ...interface{}) {
	if l.logLevel > WARN {
		return
	}
	l.RawLogger().Output(2, fmt.Sprintf("[WARN] "+format, v...))
}

// Error Log
func (l *PLogger) Error(format string, v ...interface{}) {
	if l.logLevel > ERROR {
		return
	}
	l.RawLogger().Output(2, fmt.Sprintf("[ERROR] "+format, v...))
}

// Panic
func (l *PLogger) Panic(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	l.RawLogger().Output(2, "[PANIC] "+s)
	panic(s)
}

// Fatal
func (l *PLogger) Fatal(format string, v ...interface{}) {
	l.RawLogger().Output(2, fmt.Sprintf("[FATAL] "+format, v...))
	os.Exit(1)
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// ==== ==== ==== ==== PLogger: Default Logger ==== ==== ==== ====
const (
	defaultLogPath     = "./logs/app.log" // Default log path: ${app_path}/logs/app.log
	defaultRotateCount = 7                // Default rotate file count.
	defaultTraceOn     = true             // Default trace log flag is ON.
)

var (
	defaultLogger, _ = getLogger1(defaultLogPath, DEBUG, Daily, defaultRotateCount, defaultTraceOn)
)

// Trace Log
func Trace(format string, v ...interface{}) {
	if !defaultLogger.isTraceEnable {
		return
	}
	defaultLogger.RawLogger().Output(2, fmt.Sprintf("[TRACE] "+format, v...))
}

// Debug Log
func Debug(format string, v ...interface{}) {
	if defaultLogger.logLevel > DEBUG {
		return
	}
	defaultLogger.RawLogger().Output(2, fmt.Sprintf("[DEBUG] "+format, v...))
}

// Info Log
func Info(format string, v ...interface{}) {
	if defaultLogger.logLevel > INFO {
		return
	}
	defaultLogger.RawLogger().Output(2, fmt.Sprintf("[INFO] "+format, v...))
}

// Warn Log
func Warn(format string, v ...interface{}) {
	if defaultLogger.logLevel > WARN {
		return
	}
	defaultLogger.RawLogger().Output(2, fmt.Sprintf("[WARN] "+format, v...))
}

// Error Log
func Error(format string, v ...interface{}) {
	if defaultLogger.logLevel > ERROR {
		return
	}
	defaultLogger.RawLogger().Output(2, fmt.Sprintf("[ERROR] "+format, v...))
}

// Panic
func Panic(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	defaultLogger.RawLogger().Output(2, "[PANIC] "+s)
	panic(s)
}

// Fatal
func Fatal(format string, v ...interface{}) {
	defaultLogger.RawLogger().Output(2, fmt.Sprintf("[FATAL] "+format, v...))
	os.Exit(1)
}
