package utils

import (
	"regexp"

	"github.com/go-kratos/kratos/v2/transport/http"
)

func GetClientIp(req *http.Request) string {
	ip := req.Header.Get("X-Real-IP")
	if ip == "" {
		ip = req.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		re := regexp.MustCompile(`:\d+$`)
		ip = re.ReplaceAllString(req.RemoteAddr, "")
	}
	return ip
}
