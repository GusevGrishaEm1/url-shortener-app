package trustedsubnet

import (
	"net"
	"net/http"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
)

type TrustedSubnetMiddleware struct {
	config config.Config
}

func NewTrustedSubnetMiddleware(config config.Config) *TrustedSubnetMiddleware {
	return &TrustedSubnetMiddleware{config: config}
}

func (m *TrustedSubnetMiddleware) TrustedSubnet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("X-Real-IP")

		if m.config.TrustedSubnet != "" {
			_, ipnet, err := net.ParseCIDR(m.config.TrustedSubnet)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			ipAddress := net.ParseIP(ip)
			if !ipnet.Contains(ipAddress) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		} else {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
