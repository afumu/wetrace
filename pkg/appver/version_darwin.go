package appver

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"howett.net/plist"
)

const (
	InfoFile = "Info.plist"
)

type Plist struct {
	CFBundleShortVersionString string `plist:"CFBundleShortVersionString"`
	NSHumanReadableCopyright   string `plist:"NSHumanReadableCopyright"`
}

func (i *Info) initialize() error {

	parts := strings.Split(i.FilePath, string(filepath.Separator))
	file := filepath.Join(append(parts[:len(parts)-2], InfoFile)...)
	b, err := os.ReadFile("/" + file)
	if err != nil {
		return err
	}

	p := Plist{}
	_, err = plist.Unmarshal(b, &p)
	if err != nil {
		return err
	}

	i.FullVersion = p.CFBundleShortVersionString
	i.Version, _ = strconv.Atoi(strings.Split(i.FullVersion, ".")[0])
	i.CompanyName = p.NSHumanReadableCopyright

	return nil
}
