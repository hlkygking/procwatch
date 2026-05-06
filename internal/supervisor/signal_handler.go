package supervisor

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// SignalHandler listens for OS signals and triggers graceful shutdown.
type SignalHandler struct {
	logger  *Logger
	cancel  context.CancelFunc
	signals chan os.Signal
	done    chan struct{}
}

// NewSignalHandler creates a SignalHandler that will call cancel on SIGINT or SIGTERM.
func NewSignalHandler(logger *Logger, cancel context.CancelFunc) *SignalHandler {
	return &SignalHandler{
		logger:  logger,
		cancel:  cancel,
		signals: make(chan os.Signal, 1),
		done:    make(chan struct{}),
	}
}

// Start begins listening for termination signals in a background goroutine.
func (sh *SignalHandler) Start() {
	signal.Notify(sh.signals, syscall.SIGINT, syscall.SIGTERM)
	go sh.run()
}

// Stop unregisters signal notifications and waits for the handler goroutine to exit.
func (sh *SignalHandler) Stop() {
	signal.Stop(sh.signals)
	close(sh.signals)
	<-sh.done
}

func (sh *SignalHandler) run() {
	defer close(sh.done)
	for sig := range sh.signals {
		sh.logger.Info("received signal, initiating shutdown", map[string]any{
			"signal": sig.String(),
		})
		sh.cancel()
		return
	}
}
