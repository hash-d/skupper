package frame2

import "github.com/skupperproject/skupper/pkg/images"

// Frame2-specific environment variables

const (
	// This sets the 'Allow' parameter of the retry block for the final
	// validations, and needs to be an integer value.  Any validations
	// marked as final will be retried this many times at the end of the
	// test.
	ENV_FINAL_RETRY = "SKUPPER_TEST_FINAL_RETRY"
)

// TODO: Move all skupper-specific variables to a skupper-specific file, on a
// skupper-specific package
const (

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

	// A path to the Skupper binary to be used (the actual file, not just its parent directory)
	ENV_OLD_BIN = "SKUPPER_TEST_OLD_BIN"

	// The version that ENV_OLD_BIN refers to, such as 1.2 or 1.4.0-rc3
	ENV_OLD_VERSION = "SKUPPER_TEST_OLD_VERSION"

	// All image env variables from pkg/images/image_utils.go should be here
	EnvOldRouterImageEnvKey                 string = "SKUPPER_TEST_OLD_QDROUTERD_IMAGE"
	EnvOldServiceControllerImageEnvKey      string = "SKUPPER_TEST_OLD_SKUPPER_SERVICE_CONTROLLER_IMAGE"
	EnvOldConfigSyncImageEnvKey             string = "SKUPPER_TEST_OLD_SKUPPER_CONFIG_SYNC_IMAGE"
	EnvOldFlowCollectorImageEnvKey          string = "SKUPPER_TEST_OLD_SKUPPER_FLOW_COLLECTOR_IMAGE"
	EnvOldPrometheusServerImageEnvKey       string = "SKUPPER_TEST_OLD_PROMETHEUS_SERVER_IMAGE"
	EnvOldRouterPullPolicyEnvKey            string = "SKUPPER_TEST_OLD_QDROUTERD_IMAGE_PULL_POLICY"
	EnvOldServiceControllerPullPolicyEnvKey string = "SKUPPER_TEST_OLD_SKUPPER_SERVICE_CONTROLLER_IMAGE_PULL_POLICY"
	EnvOldConfigSyncPullPolicyEnvKey        string = "SKUPPER_TEST_OLD_SKUPPER_CONFIG_SYNC_IMAGE_PULL_POLICY"
	EnvOldFlowCollectorPullPolicyEnvKey     string = "SKUPPER_TEST_OLD_SKUPPER_FLOW_COLLECTOR_IMAGE_PULL_POLICY"
	EnvOldPrometheusServerPullPolicyEnvKey  string = "SKUPPER_TEST_OLD_PROMETHEUS_SERVER_IMAGE_PULL_POLICY"
	EnvOldSkupperImageRegistryEnvKey        string = "SKUPPER_TEST_OLD_SKUPPER_IMAGE_REGISTRY"
	EnvOldPrometheusImageRegistryEnvKey     string = "SKUPPER_TEST_OLD_PROMETHEUS_IMAGE_REGISTRY"
)

// final
//
// The map between the variables that indicate the image value for the old version, and the
// environment variable that actually needs to be set on the environment for that configuration
// to be effective.  Perhaps it would be simpler to just s/SKUPPER_TEST_OLD//?
var EnvOldMap = map[string]string{
	EnvOldRouterImageEnvKey:                 images.RouterImageEnvKey,
	EnvOldServiceControllerImageEnvKey:      images.ServiceControllerImageEnvKey,
	EnvOldConfigSyncImageEnvKey:             images.ConfigSyncImageEnvKey,
	EnvOldFlowCollectorImageEnvKey:          images.FlowCollectorImageEnvKey,
	EnvOldPrometheusServerImageEnvKey:       images.PrometheusServerImageEnvKey,
	EnvOldRouterPullPolicyEnvKey:            images.RouterPullPolicyEnvKey,
	EnvOldServiceControllerPullPolicyEnvKey: images.ServiceControllerPullPolicyEnvKey,
	EnvOldConfigSyncPullPolicyEnvKey:        images.ConfigSyncPullPolicyEnvKey,
	EnvOldFlowCollectorPullPolicyEnvKey:     images.FlowCollectorPullPolicyEnvKey,
	EnvOldPrometheusServerPullPolicyEnvKey:  images.PrometheusServerPullPolicyEnvKey,
	EnvOldSkupperImageRegistryEnvKey:        images.SkupperImageRegistryEnvKey,
	EnvOldPrometheusImageRegistryEnvKey:     images.PrometheusImageRegistryEnvKey,
}
