package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"net"
	"net/http"
	"strings"

	"golang.org/x/sys/windows"
)

// El servidor solo escucha en 127.0.0.1 y exige la clave de esta sesión en cada llamada:
// otras webs o programas no pueden pedirle que instale nada.

func nuevaClave() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

type servidor struct {
	l     *Launcher
	clave string
	host  string // 127.0.0.1:puerto
}

func iniciarServidor(l *Launcher) (*servidor, net.Listener, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, nil, err
	}
	s := &servidor{l: l, clave: nuevaClave(), host: ln.Addr().String()}
	return s, ln, nil
}

func (s *servidor) handler() http.Handler {
	mux := http.NewServeMux()
	web, _ := fs.Sub(webFS, "web")
	estaticos := http.FileServer(http.FS(web))
	mux.Handle("/", s.soloLocal(estaticos))

	mux.HandleFunc("/api/estado", s.api(func(w http.ResponseWriter, r *http.Request) {
		responder(w, s.l.Estado())
	}))
	mux.HandleFunc("/api/revisar", s.post(func(w http.ResponseWriter, r *http.Request) {
		go s.l.Revisar()
		responder(w, map[string]bool{"ok": true})
	}))
	conID := func(f func(string) error) http.HandlerFunc {
		return s.post(func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				ID string `json:"id"`
			}
			if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
				fallar(w, http.StatusBadRequest, msj("peticionInvalida"))
				return
			}
			if err := f(req.ID); err != nil {
				fallar(w, http.StatusUnprocessableEntity, comoMensaje(err))
				return
			}
			responder(w, map[string]bool{"ok": true})
		})
	}
	mux.HandleFunc("/api/instalar", conID(s.l.Instalar))
	mux.HandleFunc("/api/desinstalar", conID(s.l.Desinstalar))
	mux.HandleFunc("/api/abrir", conID(s.l.Abrir))
	mux.HandleFunc("/api/launcher/actualizar", s.post(func(w http.ResponseWriter, r *http.Request) {
		if err := s.l.ActualizarLauncher(); err != nil {
			fallar(w, http.StatusUnprocessableEntity, comoMensaje(err))
			return
		}
		responder(w, map[string]bool{"ok": true})
	}))
	mux.HandleFunc("/api/ajustes", s.post(func(w http.ResponseWriter, r *http.Request) {
		var a Ajustes
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&a); err != nil {
			fallar(w, http.StatusBadRequest, msj("peticionInvalida"))
			return
		}
		if err := s.l.GuardarAjustes(a); err != nil {
			fallar(w, http.StatusUnprocessableEntity, comoMensaje(err))
			return
		}
		responder(w, map[string]bool{"ok": true})
	}))
	mux.HandleFunc("/api/enlace", s.post(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
			fallar(w, http.StatusBadRequest, msj("peticionInvalida"))
			return
		}
		if err := s.abrirEnlace(req.URL); err != nil {
			fallar(w, http.StatusForbidden, comoMensaje(err))
			return
		}
		responder(w, map[string]bool{"ok": true})
	}))
	return mux
}

// soloLocal rechaza peticiones dirigidas a otro host (protección contra DNS rebinding).
func (s *servidor) soloLocal(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != s.host {
			fallar(w, http.StatusForbidden, msj("accesoDenegado"))
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data: https:; style-src 'self'; script-src 'self'; connect-src 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		h.ServeHTTP(w, r)
	})
}

func (s *servidor) api(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Host != s.host {
			fallar(w, http.StatusForbidden, msj("accesoDenegado"))
			return
		}
		if o := r.Header.Get("Origin"); o != "" && o != "http://"+s.host {
			fallar(w, http.StatusForbidden, msj("accesoDenegado"))
			return
		}
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Codex")), []byte(s.clave)) != 1 {
			fallar(w, http.StatusUnauthorized, msj("accesoDenegado"))
			return
		}
		f(w, r)
	}
}

func (s *servidor) post(f http.HandlerFunc) http.HandlerFunc {
	return s.api(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			fallar(w, http.StatusMethodNotAllowed, msj("peticionInvalida"))
			return
		}
		f(w, r)
	})
}

// abrirEnlace abre en el navegador solo enlaces que vienen del catálogo o de las releases.
func (s *servidor) abrirEnlace(url string) error {
	if !strings.HasPrefix(url, "https://") {
		return msj("enlaceNoPermitido")
	}
	permitidos := map[string]bool{}
	e := s.l.Estado()
	for _, a := range e.Apps {
		permitidos[a.Web], permitidos[a.URLRepoGitHub] = true, true
		if a.Ultima != nil {
			permitidos[a.Ultima.Notas] = true
		}
	}
	if e.Propia != nil {
		permitidos[e.Propia.Notas] = true
	}
	if repoLauncher != "" {
		permitidos["https://github.com/"+repoLauncher] = true
	}
	delete(permitidos, "")
	if !permitidos[url] {
		return msj("enlaceNoPermitido")
	}
	verbo, _ := windows.UTF16PtrFromString("open")
	destino, _ := windows.UTF16PtrFromString(url)
	return windows.ShellExecute(0, verbo, destino, nil, nil, 1)
}

func responder(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(v)
}

// fallar responde con un error que la interfaz muestra en su idioma.
func fallar(w http.ResponseWriter, estado int, m *Mensaje) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(map[string]*Mensaje{"error": m})
}
