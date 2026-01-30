package pprof

import (
	"context"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func ListenSignals(ctx context.Context, logger *zap.SugaredLogger, cpuFilename string, cpuDuration time.Duration, memFilename string) {
	channel := make(chan os.Signal, 1)
	defer close(channel)
	signal.Notify(channel, syscall.SIGUSR1, syscall.SIGUSR2)

	for {
		select {
		case signal := <-channel:
			switch signal {
			case syscall.SIGUSR1:
				go func() {
					logger.Info(
						"SIGUSR1 received. starting CPU profile capture...",
						zap.String("filename", cpuFilename),
						zap.Duration("duration", cpuDuration),
					)
					if err := CPUCapture(ctx, cpuFilename, cpuDuration); err != nil {
						logger.Error("cpu profile capture failed", zap.Error(err))
						return
					}
					logger.Info("cpu profile capture finished")
				}()
			case syscall.SIGUSR2:
				go func() {
					logger.Info("SIGUSR2 received. starting memory profile capture...", zap.String("filename", memFilename))
					if err := Capture(Heap, memFilename); err != nil {
						logger.Error("mem capture failed", zap.Error(err))
						return
					}
					logger.Info("memory profile capture finished")
				}()
			}
		case <-ctx.Done():
			return
		}
	}
}
