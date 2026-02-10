package store

import (
	"log"
	"net"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var c Config

func Save(group, environment, host, user, alias, password string, isRDP bool) {
	err := viper.Unmarshal(&c)
	if err != nil {
		logrus.Fatal(err.Error())
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
	SortConfig(&c)
	viper.Set("groups", c.Groups)
	err = viper.WriteConfig()
	if err != nil {
		log.Fatal(err.Error())
	}
}

// SortConfig sorts all servers in every environment by Alias (a->z, case-insensitive).
func SortConfig(cfg *Config) {
	for i := range cfg.Groups {
		for j := range cfg.Groups[i].Environment {
			sort.Slice(cfg.Groups[i].Environment[j].Servers, func(a, b int) bool {
				return strings.ToLower(cfg.Groups[i].Environment[j].Servers[a].Alias) < strings.ToLower(cfg.Groups[i].Environment[j].Servers[b].Alias)
			})
		}
	}
}

func checkDuplicateServer(s Server, servers []Server) bool {
	isDuplicate := false
	for _, server := range servers {
		if server.IP == s.IP {
			isDuplicate = true
		}
	}
	return isDuplicate
}

func getIP(host string) string {
	lookupHost, err := net.LookupHost(host)
	if err != nil {
		logrus.Fatal("Could not resolve IP from hostname")
	}
	return lookupHost[0]
}
