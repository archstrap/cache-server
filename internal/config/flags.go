package config

import "github.com/spf13/pflag"

var (
	Dir        string
	Dbfilename string
	Store      map[string]string
)

func ReadFlags() {

	pflag.StringVar(&Dir, "dir", "/tmp/redis-data", "the path to the directory where the RDB file is stored (example: /tmp/redis-data)")
	pflag.StringVar(&Dbfilename, "dbfilename", "rdbfile", "the name of the RDB file (example: rdbfile)")
	pflag.Parse()

	Store = make(map[string]string)

	pflag.VisitAll(func(f *pflag.Flag) {
		Store[f.Name] = f.Value.String()
	})
}
