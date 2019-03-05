package etcdmain

import (
	"go.uber.org/zap"
	"io"
)

var svcAdapter = &serviceAdapter{
	readyc: make(chan struct{}),
	error:  make(chan error),
}

type serviceAdapter struct {
	cfg    *config
	stop   func() error
	readyc chan struct{}
	error  chan error
	EctdService
}

func (s *serviceAdapter) Init() (*zap.Logger, *config, error){
	lg, cfg, err := getConfig()
	svcAdapter.cfg = cfg
	return lg, cfg, err
}

func (s *serviceAdapter) Stop() error{
	return s.stop()
}

func (s *serviceAdapter) Start() error{
	err := start()

	if err != nil {
		return err
	}

	return nil
}

func (s *serviceAdapter) GetChannels() (<- chan struct{}, <- chan error){
	return s.readyc, s.error
}

type abstractAdapter struct {
	winCfg *windowsServiceConfig
	cmd *etcdCommand
	svc *serviceAdapter
	log io.Closer
	isInteractive bool
}

func (a *abstractAdapter) GetChannels() (<- chan struct{}, <- chan error){
	if a.cmd != nil {
		return a.cmd.GetChannels()
	} else {
		return a.svc.GetChannels()
	}
}

func (a *abstractAdapter) Init() (*zap.Logger, *config, error){
	var lg *zap.Logger
	var cfg *config
	var err error

	if a.cmd != nil {
		lg, cfg, err = a.cmd.Init()
	} else {
		lg, cfg, err = a.svc.Init()
	}

	if err == nil && !a.isInteractive {
		a.log = redirectLog(*cfg.Windows.eventSource, *cfg.Windows.logFile, a.log)
	}

	return lg, cfg, err
}

func (a *abstractAdapter) Start() error{
	var err error

	if a.cmd != nil {
		a.winCfg = a.cmd.cfg.Windows
		err = a.cmd.Start()
	} else {
		a.winCfg = a.svc.cfg.Windows
		err = a.svc.Start()
	}
	return err
}

func (a *abstractAdapter) Stop() error{
	if a.cmd != nil {
		return a.cmd.Stop()
	} else {
		return a.svc.Stop()
	}
}

func (a *abstractAdapter) Close() error{
	if a.log != nil {
		return a.log.Close()
	}

	return nil
}
