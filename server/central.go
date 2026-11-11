// Ligação com a Central Diogo. A Central é outro programa (outro serviço, outra porta, outro banco): aqui o Diogows só
// repassa, para quem está logado, as rotas dela que as janelas do desktop usam: câmeras ao vivo, contatos, agendas da casa e
// o menu/configurações do painel. Assim o navegador conversa sempre com o mesmo endereço (o do Diogows) e a Central não precisa
// ser alcançável de fora do servidor nem liberar CORS.
//
// Regras: leitura exige sessão; escrita exige sessão + origem igual à do Diogows; alterar /api/settings/* (câmeras, agendas,
// rotas, menu da Central) exige administrador. Sem "central.url" nada disso é registrado e os aplicativos que dependem da
// Central mostram "indisponível" sozinhos.
package main

import (
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// centralPrefixes são os caminhos da Central que o desktop pode alcançar (o resto da API dela não é repassado).
var centralPrefixes = []string{"/api/cameras", "/api/contatos", "/api/ics", "/api/settings"}

type centralProxy struct {
	d  *deskAPI
	rp *httputil.ReverseProxy
}

func newCentralProxy(d *deskAPI, raw string) (*centralProxy, error) {
	u, err := url.Parse(strings.TrimRight(strings.TrimSpace(raw), "/"))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, errors.New("central.url precisa ser um endereço http(s) completo, ex.: http://127.0.0.1:8080")
	}
	rp := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(u)
			pr.Out.Header.Del("Cookie") // a sessão do Diogows não vai para a Central
			pr.Out.Header.Del("Origin") // já conferida aqui; a Central compararia com o endereço interno dela
			pr.Out.Header.Del("Referer")
			pr.Out.Header.Del("X-Desk")
		},
		FlushInterval: -1, // o vídeo das câmeras precisa sair conforme chega
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("central: %s %s: %v", r.Method, r.URL.Path, err)
			deskErr(w, http.StatusBadGateway, "a Central Diogo não respondeu (veja se o serviço dela está no ar)")
		},
	}
	return &centralProxy{d: d, rp: rp}, nil
}

func (c *centralProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	u := c.d.sessionUser(r)
	if u == nil || u.MustChange {
		deskErr(w, http.StatusUnauthorized, "entre no Diogows para usar este recurso")
		return
	}
	write := r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions
	if write && !sameOrigin(r) {
		deskErr(w, http.StatusForbidden, "pedido recusado (origem inválida)")
		return
	}
	if write && strings.HasPrefix(r.URL.Path, "/api/settings/") && !u.IsAdmin {
		deskErr(w, http.StatusForbidden, "só o administrador altera a configuração da Central")
		return
	}
	if write {
		// A Central protege as próprias gravações com este cabeçalho (um formulário de outro site não consegue enviá-lo). Aqui a sessão,
		// a origem e o perfil de administrador já foram conferidos, então o Diogows o acrescenta em nome de quem passou pelas checagens.
		r.Header.Set("X-CD-Action", "confirm")
	}
	if strings.HasSuffix(r.URL.Path, "/stream") { // fluxo de câmera: sem o limite de escrita do servidor
		_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})
	}
	c.rp.ServeHTTP(w, r)
}

// registerCentral só liga as rotas se houver central.url; uma URL inválida é registrada no log sem derrubar o Diogows.
func (d *deskAPI) registerCentral(mux *http.ServeMux, raw string) {
	if d.cfg.Disabled || strings.TrimSpace(raw) == "" {
		return
	}
	cp, err := newCentralProxy(d, raw)
	if err != nil {
		log.Printf("central: %v (ligação desativada)", err)
		return
	}
	for _, pre := range centralPrefixes {
		mux.Handle(pre, cp)
		mux.Handle(pre+"/", cp)
	}
	log.Printf("central: repassando câmeras, contatos, agendas e menu de %s", raw)
}
