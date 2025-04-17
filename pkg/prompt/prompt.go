package prompt

import (
	"fmt"

	"github.com/spf13/pflag"
)

func Mandatory(fs *pflag.FlagSet, mandatory ...string) {
	fs.VisitAll(func(f *pflag.Flag) {
		for _, each := range mandatory {
			if each == f.Name {
				if f.Value.String() != "" {
					continue
				}
				fmt.Printf("%s (%s): ", f.Usage, f.Name)
				var v string
				fmt.Scanf("%s", &v)
				f.Value.Set(v)
			}
		}
		/*
				Name     string // name as it appears on command line
			Usage    string // help message
			Value    Value  // value as set
			DefValue string // default value (as text); for usage message*/
	})
}
