package frame2

// Frame2-specific environment variables

const (
	// This sets the 'Allow' parameter of the retry block for the final
	// validations, and needs to be an integer value.  Any validations
	// marked as final will be retried this many times at the end of the
	// test.
	ENV_FINAL_RETRY = "SKUPPER_TEST_FINAL_RETRY"
)
