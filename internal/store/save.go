package store

import (
	"net"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Save(group, environment, host, user, alias, password string, isRDP bool) {
	var c Config
	err := viper.Unmarshal(&c)
	if err != nil {
		logrus.Errorf("Error reading configuration: %v", err)
		return
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

	env := Env{
		Name:    environment,
		Servers: []Server{server},
	}
	if !doesGroupExist {
		newGroup := Group{
			Name:        group,
			Environment: []Env{env},
		}
		c.Groups = append(c.Groups, newGroup)
		groupIndex = len(c.Groups) - 1

	} else {
		if !doesEnvironmentExist {
			newEnv := Env{
				Name:    environment,
				Servers: []Server{server},
			}
			c.Groups[groupIndex].Environment = append(c.Groups[groupIndex].Environment, newEnv)
			environmentIndex = len(c.Groups[groupIndex].Environment) - 1
		} else {
			isDuplicate := checkDuplicateServer(server, c.Groups[groupIndex].Environment[environmentIndex].Servers)
			if isDuplicate {
				logrus.Println("Duplicate server found in group")
			} else {
				c.Groups[groupIndex].Environment[environmentIndex].Servers = append(c.Groups[groupIndex].Environment[environmentIndex].Servers, server)
			}
		}
	}

	viper.Set("groups", c.Groups)
	err = viper.WriteConfig()
	if err != nil {
		logrus.Errorf("Error writing config: %v", err)
	}
}

func checkDuplicateServer(s Server, servers []Server) bool {
	for _, server := range servers {
		if (server.IP != "" && server.IP == s.IP) || (server.Alias != "" && server.Alias == s.Alias) || (server.HostName != "" && server.HostName == s.HostName) {
			return true
		}
	}
	return false
}

func getIP(host string) string {
	lookupHost, err := net.LookupHost(host)
	if err != nil || len(lookupHost) == 0 {
		logrus.Debugf("Could not resolve IP from hostname %s, using hostname as IP", host)
		return host
	}
	return lookupHost[0]
}
