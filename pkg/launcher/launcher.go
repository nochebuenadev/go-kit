package launcher

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nochebuenadev/go-kit/pkg/logz"
)

type (
	// Hook defines a function to be executed during the assembly/DI phase.
	Hook func() error

	// Component defines the lifecycle interface for application components.
	Component interface {
		// OnInit initializes the component.
		OnInit() error
		// OnStart starts the component services.
		OnStart() error
		// OnStop stops the component services.
		OnStop() error
	}

	// Launcher defines the interface for managing the application lifecycle.
	Launcher interface {
		// Append adds one or more components to the launcher.
		Append(components ...Component)
		// BeforeStart registers hooks to be executed before starting the components.
		BeforeStart(hooks ...Hook)
		// Run starts the initialization, assembly, and component startup sequence.
		// It also handles OS signals for graceful shutdown.
		Run()
	}

	// launcher is the concrete implementation of the Launcher interface.
	launcher struct {
		// logger is used for tracking the lifecycle transitions.
		logger logz.Logger
		// components is the list of registered components to manage.
		components []Component
		// onBeforeStart is the list of hooks to execute before startup.
		onBeforeStart []Hook
	}
)

// New creates a new Launcher instance.
func New(logger logz.Logger) Launcher {
	return &launcher{
		logger:     logger,
		components: make([]Component, 0),
	}
}

// Append adds components to the launcher's registry.
func (l *launcher) Append(components ...Component) {
	l.components = append(l.components, components...)
}

// BeforeStart registers hooks to be executed before the OnStart phase.
func (l *launcher) BeforeStart(hooks ...Hook) {
	l.onBeforeStart = append(l.onBeforeStart, hooks...)
}

// Run executes the full application lifecycle:
// 1. OnInit for all components.
// 2. BeforeStart hooks (Assembly/DI).
// 3. OnStart for all components.
// 4. Waits for termination signal.
// 5. Graceful shutdown.
func (l *launcher) Run() {
	l.logger.Info("launcher: iniciando fase de inicialización (OnInit)")
	for _, c := range l.components {
		if err := c.OnInit(); err != nil {
			l.logger.Fatal("launcher: fallo crítico en OnInit", err)
		}
	}

	l.logger.Info("launcher: ejecutando ganchos de ensamblaje (DI)")
	for _, hook := range l.onBeforeStart {
		if err := hook(); err != nil {
			l.logger.Fatal("launcher: fallo crítico en el ensamblaje de dependencias", err)
		}
	}

	l.logger.Info("launcher: activando componentes (OnStart)")
	for _, c := range l.components {
		if err := c.OnStart(); err != nil {
			l.logger.Error("launcher: fallo en OnStart, iniciando apagado preventivo", err)
			l.shutdown()
			os.Exit(1)
		}
	}

	l.logger.Info("launcher: aplicación lista y operando")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	s := <-quit
	l.logger.Info("launcher: señal de terminación recibida", "signal", s.String())

	l.shutdown()
}

// shutdown stops all components in reverse order of their registration.
func (l *launcher) shutdown() {
	l.logger.Info("launcher: iniciando apagado controlado (Graceful Shutdown)")

	for i := len(l.components) - 1; i >= 0; i-- {
		done := make(chan struct{})
		go func(c Component) {
			if err := c.OnStop(); err != nil {
				l.logger.Error("launcher: error durante OnStop", err)
			}
			close(done)
		}(l.components[i])

		select {
		case <-done:
			continue
		case <-time.After(15 * time.Second):
			l.logger.Error("launcher: timeout alcanzado durante el OnStop de un componente", nil)
		}
	}
	l.logger.Info("launcher: sistema apagado correctamente")
}
