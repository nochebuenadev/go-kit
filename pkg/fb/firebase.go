package fb

import (
	"context"
	"sync"

	firebase "firebase.google.com/go/v4"
	"github.com/nochebuenadev/go-kit/pkg/launcher"
	"github.com/nochebuenadev/go-kit/pkg/logz"
)

type (
	// Provider defines the interface for accessing the Firebase App.
	Provider interface {
		// App returns the underlying firebase.App instance.
		App() *firebase.App
	}

	// Component extends Provider with lifecycle management methods.
	Component interface {
		launcher.Component
		Provider
	}

	// firebaseComponent is the concrete implementation of the Firebase component.
	firebaseComponent struct {
		// cfg is the firebase configuration.
		cfg *Config
		// logger is used for reporting firebase events.
		logger logz.Logger
		// app is the underlying firebase app instance.
		app *firebase.App
	}
)

var (
	// fbInstance is the singleton firebase component.
	fbInstance Component
	// fbOnce ensures that the component is initialized only once.
	fbOnce sync.Once
)

// GetFirebase returns the singleton instance of the Firebase component.
func GetFirebase(logger logz.Logger, cfg *Config) Component {
	fbOnce.Do(func() {
		fbInstance = &firebaseComponent{
			cfg:    cfg,
			logger: logger,
		}
	})
	return fbInstance
}

// OnInit implements the launcher.Component interface to initialize the Firebase App.
func (f *firebaseComponent) OnInit() error {
	f.logger.Info("fb: inicializando aplicación...", "project_id", f.cfg.ProjectID)

	app, err := firebase.NewApp(context.Background(), &firebase.Config{
		ProjectID: f.cfg.ProjectID,
	})
	if err != nil {
		f.logger.Error("fb: error al crear la aplicación", err)
		return err
	}

	f.app = app
	return nil
}

// OnStart implements the launcher.Component interface.
func (f *firebaseComponent) OnStart() error {
	f.logger.Info("fb: motor de google activo")
	return nil
}

// OnStop implements the launcher.Component interface.
func (f *firebaseComponent) OnStop() error {
	f.logger.Info("fb: cerrando conexiones")
	return nil
}

// App returns the underlying firebase.App instance.
func (f *firebaseComponent) App() *firebase.App {
	return f.app
}
