package prompt

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"

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
		for {
			_, err = fmt.Printf("%s [%s]: ", f.Usage, f.Name)
			if err != nil {
				return
			}
			reader := bufio.NewReader(os.Stdin)
			v, _ := reader.ReadString('\n')
			v = strings.TrimSpace(v)
			if v != "" {
				err = f.Value.Set(v)
				break
			}
		}
	})
	return
}
