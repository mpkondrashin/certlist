package prompt

import (
	"fmt"
	"slices"

	"github.com/spf13/pflag"
)

func Mandatory(fs *pflag.FlagSet, mandatory ...string) (err error) {
	fs.VisitAll(func(f *pflag.Flag) {
		if err != nil {
			return
		}
		if f.Value.String() != "" {
			return
		}
		if !slices.Contains(mandatory, f.Name) {
			return
		}
		fmt.Printf("%s [%s]: ", f.Usage, f.Name)
		var v string
		_, err = fmt.Scanf("%s", &v)
		if err != nil {
			return
		}
		err = f.Value.Set(v)
	})
	return
}
