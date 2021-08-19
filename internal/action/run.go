package action

import (
	"go.uber.org/zap"
	"strings"

	"github.com/shipengqi/container/internal/container"
	"github.com/shipengqi/container/pkg/log"
)

type run struct {
	*action
}

func NewRunAction(command string) Interface {
	return &run{&action{
		name:    "run",
		command: command,
	}}
}

func (r *run) Name() string {
	return r.name
}

func (r *run) Run() error {
	log.Debugf("***** %s Run *****", strings.ToUpper(r.name))
	p := container.NewParentProcess(true, r.command)
	if err := p.Start(); err != nil {
		log.Errort("parent run", zap.Error(err))
	}
	err := p.Wait()
	if err != nil {
		log.Errort("parent wait", zap.Error(err))
		return err
	}
	return nil
}
