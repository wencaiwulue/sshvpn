package dns

import "context"

func Append(ctx context.Context, strings []string) error {
	servers, err := GetDnsServers(ctx)
	if err != nil {
		return err
	}
	servers = append(servers, strings...)
	err = SetDnsServers(ctx, servers)
	return err
}

func Remove(ctx context.Context, strings []string) error {
	servers, err := GetDnsServers(ctx)
	if err != nil {
		return err
	}
	for i := 0; i < len(servers); i++ {
	out:
		for _, s := range strings {
			if servers[i] == s {
				servers = append(servers[:i], servers[i+1:]...)
				i--
				continue out
			}
		}
	}
	err = SetDnsServers(ctx, servers)
	return err
}
