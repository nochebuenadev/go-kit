package fb

// Config defines the configuration for the Firebase component.
type Config struct {
	// ProjectID is the Google Cloud Project ID for Firebase.
	ProjectID string `env:"FIREBASE_PROJECT_ID"`
}
