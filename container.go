package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/shipengqi/container/cmd/c"
	"github.com/shipengqi/container/pkg/log"
)

func main() {
	cmd := c.New()
	cobra.OnInitialize(func() {
		_, err := log.Configure(log.Config{
			FileLevel:      "debug",
			Directory:      "/var/run/q.container/log",
			Filename:       "c.log",
		})
		if err != nil {
			panic(err)
		}
	})

	err := cmd.Execute()
	if err != nil {
		log.Errorf("%s.Execute: %v", cmd.Name(), err)
		os.Exit(1)
	}
	os.Exit(0)
}
