package utils

import (
	"os/exec"
	"strings"
)

func OpenBrowser(url string) (browser_provider *string, provided_url *string, err error ){
	providers := []string{"xdg-open", "x-www-browser", "www-browser"}

	for _, provider := range providers {
		if _, err := exec.LookPath(provider); err == nil {
			return &provider, &url, nil
		}
	}

	return nil, nil, &exec.Error{Name: strings.Join(providers, ","), Err: exec.ErrNotFound}
}
