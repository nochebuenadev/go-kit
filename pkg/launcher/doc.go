/*
Package launcher provides a simple dependency injection and lifecycle management helper.

It automates the initialization, dependency assembly, and startup sequence of application
components, while also ensuring a graceful shutdown by handling OS interruption signals.

Components registered with the Launcher must implement the Component interface, which
defines methods for initialization (OnInit), startup (OnStart), and shutdown (OnStop).

The launcher follows a strict lifecycle:
1. OnInit: Synchronously initializes all components in the order they were appended.
2. BeforeStart: Executes assembly hooks (useful for manual DI or late binding).
3. OnStart: Synchronously starts all components.
4. Signal Handling: Waits for SIGINT (Interrupt) or SIGTERM.
5. Graceful Shutdown: Synchronously stops all components in reverse order.

Example usage:

	appLauncher := launcher.New(logger)
	appLauncher.Append(dbComponent, serverComponent)
	appLauncher.BeforeStart(func() error {
		// Manual DI here
		return nil
	})

	appLauncher.Run() // Blocks until signal received
*/
package launcher
