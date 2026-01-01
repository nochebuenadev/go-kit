/*
Package fb provides a wrapper around the Firebase Admin SDK for Go.

It integrates Firebase with the application lifecycle (launcher.Component) and provides
a singleton provider for accessing the Firebase App instance.

Features:
- Integration with application lifecycle.
- Singleton provider for Firebase App.
- Automated initialization via environment-based configuration.

Example usage:

	cfg := &fb.Config{ProjectID: "my-project"}
	fbComp := fb.GetFirebase(logger, cfg)

	// Access the app:
	app := fbComp.App()
*/
package fb
