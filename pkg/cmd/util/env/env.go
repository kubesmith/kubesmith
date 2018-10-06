package env

import (
	"os"
	"strings"

	"github.com/spf13/pflag"
)

func BindEnvToFlag(flagName string, flags *pflag.FlagSet) {
	envKey := strings.ToUpper(strings.Replace(flagName, "-", "_", -1))
	flag := flags.Lookup(flagName)

	if !flag.Changed {
		value := os.Getenv(envKey)

		if value != "" {
			flags.Set(flagName, value)
		}
	}
}
