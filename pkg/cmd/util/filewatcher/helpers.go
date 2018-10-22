package filewatcher

import (
	"context"
	"os"
	"time"
)

func WatchForFile(ctx context.Context, path string, secondsInterval int, flagFileErr chan bool) {
	for {
		select {
		case <-ctx.Done():
			flagFileErr <- false
			break
		default:
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				flagFileErr <- true
				break
			}
		}

		time.Sleep(time.Second * time.Duration(secondsInterval))
	}
}
