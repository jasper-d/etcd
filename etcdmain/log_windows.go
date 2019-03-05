package etcdmain

import (
	"fmt"
	"github.com/coreos/pkg/capnslog"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sys/windows/svc/eventlog"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
)

type eventLogWriter struct {
	elog *eventlog.Log
}

func (log *eventLogWriter) Open(src string) (writer *eventLogWriter, err error) {
	log.elog, err = eventlog.Open(src)
	return log, err
}

func (log *eventLogWriter) Close() error{
	return log.elog.Close()
}

func (log *eventLogWriter) Write(p []byte) (n int, err error){
	str := string(p)
	err = log.elog.Info(0, str)
	return len(p), err
}

func (log *eventLogWriter) Sync() (err error){
	return nil
}

func getFileName(logFile string) (string, error){
	if logFile == "" {
		return "", fmt.Errorf("empty")
	}

	fi, err := os.Stat(logFile)

	if fi != nil && fi.IsDir() {
		logFile = path.Join(logFile, "log.txt")
	}

	logFile, err = filepath.Abs(logFile)

	if err != nil {
		return "", err
	}

	return logFile, nil
}

func redirectToFile(logFile string) io.Closer {
	logFile, err := getFileName(logFile)

	if err == nil {
		ll := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    100, //100 MiB
			MaxBackups: 3,
			MaxAge:     28, //days
		}
		setLogOutput(ll)

		return ll
	}

	return nil
}

func redirectToSysLog(eventSource string) io.Closer {
	_ = eventlog.InstallAsEventCreate(eventSource, eventlog.Info | eventlog.Warning | eventlog.Error)
	elw :=  &eventLogWriter{}
	elw.elog, _ = eventlog.Open(eventSource)

	_ = elw.elog.Info(1, fmt.Sprintf("Opened log with name %s", eventSource))

	setLogOutput(elw)

	return elw
}

func redirectLog(eventSource string, logFile string, log io.Closer) io.Closer {
	switch log.(type) {
	case *lumberjack.Logger:
		return log
	}

	// close the log, we want to create a new one with the latest config
	if log != nil {
		log.Close()
		log = nil
	}

	log = redirectToFile(logFile)

	if log != nil || eventSource == "" {
		return log
	}

	return redirectToSysLog(eventSource)
}

func setLogOutput(writer io.Writer){
	// todo: i really don't know which log to use here...
	//       we still dont capture grpclog, tcpproxy.TCPProxy.logger (it uses zap), raft logger
	//       and possibly others though and would need to setup them as well :(
	log.SetOutput(writer)
	zapcore.AddSync(writer)
	capnslog.SetFormatter(capnslog.NewPrettyFormatter(writer, false))
}