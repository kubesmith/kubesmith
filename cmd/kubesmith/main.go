package main

import (
	"os"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/kubesmith/kubesmith/pkg/cmd"
	"github.com/kubesmith/kubesmith/pkg/cmd/kubesmith"
)

func main() {
	defer glog.Flush()

	baseName := filepath.Base(os.Args[0])

	err := kubesmith.NewCommand(baseName).Execute()
	cmd.CheckError(err)
}
