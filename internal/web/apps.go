package web

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Sheila724/incidente-zero/internal/lab"
	"github.com/Sheila724/incidente-zero/internal/scenario"
)

// Endereço do lab em cache: cada página do Zabbix carrega dezenas de arquivos.
type ipEntry struct {
	ip  string
	exp time.Time
}

var (
	ipMu    sync.Mutex
	ipCache = map[string]ipEntry{}
)

func labIP(sc *scenario.Scenario) (string, error) {
	ipMu.Lock()
	e, ok := ipCache[sc.ID]
	ipMu.Unlock()
	if ok && time.Now().Before(e.exp) {
		return e.ip, nil
	}
	ip, err := lab.IP(sc)
	if err != nil {
		return "", err
	}
	ipMu.Lock()
	ipCache[sc.ID] = ipEntry{ip, time.Now().Add(15 * time.Second)}
	ipMu.Unlock()
	return ip, nil
}

var cookiePath = regexp.MustCompile(`(?i)(path=)/(;|$)`)

// appProxy publica uma interface web do lab em /apps/{id}/{app}/...
// O acesso passa pela mesma autenticação da interface do iz.
func (s *Server) appProxy(w http.ResponseWriter, r *http.Request) {
	sc, ok := s.findRunning(w, r)
	if !ok {
		return
	}
	app, ok := sc.App(r.PathValue("app"))
	if !ok {
		writeErr(w, http.StatusNotFound, "este cenário não tem essa interface")
		return
	}
	ip, err := labIP(sc)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	prefix := "/apps/" + r.PathValue("id") + "/" + app.Name
	if r.URL.Path == prefix {
		http.Redirect(w, r, prefix+"/", http.StatusFound)
		return
	}
	target := &url.URL{Scheme: "http", Host: ip + ":" + strconv.Itoa(app.Port)}
	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(target)
			pr.Out.URL.Path = strings.TrimPrefix(pr.In.URL.Path, prefix)
			pr.Out.URL.RawPath = ""
			pr.Out.Host = target.Host
			pr.SetXForwarded()
		},
		ModifyResponse: func(res *http.Response) error {
			// Redirecionamentos absolutos ("/index.php") voltam para dentro do prefixo
			if loc := res.Header.Get("Location"); strings.HasPrefix(loc, "/") && !strings.HasPrefix(loc, "//") {
				res.Header.Set("Location", prefix+loc)
			}
			// Cookies do app ficam restritos ao prefixo dele
			cookies := res.Header.Values("Set-Cookie")
			res.Header.Del("Set-Cookie")
			for _, c := range cookies {
				res.Header.Add("Set-Cookie", cookiePath.ReplaceAllString(c, "${1}"+prefix+"/${2}"))
			}
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, _ *http.Request, _ error) {
			writeErr(w, http.StatusBadGateway, app.Title+" ainda não respondeu: o lab pode estar iniciando")
		},
	}
	proxy.ServeHTTP(w, r)
}
