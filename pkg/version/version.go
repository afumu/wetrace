package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

var (
	Version   = "(dev)"
	buildInfo = debug.BuildInfo{}
)

func init() {
	if bi, ok := debug.ReadBuildInfo(); ok {
		buildInfo = *bi
		if len(bi.Main.Version) > 0 {
			Version = bi.Main.Version
		}
	}
}

func GetMore(mod bool) string {
	if mod {
		mod := buildInfo.String()
		if len(mod) > 0 {
			return fmt.Sprintf("\t%s\n", strings.ReplaceAll(mod[:len(mod)-1], "\n", "\n\t"))
		}
	}
	return fmt.Sprintf("version %s %s %s/%s\n", Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
