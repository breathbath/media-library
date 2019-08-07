package helper

import (
	"os"
)

func SetEnvs(envs map[string]string) error {
	for envNam, envVal := range envs {
		err := os.Setenv(envNam, envVal)
		if err != nil {
			return err
		}
	}

	return nil
}
