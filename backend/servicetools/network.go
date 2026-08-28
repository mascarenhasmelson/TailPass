package servicetools

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

// tailscaleCGNAT is Tailscale's carrier-grade NAT range that every device's
// Tailscale IPv4 address is drawn from.
var tailscaleCGNAT = func() *net.IPNet {
	_, cidr, _ := net.ParseCIDR("100.64.0.0/10")
	return cidr
}()

// GetTailscaleIP always resolves this host's Tailscale interface IP (a
// 100.64.0.0/10 address) rather than requiring it to be typed in manually.
// It first asks the local `tailscale` CLI, then falls back to scanning
// network interfaces for an address in Tailscale's CGNAT range so it also
// works on hosts where the CLI isn't on PATH but the interface is up.
func GetTailscaleIP() (string, error) {
	if ip, err := tailscaleIPFromCLI(); err == nil && ip != "" {
		return ip, nil
	}
	if ip, err := tailscaleIPFromInterfaces(); err == nil && ip != "" {
		return ip, nil
	}
	return "", fmt.Errorf("no Tailscale IP found: is tailscaled running and this device joined to a tailnet?")
}

func tailscaleIPFromCLI() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "tailscale", "ip", "-4")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	ip := strings.TrimSpace(out.String())
	if parsed := net.ParseIP(ip); parsed == nil || parsed.To4() == nil {
		return "", fmt.Errorf("unexpected output from tailscale CLI: %q", ip)
	}
	return ip, nil
}

func tailscaleIPFromInterfaces() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	// Check likely Tailscale interface names first (tailscale0 on Linux,
	// utun* on macOS, Tailscale on Windows), then fall back to scanning
	// every interface for a CGNAT address in case naming differs.
	var preferred, other []net.Interface
	for _, iface := range ifaces {
		name := strings.ToLower(iface.Name)
		if strings.Contains(name, "tailscale") || strings.HasPrefix(name, "utun") || strings.HasPrefix(name, "tun") {
			preferred = append(preferred, iface)
		} else {
			other = append(other, iface)
		}
	}

	for _, group := range [][]net.Interface{preferred, other} {
		for _, iface := range group {
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if ip == nil {
					continue
				}
				if v4 := ip.To4(); v4 != nil && tailscaleCGNAT.Contains(v4) {
					return v4.String(), nil
				}
			}
		}
	}
	return "", fmt.Errorf("no interface with a 100.64.0.0/10 Tailscale address found")
}

// GetFreePort asks the OS for a free TCP port on the given IP and returns it.
// This is used to generate a random, currently-unused local port whenever the
// caller doesn't pin a specific one. Pass "" to probe on all interfaces.
func GetFreePort(ip string) (int, error) {
	bindIP := ip
	if bindIP == "" {
		bindIP = "0.0.0.0"
	}
	l, err := net.Listen("tcp", net.JoinHostPort(bindIP, "0"))
	if err != nil {
		return 0, fmt.Errorf("failed to find a free port on %s: %w", bindIP, err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
