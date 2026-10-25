package forecast

import (
	"github.com/awesome-foundation/cfnrain/internal/config"
	fc "github.com/awesome-foundation/cfnrain/plugins/forecast"
)

// AWS::AutoScaling::LaunchConfiguration

func CheckAutoScalingLaunchConfiguration(input fc.PredictionInput) fc.Forecast {

	forecast := fc.MakeForecast(&input)

	config.Debugf("About to check key name for launch config")

	// Check to see if the key name exists
	checkKeyName(&input, &forecast)

	return forecast

}
