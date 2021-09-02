package main

import (
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/shipengqi/container/cmd/c"
	"github.com/shipengqi/container/pkg/log"
)

func main() {
	cmd := c.New()
	cobra.OnInitialize(func() {
		_, err := log.Configure(createLogConfig())
		if err != nil {
			panic(err)
		}
	})
	// If the command returns an error, cli takes upon itself to print
	// the error on cmd.ErrOrStderr and exit.
	// Use our own writer here to ensure the log gets sent to the right location.
	cmd.SetErr(&FatalWriter{cmd.ErrOrStderr()})
	err := cmd.Execute()
	if err != nil {
		log.Errorf("%s.Execute: %v", cmd.Name(), err)
		os.Exit(1)
	}
	os.Exit(0)
}

type FatalWriter struct {
	cliErrWriter io.Writer
}

func (f *FatalWriter) Write(p []byte) (n int, err error) {
	log.Error(string(p))
	return f.cliErrWriter.Write(p)
}

func createLogConfig() log.Config {
	return log.Config{
		FileLevel: "debug",
		Directory: "/var/run/q.container/log",
		Filename:  "c.log",
	}
}
