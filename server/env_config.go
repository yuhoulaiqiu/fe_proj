package main

import (
	"os"
	"strings"
)

func envOr(k, def string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	return v
}

func parseOrigins(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func adminInitEnabled() bool {
	v := strings.TrimSpace(os.Getenv("ADMIN_INIT"))
	if v == "" {
		return true
	}
	v = strings.ToLower(v)
	return !(v == "0" || v == "false" || v == "off" || v == "no")
}
