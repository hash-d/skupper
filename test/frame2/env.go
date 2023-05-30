package frame2

// Frame2-specific environment variables

type TestUpgradeStrategy string

const (
	// This sets the 'Allow' parameter of the retry block for the final
	// validations, and needs to be an integer value.  Any validations
	// marked as final will be retried this many times at the end of the
	// test.
	ENV_FINAL_RETRY = "SKUPPER_TEST_FINAL_RETRY"

	// Define the upgrade strategy used by the Upgrade disruptor (possibly
	// other points as well?)
	//
	// Values:
	//
	// CREATION (default): order of skupper init
	// PUB_FIRST
	// PRV_FIRST
	// PUB_ONLY
	// PRV_ONLY
	// LEAVES_FIRST
	// LEAVES_ONLY
	// CORE_FIRST
	// CORE_ONLY
	// EDGES_FIRST
	// EDGES_ONLY
	// INTERIOR_FIRST
	// INTERIOR_ONLY
	//
	// Currently, only CREATION and CREATION:INVERSE are implemented
	//
	// For any of the options, if the value ends with :INVERSE, the order
	// is inverted.  For example, ":INVERSE" or "CREATION:INVERSE" will
	// upgrade the lastly installed skupper site first; the first last.
	//
	// Valid values are of the string type TestUpgradeStrategy
	ENV_UPGRADE_STRATEGY = "SKUPPER_TEST_UPGRADE_STRATEGY"
)

// Upgrade strategies accepted by ENV_UPGRADE_STRATEGY
const (
	UPGRADE_STRATEGY_CREATION TestUpgradeStrategy = "CREATION"

	// This one is special; it is set after a colon and inverts the
	// result. For example: ":INVERSE" or "CREATION:INVERSE"
	UPGRADE_STRATEGY_INVERSE TestUpgradeStrategy = "INVERSE"
)
