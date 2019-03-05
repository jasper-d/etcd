// +build !windows

package etcdmain

import (
	"github.com/coreos/go-systemd/daemon"
	systemdutil "github.com/coreos/go-systemd/util"
	"go.uber.org/zap"
	"log"
)

type EctdService interface {
	Init() (*zap.Logger, error)
	Start() error
	Stop() error
}

func Run(service *abstractAdapter) error {
	lg, _, err := service.Init()

	if err != nil {
		plog.Errorf("failed to init service, %v", err)
		return err
	}

	readyc, errorc := service.GetChannels()
	err = service.Start()

	go service.Start()
startLoop:
	for {
		select {
		case _ = <-readyc:
			break startLoop
		case e := <-errorc:
			service.Close()
			log.Fatalf("error during start: %v", e)
		}
	}

	notifySystemd(lg)

	for {
		select {
		case e := <-errorc:
			return e
		}
	}
}

func notifySystemd(lg *zap.Logger) {
	if !systemdutil.IsRunningSystemd() {
		return
	}

	if lg != nil {
		lg.Info("host was booted with systemd, sends READY=1 message to init daemon")
	}

	sent, err := daemon.SdNotify(false, "READY=1")
	if err != nil {
		if lg != nil {
			lg.Error("failed to notify systemd for readiness", zap.Error(err))
		} else {
			plog.Errorf("failed to notify systemd for readiness: %v", err)
		}
	}

	if !sent {
		if lg != nil {
			lg.Warn("forgot to set Type=notify in systemd service file?")
		} else {
			plog.Errorf("forgot to set Type=notify in systemd service file?")
		}
	}
}

func redirectLog(eventSource string, logFile string, log io.Closer) io.Closer{
	// noop on when not on windows
	return nil
}