package prompt

import (
	"fmt"
	"slices"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func Mandatory(fs *pflag.FlagSet, mandatory ...string) (err error) {
	fs.VisitAll(func(f *pflag.Flag) {
		if err != nil {
			return
		}
		if viper.GetString(f.Name) != "" {
			return
		}
		if !slices.Contains(mandatory, f.Name) {
			return
		}
		_, err = fmt.Printf("%s [%s]: ", f.Usage, f.Name)
		if err != nil {
			return
		}
		var v string
		_, err = fmt.Scanf("%s", &v)
		if err != nil {
			return
		}
		err = f.Value.Set(v)
	})
	return
}
