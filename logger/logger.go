package logger

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"sync"
	"time"
)

type Logger struct {
	file     *os.File
	debug    *log.Logger
	info     *log.Logger
	warn     *log.Logger
	err      *log.Logger
	fatal    *log.Logger
	dir      string
	filename string
	lock     sync.Mutex
	done     chan struct{}
}

func NewLogger(dir string, filename string) (*Logger, error) {
	filePath := path.Join(dir, fmt.Sprintf("%s_%s.log", filename, time.Now().Format("2006_01_02_15")))

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		file:     file,
		debug:    log.New(file, "[DEBUG] ", log.Ldate|log.Ltime),
		info:     log.New(file, "[INFO] ", log.Ldate|log.Ltime),
		warn:     log.New(file, "[WARN] ", log.Ldate|log.Ltime),
		err:      log.New(file, "[ERROR] ", log.Ldate|log.Ltime),
		fatal:    log.New(file, "[FATAL] ", log.Ldate|log.Ltime),
		filename: filename,
		dir:      dir,
		done:     make(chan struct{}),
	}, nil
}

func (l *Logger) Close() {
	close(l.done)
	l.file.Close()
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.rotateFile()
	_, file, line, ok := runtime.Caller(1)
	if ok {
		l.debug.Printf("%s:%d %s", file, line, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Debug(v ...interface{}) {
	l.rotateFile()
	_, file, line, ok := runtime.Caller(1)
	if ok {
		l.debug.Printf("%s:%d %s", file, line, fmt.Sprint(v...))
	}
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.rotateFile()
	_, file, line, ok := runtime.Caller(1)
	if ok {
		l.info.Printf("%s:%d %s", file, line, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Info(v ...interface{}) {
	l.rotateFile()
	_, file, line, ok := runtime.Caller(1)
	if ok {
		l.info.Printf("%s:%d %s", file, line, fmt.Sprint(v...))
	}
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.rotateFile()
	_, file, line, ok := runtime.Caller(1)
	if ok {
		l.warn.Printf("%s:%d %s", file, line, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Warn(v ...interface{}) {
	l.rotateFile()
	_, file, line, ok := runtime.Caller(1)
	if ok {
		l.warn.Printf("%s:%d %s", file, line, fmt.Sprint(v...))
	}
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.rotateFile()
	_, file, line, ok := runtime.Caller(1)
	if ok {
		l.err.Printf("%s:%d %s", file, line, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Error(v ...interface{}) {
	l.rotateFile()
	_, file, line, ok := runtime.Caller(1)
	if ok {
		l.err.Printf("%s:%d %s", file, line, fmt.Sprint(v...))
	}
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.rotateFile()
	_, file, line, ok := runtime.Caller(1)
	if ok {
		l.fatal.Printf("%s:%d %s", file, line, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Fatal(v ...interface{}) {
	l.rotateFile()
	_, file, line, ok := runtime.Caller(1)
	if ok {
		l.fatal.Printf("%s:%d %s", file, line, fmt.Sprint(v...))
	}
}

func (l *Logger) rotateFile() {
	l.lock.Lock()
	defer l.lock.Unlock()

	currentFileName := path.Join(l.dir, fmt.Sprintf("%s_%s.log", l.filename, time.Now().Format("2006_01_02_15")))
	if l.file.Name() != currentFileName {
		if err := l.file.Close(); err != nil {
			log.Fatalf("Error closing log file: %v", err)
		}
		file, err := os.OpenFile(currentFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("Error creating log file: %v", err)
		}
		l.file = file
		l.debug.SetOutput(file)
		l.info.SetOutput(file)
		l.warn.SetOutput(file)
		l.err.SetOutput(file)
		l.fatal.SetOutput(file)
	}
}
