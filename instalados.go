package main

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// Instalada describe una app encontrada en el registro de Windows (las mismas entradas
// que muestra Configuración → Aplicaciones instaladas).
type Instalada struct {
	Version     string
	Carpeta     string
	Icono       string
	Desinstalar string // comando silencioso si existe, si no el normal
	Silencioso  bool
}

const claveDesinstalar = `Software\Microsoft\Windows\CurrentVersion\Uninstall`

// buscarInstalada busca por DisplayName en HKLM (64 y 32 bits) y HKCU.
func buscarInstalada(nombre string) *Instalada {
	lugares := []struct {
		raiz   registry.Key
		acceso uint32
	}{
		{registry.LOCAL_MACHINE, registry.WOW64_64KEY},
		{registry.LOCAL_MACHINE, registry.WOW64_32KEY},
		{registry.CURRENT_USER, 0},
	}
	for _, l := range lugares {
		k, err := registry.OpenKey(l.raiz, claveDesinstalar, registry.READ|l.acceso)
		if err != nil {
			continue
		}
		subs, _ := k.ReadSubKeyNames(-1)
		k.Close()
		for _, s := range subs {
			sk, err := registry.OpenKey(l.raiz, claveDesinstalar+`\`+s, registry.QUERY_VALUE|l.acceso)
			if err != nil {
				continue
			}
			dn, _, _ := sk.GetStringValue("DisplayName")
			if !strings.EqualFold(strings.TrimSpace(dn), nombre) {
				sk.Close()
				continue
			}
			inst := &Instalada{}
			inst.Version, _, _ = sk.GetStringValue("DisplayVersion")
			inst.Carpeta, _, _ = sk.GetStringValue("InstallLocation")
			inst.Icono, _, _ = sk.GetStringValue("DisplayIcon")
			if q, _, err := sk.GetStringValue("QuietUninstallString"); err == nil && q != "" {
				inst.Desinstalar, inst.Silencioso = q, true
			} else {
				inst.Desinstalar, _, _ = sk.GetStringValue("UninstallString")
			}
			sk.Close()
			return inst
		}
	}
	return nil
}

// ejecutableDe devuelve la ruta del .exe de la app instalada. Prueba, en orden: la carpeta
// que registró el instalador, la que guarda en HKLM\Software\<App>\InstallDir (los MSI de
// ArtCraft), la carpeta de Codex y el icono registrado.
func ejecutableDe(app AppCatalogo, inst *Instalada) string {
	carpetas := []string{strings.Trim(inst.Carpeta, `"`)}
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE, `Software\`+app.Registro(), registry.QUERY_VALUE|registry.WOW64_64KEY); err == nil {
		dir, _, _ := k.GetStringValue("InstallDir")
		k.Close()
		carpetas = append(carpetas, dir)
	}
	carpetas = append(carpetas, filepath.Join(carpetaSuite(), app.NombreCompleto()))
	for _, c := range carpetas {
		if c == "" {
			continue
		}
		p := filepath.Join(c, app.Exe())
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	// DisplayIcon suele ser la ruta del .exe, a veces con ",0" al final.
	ico := strings.Trim(inst.Icono, `"`)
	if i := strings.LastIndex(ico, ","); i > 0 && !strings.Contains(ico[i:], `\`) {
		ico = ico[:i]
	}
	if strings.EqualFold(filepath.Ext(ico), ".exe") {
		if _, err := os.Stat(ico); err == nil {
			return ico
		}
	}
	return ""
}

// separarComando divide un comando del registro en ejecutable y argumentos.
func separarComando(cmd string) (exe, args string) {
	cmd = strings.TrimSpace(cmd)
	if strings.HasPrefix(cmd, `"`) {
		if fin := strings.Index(cmd[1:], `"`); fin >= 0 {
			return cmd[1 : fin+1], strings.TrimSpace(cmd[fin+2:])
		}
	}
	if i := strings.Index(strings.ToLower(cmd), ".exe"); i >= 0 {
		return cmd[:i+4], strings.TrimSpace(cmd[i+4:])
	}
	return cmd, ""
}
