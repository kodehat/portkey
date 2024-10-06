package utils

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"strings"
)

func IpCheck(logger *slog.Logger, ipRanges []string, h http.Handler) http.Handler {
	prefixes := ipRangesFromString(ipRanges)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, err := ipFromRequest(r)
		if err != nil {
			logger.Error("unable to parse ip address", "address", r.RemoteAddr)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		for _, prefix := range prefixes {
			if prefix.Contains(ip) {
				logger.Debug("ip allowed to access resource", "pattern", r.Pattern, "ip", ip.String(), "range", prefix.String())
				h.ServeHTTP(w, r)
				return
			}
		}
		logger.Warn("ip not allowed to access metrics", "ip", ip.String())
		w.WriteHeader(http.StatusForbidden)
	})
}

func ipFromRequest(r *http.Request) (netip.Addr, error) {
	ip := r.RemoteAddr
	if strings.IndexByte(r.RemoteAddr, byte(':')) >= 0 {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return netip.ParseAddr(ip)
}

func ipRangesFromString(ipRanges []string) []netip.Prefix {
	prefixes := make([]netip.Prefix, 0)
	for _, ipRange := range ipRanges {
		prefix, err := netip.ParsePrefix(ipRange)
		if err != nil {
			panic(fmt.Errorf("unable to parse CIDR: %w", err))
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes
}
