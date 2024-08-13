package store

type Server struct {
	HostName string `yaml:"hostname"`
	IP       string `yaml:"ip"`
	Alias    string `yaml:"alias"`
	User     string `yaml:"user"`
}

type Env struct {
	Name    string   `yaml:"name"`
	Servers []Server `yaml:"servers"`
}

type Group struct {
	Name        string `yaml:"name"`
	Environment []Env  `yaml:"environment"`
}

type Config struct {
	Groups []Group `yaml:"groups"`
}
