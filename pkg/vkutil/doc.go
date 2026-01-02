/*
Package vkutil provides a wrapper around the valkey-go client for interacting with Valkey/Redis.

It integrates with the launcher package to provide a managed lifecycle (initialization,
startup verification, and graceful shutdown) for the Valkey client.

Features:
- Managed lifecycle via launcher.Component.
- Configurable connection through Environment variables.
- Support for client-side caching.
- Standardized logging of connection events.

Example usage:

	cfg := &vkutil.Config{Addrs: []string{"localhost:6379"}}
	vk := vkutil.New(cfg, logger)

	if err := vk.OnInit(); err != nil {
		logger.Fatal("vkutil: fallo al inicializar Valkey", err)
	}

	if err := vk.OnStart(); err != nil {
		logger.Fatal("vkutil: fallo al conectar con Valkey", err)
	}

	err := vk.Client().Do(ctx, vk.Client().B().Set().Key("foo").Value("bar").Build()).Error()
*/
package vkutil
