package c

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/internal/images"
	"github.com/shipengqi/container/pkg/log"
)


func newCommitCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "commit [options]",
		Short:   "commit a container into image.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("missing container name")
			}
			log.Info("committing")
			images.CommitContainer(args[0])
			return nil
		},
	}
	c.DisableFlagsInUseLine = true
	return c
}

// [root@shcCDFrh75vm7 container]# ./container commit image
// 2021-08-21T17:04:39.054+0800	INFO	committing
// 2021-08-21T17:04:39.054+0800	DEBUG	commit: /root/image.tar
// [root@shcCDFrh75vm7 container]# cd ..
// [root@shcCDFrh75vm7 ~]# mkdir untar
// [root@shcCDFrh75vm7 ~]# cd  untar/
// [root@shcCDFrh75vm7 untar]# tar -xvf /root/image.tar
// [root@shcCDFrh75vm7 untar]# ls
// bin  dev  etc  home  lib  lib64  proc  root  sys  tmp  usr  var  version.txt