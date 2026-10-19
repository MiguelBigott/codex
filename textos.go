package main

// Textos e idioma. Casi todos los textos viven en web/i18n.js: Go manda una clave con sus datos
// y la interfaz la muestra en el idioma elegido. Aquí quedan solo los que Go muestra por su
// cuenta: el menú de la bandeja y el aviso de que falta WebView2.

import (
	"errors"
	"slices"
	"strings"

	"golang.org/x/sys/windows"
)

// Mensaje es un texto para la interfaz: una clave del diccionario de web/i18n.js y sus datos
// (por ejemplo {"app": "ClipVault"}). También sirve como error.
type Mensaje struct {
	Clave string            `json:"clave"`
	Datos map[string]string `json:"datos,omitempty"`
}

func (m *Mensaje) Error() string { return m.Clave }

// msj arma un Mensaje; los datos van en pares nombre, valor.
func msj(clave string, datos ...string) *Mensaje {
	m := &Mensaje{Clave: clave}
	if len(datos) > 0 {
		m.Datos = map[string]string{}
		for i := 0; i+1 < len(datos); i += 2 {
			m.Datos[datos[i]] = datos[i+1]
		}
	}
	return m
}

// comoMensaje convierte cualquier error en un Mensaje: los que no vienen del diccionario
// (errores del sistema) se muestran tal cual dentro de una frase traducida.
func comoMensaje(err error) *Mensaje {
	var m *Mensaje
	if errors.As(err, &m) {
		return m
	}
	return msj("errorDetalle", "detalle", err.Error())
}

// idiomas que sabe mostrar Codex, en el orden del selector de ajustes.
var idiomas = []string{"es", "en", "fr", "de", "ru", "pt", "zh", "ja"}

// idiomaSistema es el primero de los idiomas de Windows que Codex sabe mostrar; si no sabe
// ninguno, inglés.
func idiomaSistema() string {
	lista, _ := windows.GetUserPreferredUILanguages(windows.MUI_LANGUAGE_NAME)
	for _, l := range lista {
		if base := strings.ToLower(strings.SplitN(l, "-", 2)[0]); slices.Contains(idiomas, base) {
			return base
		}
	}
	return "en"
}

// idiomaElegido resuelve «auto» (el valor por defecto) al idioma del PC.
func idiomaElegido(a Ajustes) string {
	if slices.Contains(idiomas, a.Idioma) {
		return a.Idioma
	}
	return idiomaSistema()
}

type textosNativos struct{ Abrir, Salir, FaltaWebView string }

var nativos = map[string]textosNativos{
	"es": {"Abrir Codex", "Salir", "Codex necesita Microsoft Edge WebView2, que viene con Windows 11. Instálalo desde https://developer.microsoft.com/microsoft-edge/webview2/ y vuelve a abrir Codex."},
	"en": {"Open Codex", "Quit", "Codex needs Microsoft Edge WebView2, which comes with Windows 11. Install it from https://developer.microsoft.com/microsoft-edge/webview2/ and open Codex again."},
	"fr": {"Ouvrir Codex", "Quitter", "Codex a besoin de Microsoft Edge WebView2, inclus dans Windows 11. Installez-le depuis https://developer.microsoft.com/microsoft-edge/webview2/ puis rouvrez Codex."},
	"de": {"Codex öffnen", "Beenden", "Codex benötigt Microsoft Edge WebView2, das in Windows 11 enthalten ist. Installiere es von https://developer.microsoft.com/microsoft-edge/webview2/ und öffne Codex erneut."},
	"ru": {"Открыть Codex", "Выйти", "Для работы Codex нужен Microsoft Edge WebView2, который входит в Windows 11. Установите его с https://developer.microsoft.com/microsoft-edge/webview2/ и снова откройте Codex."},
	"pt": {"Abrir o Codex", "Sair", "O Codex precisa do Microsoft Edge WebView2, que vem com o Windows 11. Instale-o em https://developer.microsoft.com/microsoft-edge/webview2/ e abra o Codex novamente."},
	"zh": {"打开 Codex", "退出", "Codex 需要 Microsoft Edge WebView2（Windows 11 已自带）。请从 https://developer.microsoft.com/microsoft-edge/webview2/ 安装后重新打开 Codex。"},
	"ja": {"Codex を開く", "終了", "Codex には Microsoft Edge WebView2 が必要です（Windows 11 に付属しています）。https://developer.microsoft.com/microsoft-edge/webview2/ からインストールして、Codex をもう一度開いてください。"},
}

func textosDe(idioma string) textosNativos {
	if t, ok := nativos[idioma]; ok {
		return t
	}
	return nativos["en"]
}
