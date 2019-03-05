package etcdmain

import (
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"log"
	"os"
	"path/filepath"
)

type EctdService interface {
	// initializes the services, setups the log using a best effort approach
	// and provides access to zap and the config
	Init() (*zap.Logger, *config, error)

	// starts the service and setups the log using the provided config
	Start() error

	// stops the service
	Stop() error

	// executes the service (implements golang.org/x/sys/windows/svc/Handler)
	Execute([]string, <-chan svc.ChangeRequest, chan<- svc.Status) (bool, uint32)
}

func (a *abstractAdapter) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted= svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	fmt.Println("Runnig execute")
	ready, errc := a.GetChannels()
	go a.Start()
startLoop:
	for {
		select {
		case _ = <-ready:
			break startLoop
		case e := <-errc:
			plog.Errorf("error during start: %v", e)
			errno = 1; ssec = true
			a.Close()
			return
		}
	}

	a.log = redirectLog(*a.winCfg.eventSource, *a.winCfg.logFile, a.log)
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Stop, svc.Shutdown:
				plog.Infof("received stop signal %d, stopping", errno)
				break loop
			}
		case e := <-errc:
			plog.Infof("received error %v, stopping", e)
			errno = 2; ssec = true
			break loop
		}
	}

	changes <- svc.Status{State: svc.StopPending}

	err := a.Stop()
	if err != nil {
		log.Printf("Error stopping service, %v", err)
	}

	_ = a.Close()
	changes <- svc.Status{State: svc.Stopped}
	return
}

func Run(service *abstractAdapter) error {
	var err error
	run := debug.Run

	service.isInteractive, err = svc.IsAnInteractiveSession()

	if err != nil {
		return err
	}

	if !service.isInteractive {
		fmt.Print("Running as service")

		run = svc.Run
		err = setCwd()

		if err != nil {
			os.Exit(1)
		}
	} else {
		fmt.Println("Running in interactive mode")
	}

	_, _, err = service.Init()

	if err != nil {
		log.Printf("Error during init service: %v", err)
	}


	if err != nil {
		_ = service.log.Close()
		log.Fatalf("Failed to initialize service: %v", err)
	}

	return run("etcd", service)
}

func setCwd() error{
	var exe, err = getExePath()

	if err != nil {
		return err
	}

	return os.Chdir(filepath.Dir(exe))
}

func getExePath() (string, error){
	exe, err := os.Executable()

	if err != nil {
		return "", err
	}

	return exe, nil
}