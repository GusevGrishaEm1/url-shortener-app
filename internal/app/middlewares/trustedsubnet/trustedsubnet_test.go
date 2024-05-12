package trustedsubnet

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
)

func TestTrustedSubnet(t *testing.T) {
	config := config.GetDefault()
	config.TrustedSubnet = "192.168.0.0/24"
	middleware := NewTrustedSubnetMiddleware(config)
	handler := middleware.TrustedSubnet(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	tests := []struct {
		ip          string
		isForbidden bool
	}{{
		ip:          "89.0.142.86",
		isForbidden: true,
	}, {
		ip:          "192.168.0.1",
		isForbidden: false,
	},
	}
	for _, test := range tests {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.Header.Add("X-Real-IP", test.ip)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, request)
		if rr.Code == http.StatusOK {
			if test.isForbidden {
				t.Errorf("TrustedSubnetMiddleware() forbidden request, got %v", rr.Code)
			}
		}
	}
}
