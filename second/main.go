package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"

	"golang.org/x/sync/errgroup"
)

func main() {
	ports := []string{":8081", "8082", "8083"}
	ctx, cancel := context.WithCancel(context.Background())
	eg, eCtx := errgroup.WithContext(ctx)
	wg := &sync.WaitGroup{}
	for _, port := range ports {
		eg.Go(func() error {
			<-eCtx.Done()
			return ShutDown(eCtx)
		})
		wg.Add(1)
		eg.Go(func() error {
			defer wg.Done()
			return Start(port)
		})
	}
	wg.Wait()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	eg.Go(func() error {
		for {
			select {
			case <-eCtx.Done():
				return eCtx.Err()
			case <-c:
				// TODO:handler
				cancel()
			}
		}
	})
	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		logrus.Errorf("server wait fail:%v", err)
	}
}

func Start(port string) error {
	// TODO:add http handler
	return http.ListenAndServe(port, nil)
}

func ShutDown(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
