package cmd

import (
	"context"
	"fmt"

	"github.com/golang/glog"
)

func CheckError(err error) {
	if err != nil {
		if err != context.Canceled {
			glog.Exit(fmt.Errorf("An error occurred: %v", err))
		}
	}
}
