// Пакет trustedsubnet предоставляет middleware,
// который проверяет, находится ли IP-адрес клиента в доверенном диапазоне подсети.
package trustedsubnet

import (
	"net"
	"net/http"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
)

// TrustedSubnetMiddleware представляет middleware, который проверяет,
// находится ли IP-адрес клиента в доверенном диапазоне подсети.
// Если IP-адрес не находится в доверенном диапазоне подсети, то запрос будет отклонен с кодом отказа "Forbidden".
type TrustedSubnetMiddleware struct {
	config config.Config
}

// NewTrustedSubnetMiddleware создает новый экземпляр TrustedSubnetMiddleware с использованием указанной конфигурации.
func NewTrustedSubnetMiddleware(config config.Config) *TrustedSubnetMiddleware {
	return &TrustedSubnetMiddleware{config: config}
}

// TrustedSubnet выполняет проверку IP-адреса клиента на соответствие доверенному диапазону подсети.
// Если IP-адрес не находится в диапазоне, то возвращается код отказа "Forbidden".
// В противном случае вызывается следующий обработчик в цепочке обработчиков.
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
