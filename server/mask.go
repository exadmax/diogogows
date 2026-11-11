package main

import (
	"net/url"
	"strings"
)

// isMasked: a URL é a versão mascarada devolvida na leitura (a tela a reenviou sem editar): vale como "manter a atual".
func isMasked(u string) bool { return strings.Contains(u, ":*****@") || strings.Contains(u, "/…") }

// maskURL esconde o segredo de uma URL: a senha de um RTSP, ou o caminho (token) de um iCal.
func maskURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return ""
	}
	if _, has := u.User.Password(); has {
		return u.Scheme + "://" + u.User.Username() + ":*****@" + u.Host + u.Path
	}
	if strings.HasPrefix(u.Scheme, "http") {
		tail := u.Path
		if len(tail) > 12 {
			tail = tail[len(tail)-12:]
		}
		return u.Scheme + "://" + u.Host + "/…" + tail
	}
	return u.Scheme + "://" + u.Host + u.Path
}
