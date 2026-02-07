//go:build !windows

package util

func FindWeChatInstallPaths() []string {
	return []string{}
}

func FindWeChatDataPaths() []string {
	return []string{}
}
