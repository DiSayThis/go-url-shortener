package auth

import (
	"go-api/pkg/jwt"
	"net/http"
	"net/netip"
)

func RequirePrincipal(w http.ResponseWriter, req *http.Request) (jwt.Principal, bool) {
	principal, ok := jwt.PrincipalFromContext(req.Context())
	if !ok {
		jwt.WriteUnauthorized(w)
		return jwt.Principal{}, false
	}

	return principal, true
}

func requestRemoteIP(req *http.Request) *netip.Addr {
	addrPort, err := netip.ParseAddrPort(req.RemoteAddr)
	if err == nil {
		addr := addrPort.Addr().Unmap()
		return &addr
	}

	// Удобно для тестов или нестандартного входа без порта.
	addr, err := netip.ParseAddr(req.RemoteAddr)
	if err != nil {
		return nil
	}

	addr = addr.Unmap()
	return &addr
}
