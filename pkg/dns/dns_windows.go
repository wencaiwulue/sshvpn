//go:build windows

package dns

import (
	"context"
)

func GetDnsServers(ctx context.Context) ([]string, error) {
	return nil, nil
}

func SetDnsServers(ctx context.Context, servers []string) error {
	return nil
}
