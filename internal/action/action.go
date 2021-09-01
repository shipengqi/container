package action

import (
	"strings"

	jsoniter "github.com/json-iterator/go"

	"github.com/shipengqi/container/pkg/log"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Interface interface {
	Name() string
	PreRun() error
	Run() error
	PostRun() error
}

type action struct {
	name    string
}

func (a *action) Name() string {
	return "[action]"
}

func (a *action) PreRun() error {
	log.Debugf("***** [%s] PreRun *****", strings.ToUpper(a.name))
	return nil
}

func (a *action) Run() error {
	log.Debugf("***** [%s] Run *****", strings.ToUpper(a.name))
	return nil
}

func (a *action) PostRun() error {
	log.Debugf("***** [%s] PostRun *****", strings.ToUpper(a.name))
	return nil
}

func Execute(a Interface) error {
	err := a.PreRun()
	if err != nil {
		return err
	}
	err = a.Run()
	if err != nil {
		return err
	}
	return a.PostRun()
}
