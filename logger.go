package p_log4go

import (
	"fmt"
	"github.com/thiinbit/p-log4go/file"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
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

// timedRotatingWriter
type timedRotatingWriter struct {
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
func newTimedRotateWriter(filename string, interval RotateInterval, rotate int64) (*timedRotatingWriter, error) {
	w := &timedRotatingWriter{
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
func (w *timedRotatingWriter) initialize() error {
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
func (w *timedRotatingWriter) tryRotate() (err error) {
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

func (w *timedRotatingWriter) Write(output []byte) (int, error) {
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

// These flags define which text to prefix to each log entry generated by the Logger.
// Bits are or'ed together to control what's printed.
// With the exception of the Lmsgprefix flag, there is no
// control over the order they appear (the order listed here)
// or the format they present (as described in the comments).
// The prefix is followed by a colon only when Llongfile or Lshortfile
// is specified.
// For example, flags Ldate | Ltime (or LstdFlags) produce,
//	2009/01/23 01:23:23 message
// while flags Ldate | Ltime | Lmicroseconds | Llongfile produce,
//	2009/01/23 01:23:23.123123 /a/b/c/d.go:23: message
const (
	Ldate         = 1 << iota     // the date in the local time zone: 2009/01/23
	Ltime                         // the time in the local time zone: 01:23:23
	Lmicroseconds                 // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                     // full file name and line number: /a/b/c/d.go:23
	Lshortfile                    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                          // if Ldate or Ltime is set, use UTC rather than the local time zone
	Lmsgprefix                    // move the "prefix" from the beginning of the line to before the message
	LstdFlags     = Ldate | Ltime // initial values for the standard logger
)

// Appender
type Appender int8

// Log appender ,
// - output to file `File`
// - output to console `Console`
// - output to file and console `File | Console`
const (
	ConsoleAppender = 1 << iota
	FileAppender
)

// Log level DEBUG/INFO/WARN/ERROR
type LogLevel int8

// Log level iota
const (
	trace LogLevel = iota
	DEBUG
	INFO
	WARN
	ERROR
	PANIC
	FATAL
)

type PLogger struct {
	//loggerInst    *log.Logger
	logLevel      LogLevel // Loglevel DEBUG INFO WARN ERROR
	isTraceEnable bool     // Is trace enable
	// log.logger
	mu     sync.Mutex // ensures atomic writes; protects the following fields
	prefix string     // prefix on each line to identify the logger (but see Lmsgprefix)
	flag   int        // properties
	out    io.Writer  // destination for output
	buf    []byte     // for accumulating text to write
}

func GetLogger(filePath string, logLevel LogLevel, interval RotateInterval, rotate int64) (*PLogger, error) {

	return GetLogger2(filePath, logLevel, interval, rotate, false, FileAppender)
}

func GetLogger0(filepath string) (*PLogger, error) {
	return GetLogger2(filepath, DEBUG, Daily, defaultRotateCount, defaultTraceOn, FileAppender)
}

func GetLogger1(filePath string, logLevel LogLevel, interval RotateInterval, rotate int64, appender Appender) (*PLogger, error) {

	return GetLogger2(filePath, logLevel, interval, rotate, false, appender)
}

func GetLogger2(filePath string, logLevel LogLevel, interval RotateInterval, rotate int64, traceOn bool, appender Appender) (*PLogger, error) {

	fileDir := filepath.Dir(filePath)
	exist, err := pathExists(fileDir)
	if err != nil {
		return nil, fmt.Errorf("log file path err, %v", err)
	}

	if !exist {
		if err = os.MkdirAll(fileDir, 0755); err != nil {
			return nil, fmt.Errorf("mkdir logfile dir err, %v", err)
		}
	}

	// TODO: Multi writer
	var writers = make([]io.Writer, 0)

	if appender&FileAppender != 0 {
		var fileWriter *timedRotatingWriter
		fileWriter, err = newTimedRotateWriter(filePath, interval, rotate)
		if err != nil {
			return nil, fmt.Errorf("create RotateRiter err, %v", err)
		}
		writers = append(writers, fileWriter)
	}

	if appender&ConsoleAppender != 0 {
		writers = append(writers, os.Stdout)
	}

	return &PLogger{
		logLevel:      logLevel,
		isTraceEnable: traceOn,
		prefix:        "",
		flag:          Ldate | Ltime | Lmicroseconds | Lshortfile,
		out:           io.MultiWriter(writers...),
	}, nil
}

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

// formatHeader writes log header to buf in following order:
//   * l.prefix (if it's not blank and Lmsgprefix is unset),
//   * date and/or time (if corresponding flags are provided),
//   * file and line number (if corresponding flags are provided),
//   * l.prefix (if it's not blank and Lmsgprefix is set).
func (l *PLogger) formatHeader(buf *[]byte, level LogLevel, t time.Time, file string, line int) {
	// Log level
	switch level {
	case trace:
		*buf = append(*buf, "[TRACE] "...)
	case DEBUG:
		*buf = append(*buf, "[DEBUG] "...)
	case INFO:
		*buf = append(*buf, "[INFO] "...)
	case WARN:
		*buf = append(*buf, "[WARN] "...)
	case ERROR:
		*buf = append(*buf, "[ERROR] "...)
	case PANIC:
		*buf = append(*buf, "[PANIC] "...)
	case FATAL:
		*buf = append(*buf, "[FATAL] "...)
	}

	// Log msg prefix false
	if l.flag&Lmsgprefix == 0 {
		*buf = append(*buf, l.prefix...)
	}

	// Log date time microseconds
	if l.flag&(Ldate|Ltime|Lmicroseconds) != 0 {
		if l.flag&LUTC != 0 {
			t = t.UTC()
		}
		if l.flag&Ldate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '/')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '/')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if l.flag&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if l.flag&Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}

	// Log short file | long file
	if l.flag&(Lshortfile|Llongfile) != 0 {
		if l.flag&Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
		*buf = append(*buf, ": "...)
	}

	// Log msg prefix true
	if l.flag&Lmsgprefix != 0 {
		*buf = append(*buf, l.prefix...)
	}
}

// Output writes the output for a logging event. The string s contains
// the text to print after the prefix specified by the flags of the
// Logger. A newline is appended if the last character of s is not
// already a newline. Calldepth is used to recover the PC and is
// provided for generality, although at the moment on all pre-defined
// paths it will be 2.
func (l *PLogger) Output(calldepth int, logLevel LogLevel, s string) error {
	now := time.Now() // get this early.
	var file string
	var line int
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.flag&(Lshortfile|Llongfile) != 0 {
		// Release lock while getting caller info - it's expensive.
		l.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		l.mu.Lock()
	}
	l.buf = l.buf[:0]
	l.formatHeader(&l.buf, logLevel, now, file, line)
	l.buf = append(l.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	_, err := l.out.Write(l.buf)
	return err
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
	l.Output(2, trace, fmt.Sprintf(format, v...))
}

// debug Log
func (l *PLogger) Debug(format string, v ...interface{}) {
	if l.logLevel > DEBUG {
		return
	}
	l.Output(2, DEBUG, fmt.Sprintf(format, v...))
}

// Info Log
func (l *PLogger) Info(format string, v ...interface{}) {
	if l.logLevel > INFO {
		return
	}
	l.Output(2, INFO, fmt.Sprintf(format, v...))
}

// Warn Log
func (l *PLogger) Warn(format string, v ...interface{}) {
	if l.logLevel > WARN {
		return
	}
	l.Output(2, WARN, fmt.Sprintf(format, v...))
}

// Error Log
func (l *PLogger) Error(format string, v ...interface{}) {
	if l.logLevel > ERROR {
		return
	}
	l.Output(2, ERROR, fmt.Sprintf(format, v...))
}

func (l *PLogger) Panic(format string, v ...interface{}) {
	if l.logLevel > PANIC {
		return
	}
	s := fmt.Sprintf(format, v...)
	l.Output(2, PANIC, s)
	panic(s)
}

// Fatal
func (l *PLogger) Fatal(format string, v ...interface{}) {
	if l.logLevel > FATAL {
		return
	}
	l.Output(2, FATAL, fmt.Sprintf(format, v...))
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
	defaultLogger, _ = GetLogger0(defaultLogPath)
)

// Trace Log
func Trace(format string, v ...interface{}) {
	if !defaultLogger.isTraceEnable {
		return
	}
	defaultLogger.Output(2, trace, fmt.Sprintf(format, v...))
}

// Debug Log
func Debug(format string, v ...interface{}) {
	if defaultLogger.logLevel > DEBUG {
		return
	}
	defaultLogger.Output(2, DEBUG, fmt.Sprintf(format, v...))
}

// Info Log
func Info(format string, v ...interface{}) {
	if defaultLogger.logLevel > INFO {
		return
	}
	defaultLogger.Output(2, INFO, fmt.Sprintf(format, v...))
}

// Warn Log
func Warn(format string, v ...interface{}) {
	if defaultLogger.logLevel > WARN {
		return
	}
	defaultLogger.Output(2, WARN, fmt.Sprintf(format, v...))
}

// Error Log
func Error(format string, v ...interface{}) {
	if defaultLogger.logLevel > ERROR {
		return
	}
	defaultLogger.Output(2, ERROR, fmt.Sprintf(format, v...))
}

// Panic
func Panic(format string, v ...interface{}) {
	if defaultLogger.logLevel > PANIC {
		return
	}
	s := fmt.Sprintf(format, v...)
	defaultLogger.Output(2, PANIC, s)
	panic(s)
}

// Fatal
func Fatal(format string, v ...interface{}) {
	if defaultLogger.logLevel > FATAL {
		return
	}
	defaultLogger.Output(2, FATAL, fmt.Sprintf(format, v...))
	os.Exit(1)
}
