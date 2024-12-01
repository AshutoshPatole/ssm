package store

import (
	"net"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var c Config

func Save(group, environment, host, user, alias, password string, isRDP bool) {
	err := viper.Unmarshal(&c)
	if err != nil {
		logrus.Fatal("Failed to unmarshal configuration:", err)
	}

	doesGroupExist := false
	doesEnvironmentExist := false
	groupIndex := -1
	environmentIndex := -1

	// Find if group and environment already exists in yaml config file
	for i, grp := range c.Groups {
		if grp.Name == group {
			doesGroupExist = true
			groupIndex = i
			for j, env := range grp.Environment {
				if env.Name == environment {
					doesEnvironmentExist = true
					environmentIndex = j
					break
				}
			}
			break
		}
	}

	server := Server{
		HostName: host,
		IP:       getIP(host),
		Alias:    alias,
		User:     user,
		IsRDP:    isRDP,
		Password: password,
	}

	if !doesGroupExist {
		newGroup := Group{
			Name: group,
			Environment: []Env{{
				Name:    environment,
				Servers: []Server{server},
			}},
		}
		c.Groups = append(c.Groups, newGroup)
	} else {
		if !doesEnvironmentExist {
			newEnv := Env{
				Name:    environment,
				Servers: []Server{server},
			}
			c.Groups[groupIndex].Environment = append(c.Groups[groupIndex].Environment, newEnv)
		} else {
			isDuplicate := checkDuplicateServer(server, c.Groups[groupIndex].Environment[environmentIndex].Servers)
			if isDuplicate {
				logrus.Warn("Duplicate server found in group")
				return
			}
			c.Groups[groupIndex].Environment[environmentIndex].Servers = append(c.Groups[groupIndex].Environment[environmentIndex].Servers, server)
		}
	}

	err = viper.WriteConfig()
	if err != nil {
		logrus.Fatal("Failed to write configuration:", err)
	}
}

func checkDuplicateServer(s Server, servers []Server) bool {
	for _, server := range servers {
		// Check both IP and hostname to be more thorough
		if server.IP == s.IP || server.HostName == s.HostName {
			return true
		}
	}
	return false
}

func getIP(host string) string {
	lookupHost, err := net.LookupHost(host)
	if err != nil {
		logrus.Error("Could not resolve IP from hostname:", err)
		return host
	}
	if len(lookupHost) == 0 {
		return host
	}
	return lookupHost[0]
}
