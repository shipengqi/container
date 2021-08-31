package action

import (
	"github.com/pkg/errors"

	"github.com/shipengqi/container/internal/network"
)

type NetworkCreateActionOptions struct {
	Subnet string
	Driver string
}

type networkCreateA struct {
	*action

	networkName string
	options     *NetworkCreateActionOptions
}

type networkRemoveA struct {
	*action

	networkName string
}

type networkListA struct {
	*action
}

func NewNetworkCreateAction(name string, options *NetworkCreateActionOptions) Interface {
	return &networkCreateA{
		action: &action{
			name: "network-create",
		},
		options:     options,
		networkName: name,
	}
}

func NewNetworkRemoveAction(name string) Interface {
	return &networkRemoveA{
		action: &action{
			name: "network-rm",
		},
		networkName: name,
	}
}

func NewNetworkListAction() Interface {
	return &networkListA{
		action: &action{
			name: "network-ls",
		},
	}
}

func (a *networkCreateA) Run() error {
	err := network.Init()
	if err != nil {
		return err
	}
	err = network.Create(a.options.Driver, a.networkName, a.options.Subnet)
	if err != nil {
		return errors.Wrap(err, "create network")
	}
	return nil
}

func (a *networkRemoveA) Run() error {
	err := network.Init()
	if err != nil {
		return err
	}
	err = network.Delete(a.networkName)
	if err != nil {
		return errors.Wrap(err, "remove network")
	}
	return nil
}

func (a *networkListA) Run() error {
	err := network.Init()
	if err != nil {
		return err
	}
	network.List()
	return nil
}
