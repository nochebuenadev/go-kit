package vkutil

import (
	"context"
	"fmt"

	"github.com/nochebuenadev/go-kit/pkg/health"
	"github.com/nochebuenadev/go-kit/pkg/launcher"
	"github.com/nochebuenadev/go-kit/pkg/logz"
	"github.com/valkey-io/valkey-go"
)

type (
	// ValkeyProvider defines the interface for accessing the Valkey client.
	ValkeyProvider interface {
		// Client returns the underlying valkey.Client instance.
		Client() valkey.Client
	}

	// ValkeyComponent extends ValkeyProvider with lifecycle management methods.
	ValkeyComponent interface {
		launcher.Component
		ValkeyProvider
	}

	// vkComponent is the concrete implementation of ValkeyComponent.
	vkComponent struct {
		// client is the underlying valkey-go client.
		client valkey.Client
		// cfg is the valkey connection configuration.
		cfg *Config
		// logger is used for tracking valkey operations.
		logger logz.Logger
	}
)

// New creates a new ValkeyComponent instance.
func New(cfg *Config, logger logz.Logger) ValkeyComponent {
	return &vkComponent{
		cfg:    cfg,
		logger: logger,
	}
}

// OnInit initializes the Valkey client with the provided configuration.
func (v *vkComponent) OnInit() error {
	opts := valkey.ClientOption{
		InitAddress: v.cfg.Addrs,
		Password:    v.cfg.Password,
		SelectDB:    v.cfg.SelectDB,
	}

	if v.cfg.CacheSizeEachConn > 0 {
		opts.DisableCache = false
		opts.CacheSizeEachConn = v.cfg.CacheSizeEachConn * 1024 * 1024
	}

	client, err := valkey.NewClient(opts)
	if err != nil {
		return fmt.Errorf("vkutil: error creando cliente: %w", err)
	}

	v.client = client
	return nil
}

// OnStart verifies the connection to Valkey by sending a PING command.
func (v *vkComponent) OnStart() error {
	v.logger.Info("vkutil: Verificando conexión con Valkey...")

	if v.client == nil {
		return fmt.Errorf("vkutil: cliente no inicializado")
	}

	// Un simple PING para asegurar que el componente está arriba
	err := v.client.Do(context.Background(), v.client.B().Ping().Build()).Error()
	if err != nil {
		return fmt.Errorf("vkutil: no se pudo conectar a Valkey: %w", err)
	}

	v.logger.Info("vkutil: Conexión establecida correctamente")
	return nil
}

// OnStop closes the Valkey client connections.
func (v *vkComponent) OnStop() error {
	v.logger.Info("vkutil: Cerrando cliente de Valkey...")
	if v.client != nil {
		v.client.Close()
	}
	return nil
}

// Client returns the underlying Valkey client.
func (v *vkComponent) Client() valkey.Client {
	return v.client
}

// HealthCheck implements the health.Checkable interface.
func (v *vkComponent) HealthCheck(ctx context.Context) error {
	return v.client.Do(ctx, v.client.B().Ping().Build()).Error()
}

// Name implements the health.Checkable interface.
func (v *vkComponent) Name() string { return "valkey" }

// Priority implements the health.Checkable interface.
func (v *vkComponent) Priority() health.Level { return health.LevelDegraded }
