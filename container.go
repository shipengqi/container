package main

import (
	"os"

	"github.com/shipengqi/container/cmd/c"
	"github.com/shipengqi/container/pkg/log"
)

func main() {
	cmd := c.New()
	// cobra.OnInitialize(func() {
	// 	_, err := log.Configure(log.Config{
	// 		FileEnabled:    true,
	// 		FileJson:       true,
	// 		ConsoleEnabled: true,
	// 		ConsoleJson:    true,
	// 		FileLevel:      "debug",
	// 		ConsoleLevel:   "debug",
	// 		Directory:      "/var/run/q.container/log",
	// 		Filename:       "c.log",
	// 		MaxSize:        100,
	// 		MaxBackups:     7,
	// 		MaxAge:         7,
	// 	})
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// })

	err := cmd.Execute()
	if err != nil {
		log.Errorf("%s.Execute(): %v", cmd.Name(), err)
		os.Exit(1)
	}
	os.Exit(0)
}
