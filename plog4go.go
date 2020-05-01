// Copyright 2020 @thiinbit(thiinbit@gmail.com). All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package plog4go

import (
	"io"
	"log"
	"os"
	"sync"
)

var mu sync.Mutex

// loggerMap exist logger cache
var loggerMap = make(map[string]*PLogger)

type PLogger struct {
	logLogger *log.Logger
}

var defaultLogger = GetLogger(DefaultLogFile)

// GetLogger return exist logger or init.
func GetLogger(fileName string) *PLogger {
	mu.Lock()
	defer mu.Unlock()

	if val, ok := loggerMap[fileName]; ok {
		return val
	}

	logLogger := initLogger(fileName)

	return &PLogger{logLogger: logLogger}
}

// Printf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func Printf(format string, v ...interface{}) { defaultLogger.logLogger.Printf(format, v...) }

// Print calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func Print(v ...interface{}) { defaultLogger.logLogger.Print(v...) }

// Println calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func Println(v ...interface{}) { defaultLogger.logLogger.Println(v...) }

// Fatal is equivalent to l.Print() followed by a call to os.Exit(1).
func Fatal(v ...interface{}) { defaultLogger.logLogger.Fatal(v...) }

// Fatalf is equivalent to l.Printf() followed by a call to os.Exit(1).
func Fatalf(format string, v ...interface{}) { defaultLogger.logLogger.Fatalf(format, v...) }

// Fatalln is equivalent to l.Println() followed by a call to os.Exit(1).
func Fatalln(v ...interface{}) { defaultLogger.logLogger.Fatalln(v...) }

// Panic is equivalent to l.Print() followed by a call to panic().
func Panic(v ...interface{}) { defaultLogger.logLogger.Panic(v...) }

// Panicf is equivalent to l.Printf() followed by a call to panic().
func Panicf(format string, v ...interface{}) { defaultLogger.logLogger.Panicf(format, v...) }

// Panicln is equivalent to l.Println() followed by a call to panic().
func Panicln(v ...interface{}) { defaultLogger.logLogger.Panicln(v...) }

// Printf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *PLogger) Printf(format string, v ...interface{}) { l.logLogger.Printf(format, v...) }

// Print calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *PLogger) Print(v ...interface{}) { l.logLogger.Print(v...) }

// Println calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *PLogger) Println(v ...interface{}) { l.logLogger.Println(v...) }

// Fatal is equivalent to l.Print() followed by a call to os.Exit(1).
func (l *PLogger) Fatal(v ...interface{}) { l.logLogger.Fatal(v...) }

// Fatalf is equivalent to l.Printf() followed by a call to os.Exit(1).
func (l *PLogger) Fatalf(format string, v ...interface{}) { l.logLogger.Fatalf(format, v...) }

// Fatalln is equivalent to l.Println() followed by a call to os.Exit(1).
func (l *PLogger) Fatalln(v ...interface{}) { l.logLogger.Fatalln(v...) }

// Panic is equivalent to l.Print() followed by a call to panic().
func (l *PLogger) Panic(v ...interface{}) { l.logLogger.Panic(v...) }

// Panicf is equivalent to l.Printf() followed by a call to panic().
func (l *PLogger) Panicf(format string, v ...interface{}) { l.logLogger.Panicf(format, v...) }

// Panicln is equivalent to l.Println() followed by a call to panic().
func (l *PLogger) Panicln(v ...interface{}) { l.logLogger.Panicln(v...) }

func initLogger(fileName string) *log.Logger {
	if err := mkRootDir(); err != nil {
		return log.New(os.Stdout, "", DefaultLogFlag)
	}

	f, err := openFile(fileName)
	if err != nil {
		return log.New(os.Stdout, "", DefaultLogFlag)
	}
	return log.New(io.MultiWriter(os.Stdout, f), "", DefaultLogFlag)
}

func mkRootDir() error {
	if _, err := os.Stat(DefaultLogRootPath); os.IsNotExist(err) {
		err := os.MkdirAll(DefaultLogRootPath, os.ModePerm)
		if err != nil {
			log.Panic(err)
			return err
		}
	}
	return nil
}

func openFile(fileName string) (*os.File, error) {

	filePath := DefaultLogRootPath + string(os.PathSeparator) + fileName

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return f, nil
}
