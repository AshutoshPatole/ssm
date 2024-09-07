package store

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

func UpdateKeyRotationTime(group, environment, hostname string) error {
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return err
	}

	for i, g := range config.Groups {
		if g.Name == group {
			for j, env := range g.Environment {
				if env.Name == environment {
					for k, server := range env.Servers {
						if server.HostName == hostname {
							config.Groups[i].Environment[j].Servers[k].KeyRotatedAt = time.Now()
							viper.Set("groups", config.Groups)
							if err := viper.WriteConfig(); err != nil {
								return fmt.Errorf("failed to write config: %w", err)
							}
							time.Sleep(time.Millisecond * 50)
							return nil
						}
					}
				}
			}
		}
	}
	return fmt.Errorf("server not found")
}
