package config

import (
	"github.com/spf13/pflag"
)

var (
	Dir        string
	Dbfilename string
	Port       string
	ReplicaOf  string
	Store      map[string]string
)

func ReadFlags() {

	pflag.StringVar(&Dir, "dir", "", "the path to the directory where the RDB file is stored (example: /tmp/redis-data)")
	pflag.StringVar(&Dbfilename, "dbfilename", "", "the name of the RDB file (example: rdbfile)")
	pflag.StringVar(&Port, "port", appConfig.GetPort(), "Custom Port to run on")
	pflag.StringVar(&ReplicaOf, "replicaof", "", "flag to start a Redis server as a replica")
	pflag.Parse()

	Store = make(map[string]string)

	pflag.VisitAll(func(f *pflag.Flag) {
		Store[f.Name] = f.Value.String()
	})
}

// RDBEnabled returns true when both Dir and Dbfilename are set (RDB load will run).
func RDBEnabled() bool {
	return Dir != "" && Dbfilename != ""
}
