//go:build !windows && !darwin

package appver

func (i *Info) initialize() error {
	return nil
}
