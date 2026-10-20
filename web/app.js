"use strict";

/* ---------------- Conexión con el launcher ---------------- */

// La clave de sesión llega en el fragmento de la URL; la guardamos y la quitamos de la barra.
const CLAVE = location.hash.slice(1);
history.replaceState(null, "", location.pathname);

async function api(ruta, cuerpo) {
  const opciones = { headers: { "X-Codex": CLAVE } };
  if (cuerpo !== undefined) {
    opciones.method = "POST";
    opciones.headers["Content-Type"] = "application/json";
    opciones.body = JSON.stringify(cuerpo);
  }
  const r = await fetch(ruta, opciones);
  const datos = await r.json().catch(() => ({}));
  if (!r.ok) throw new Error(datos.error ? tm(datos.error) : `Error ${r.status}`);
  return datos;
}

const $ = (s) => document.querySelector(s);
const $$ = (s) => [...document.querySelectorAll(s)];

function el(etiqueta, props = {}, ...hijos) {
  const n = document.createElement(etiqueta);
  for (const [k, v] of Object.entries(props)) {
    if (v === undefined || v === null || v === false) continue;
    if (k === "class") n.className = v;
    else if (k === "text") n.textContent = v;
    else if (k.startsWith("on")) n.addEventListener(k.slice(2), v);
    else if (k === "style") for (const [p, val] of Object.entries(v)) n.style.setProperty(p, val);
    else n.setAttribute(k, v === true ? "" : v);
  }
  for (const h of hijos.flat()) if (h !== null && h !== undefined && h !== false) n.append(h);
  return n;
}

/* ---------------- Estado de la interfaz ---------------- */

function leer(clave, porDefecto) {
  try { return localStorage.getItem(clave) || porDefecto; } catch { return porDefecto; }
}
function guardar(clave, valor) {
  try { localStorage.setItem(clave, valor); } catch { /* sin almacenamiento */ }
}

const ui = {
  estado: null,
  filtro: leer("ca.filtro", "todas"),
  modo: leer("ca.modo", "cuadricula"),
  orden: leer("ca.orden", "suite"),
  busqueda: "",
  bannerOculto: "",
  ultimoAviso: 0,
  firma: "",
};

// Fechas y «hace…» en el idioma elegido (los formatos los pone el propio Windows/Edge).
const formatoFecha = (iso) => new Intl.DateTimeFormat(localeActual(), { day: "numeric", month: "short", year: "numeric" })
  .format(new Date(iso)).replace(".", "");

function haceCuanto(iso) {
  const s = Math.max(0, (Date.now() - new Date(iso)) / 1000);
  if (s < 60) return t("haceSegundos");
  const rtf = new Intl.RelativeTimeFormat(localeActual(), { style: "short" });
  if (s < 3600) return rtf.format(-Math.floor(s / 60), "minute");
  if (s < 86400) return rtf.format(-Math.floor(s / 3600), "hour");
  return rtf.format(-Math.floor(s / 86400), "day");
}

// Descripción y lema de una app en el idioma elegido (si no hay traducción, en español).
const descripcion = (a) => a.traducciones?.[idiomaActual]?.descripcion || a.descripcion;
const similar = (a) => a.traducciones?.[idiomaActual]?.similar || a.similar;

/* ---------------- Colores por app ---------------- */

function rgb(hex) {
  const n = parseInt(hex.replace("#", ""), 16);
  return [(n >> 16) & 255, (n >> 8) & 255, n & 255];
}
function luminancia(hex) {
  const [r, g, b] = rgb(hex).map((c) => {
    c /= 255;
    return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
  });
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}
// Tinta legible sobre fondo oscuro: aclara los colores muy oscuros.
function tintaOscura(hex) {
  if (luminancia(hex) > 0.18) return hex;
  const [r, g, b] = rgb(hex).map((c) => Math.round(c + (255 - c) * 0.4));
  return `rgb(${r}, ${g}, ${b})`;
}
const textoSobre = (hex) => (luminancia(hex) > 0.4 ? "#0b0b0c" : "#ffffff");

/* ---------------- Datos derivados ---------------- */

const ORDEN_ESTADO = { actualizacion: 0, instalada: 1, disponible: 2, "sin-version": 3, proximamente: 4 };
const TAREAS = ["instalar", "actualizar", "desinstalar"];

function pasaFiltro(a) {
  switch (ui.filtro) {
    case "instaladas": return a.instalada;
    case "actualizaciones": return a.estado === "actualizacion";
    case "no-instaladas": return !a.instalada;
    default: return true;
  }
}

function pasaBusqueda(a) {
  const q = ui.busqueda.trim().toLowerCase();
  if (!q) return true;
  return [a.nombre + a.sufijo, descripcion(a), similar(a)].join(" ").toLowerCase().includes(q);
}

function grupos(apps) {
  const visibles = apps.filter((a) => pasaFiltro(a) && pasaBusqueda(a));
  if (ui.orden === "suite") {
    return ui.estado.categorias
      .map((c) => ({ nombre: c.nombre, apps: visibles.filter((a) => a.categoria === c.id) }))
      .concat([{ nombre: t("grupo.otras"), apps: visibles.filter((a) => !ui.estado.categorias.some((c) => c.id === a.categoria)) }])
      .filter((g) => g.apps.length);
  }
  const ordenadas = [...visibles];
  if (ui.orden === "nombre") ordenadas.sort((a, b) => (a.nombre + a.sufijo).localeCompare(b.nombre + b.sufijo, localeActual()));
  if (ui.orden === "recientes") ordenadas.sort((a, b) => new Date(b.ultima?.fecha || 0) - new Date(a.ultima?.fecha || 0));
  if (ui.orden === "estado") ordenadas.sort((a, b) => ORDEN_ESTADO[a.estado] - ORDEN_ESTADO[b.estado]);
  return ordenadas.length ? [{ nombre: t("grupo.todas"), apps: ordenadas }] : [];
}

/* ---------------- Piezas comunes ---------------- */

function iconoApp(a) {
  const caja = el("div", { class: "icono-app", style: { background: `linear-gradient(145deg, ${a.color}, ${tintaOscura(a.color)})`, color: textoSobre(a.color) } });
  caja.textContent = a.nombre.slice(0, 1).toUpperCase();
  if (a.icono) {
    // Icono oficial de la app; si no carga, queda la inicial.
    const img = el("img", { src: a.icono, alt: "", onerror: () => img.remove(), onload: () => { caja.textContent = ""; caja.append(img); caja.classList.add("con-imagen"); } });
  }
  return caja;
}

const credito = (a) => a.autor ? el("p", { class: "autor", text: t("de", { autor: a.autor }) }) : null;

function nombreApp(a, etiqueta = "h3") {
  return el(etiqueta, { class: "nombre" }, a.nombre, el("span", { text: a.sufijo, style: { color: a.colorSufijo || a.color } }));
}

function pastillaEstado(a) {
  if (a.tarea) {
    const pct = a.tarea.progreso >= 0 && a.tarea.texto === "descargando" ? ` · ${Math.round(a.tarea.progreso * 100)} %` : "";
    return el("span", { class: "estado estado-tarea", text: `${t("tarea." + a.tarea.texto)}${pct}` });
  }
  return el("span", { class: `estado estado-${a.estado}`, text: t("estado." + a.estado) });
}

function lineaVersion(a) {
  if (a.error && !a.ultima) return el("p", { class: "version" }, el("span", { class: "error", text: tm(a.error) }));
  if (!a.ultima) return el("p", { class: "version", text: t(a.estado === "proximamente" ? "version.aunSin" : "version.sin") });
  const fecha = a.ultima.fecha ? ` · ${formatoFecha(a.ultima.fecha)}` : "";
  if (a.estado === "actualizacion") return el("p", { class: "version", text: `${t("version", { version: a.versionInstalada })} → ${a.ultima.version}${fecha}` });
  return el("p", { class: "version", text: `${t("version", { version: a.instalada ? a.versionInstalada : a.ultima.version })}${fecha}` });
}

function botonPrincipal(a) {
  if (a.tarea) {
    const b = el("button", { class: "boton boton-contorno", type: "button", disabled: true });
    const p = a.tarea.progreso >= 0 ? a.tarea.progreso : 1;
    b.append(el("span", { class: "relleno", style: { width: `${Math.round(p * 100)}%` } }), el("span", { text: t(TAREAS.includes(a.tarea.tipo) ? "boton." + a.tarea.tipo : "boton.trabajando") }));
    return b;
  }
  // El botón se desactiva mientras dura la acción (evita el doble clic) y se vuelve a activar
  // al terminar: abrir una app no cambia su estado, así que la tarjeta no se redibuja sola.
  const accion = (ruta) => async (ev) => {
    const boton = ev.currentTarget;
    boton.disabled = true;
    try { await api(ruta, { id: a.id }); } catch (e) { avisar(e.message, true); }
    boton.disabled = false;
    refrescar();
  };
  switch (a.estado) {
    case "instalada": return el("button", { class: "boton boton-contorno", type: "button", onclick: accion("/api/abrir") }, el("span", { text: t("abrir") }));
    case "actualizacion": return el("button", { class: "boton boton-claro", type: "button", onclick: accion("/api/instalar") }, el("span", { text: t("actualizar") }));
    case "disponible": return el("button", { class: "boton boton-claro", type: "button", onclick: accion("/api/instalar") }, el("span", { text: t("instalar") }));
    default: return el("button", { class: "boton boton-contorno", type: "button", disabled: true }, el("span", { text: t(a.estado === "proximamente" ? "proximamente" : "noDisponible") }));
  }
}

function botonMenu(a) {
  return el("button", { class: "boton-cuadro", type: "button", "aria-label": t("masOpciones", { app: a.nombre + a.sufijo }), onclick: (ev) => abrirMenu(a, ev.currentTarget) },
    el("span", { text: "…" }));
}

/* ---------------- Vistas ---------------- */

// numeroApp: el número de la app en su colección (ArtCraft numera las suyas: 01, 02...).
function numeroApp(a, clase) {
  if (!a.numero) return null;
  return el("span", { class: clase, text: String(a.numero).padStart(2, "0"), style: { background: a.color, color: textoSobre(a.color) } });
}

function tarjeta(a) {
  const enlaces = el("div", { class: "tarjeta-enlaces" },
    a.web && el("button", { type: "button", text: t("sitioWeb"), onclick: () => abrirEnlace(a.web) }),
    a.urlRepo && el("button", { type: "button", text: t("github"), onclick: () => abrirEnlace(a.urlRepo) }));
  const estilo = { "--color-app": a.color };
  if (a.numero) estilo["--color-barra"] = a.color; // apps de ArtCraft: barra de su color
  return el("article", { class: "tarjeta", style: estilo },
    el("div", { class: "tarjeta-cabecera" },
      numeroApp(a, "numero"),
      enlaces),
    el("div", { class: "tarjeta-cuerpo" },
      el("div", { class: "identidad" }, iconoApp(a),
        el("div", {}, nombreApp(a), el("p", { class: "descripcion", text: descripcion(a) }), similar(a) && el("p", { class: "similar", text: similar(a) }), credito(a))),
      pastillaEstado(a),
      lineaVersion(a),
      el("div", { class: "acciones" }, botonPrincipal(a), botonMenu(a))));
}

function filaTabla(a) {
  const instalada = a.instalada ? a.versionInstalada || "—" : "—";
  const ultima = a.ultima ? a.ultima.version : "—";
  const fecha = a.ultima?.fecha ? formatoFecha(a.ultima.fecha) : "—";
  return el("tr", { style: { "--color-app": a.color } },
    el("td", {}, el("div", { class: "celda-app" }, el("span", { class: "hueco-numero" }, numeroApp(a, "numero")), iconoApp(a),
      el("div", {}, nombreApp(a, "div"), el("div", { class: "similar", text: a.autor ? `${descripcion(a)} · ${t("de", { autor: a.autor })}` : descripcion(a) })))),
    el("td", {}, pastillaEstado(a)),
    el("td", { class: a.instalada ? "ver" : null, text: instalada }),
    el("td", { class: a.ultima ? "ver" : null, text: ultima }),
    el("td", { text: fecha }),
    el("td", {}, el("div", { class: "acciones" }, botonPrincipal(a), botonMenu(a))));
}

function pintarContenido() {
  const e = ui.estado;
  const lista = grupos(e.apps);
  const contenido = $("#contenido");
  contenido.replaceChildren();
  if (!lista.length) {
    contenido.append(el("p", { class: "vacio", text: t(ui.busqueda ? "vacio.busqueda" : "vacio.filtro") }));
    return;
  }
  if (ui.modo === "tabla") {
    contenido.append(el("div", { class: "tabla-caja" }, el("table", { class: "tabla" },
      el("thead", {}, el("tr", {}, ...["tabla.app", "tabla.estado", "tabla.instalada", "tabla.ultima", "tabla.publicada"].map((c) => el("th", { text: t(c) })), el("th"))),
      el("tbody", {}, lista.flatMap((g) => g.apps).map(filaTabla)))));
    return;
  }
  lista.forEach((g, i) => {
    contenido.append(el("div", { class: "seccion" },
      el("span", {}, el("b", { text: `${String(i + 1).padStart(2, "0")} / ` }), g.nombre),
      el("span", { text: g.apps.length === 1 ? t("apps.una") : t("apps.varias", { n: g.apps.length }) })));
    contenido.append(el("div", { class: "rejilla" }, g.apps.map(tarjeta)));
  });
}

// textosAjustes: las opciones de los selectores de ajustes, que dependen del idioma.
function textosAjustes() {
  $("#aj-idioma").replaceChildren(el("option", { value: "auto", text: t("aj.automatico") }),
    ...Object.entries(IDIOMAS).map(([id, i]) => el("option", { value: id, text: i.nombre, lang: id })));
}

function pintar() {
  const e = ui.estado;
  if (!e) return;

  // Idioma: el que eligió el usuario o, en automático, el del PC (Codex lo resuelve).
  if (e.idioma !== idiomaActual || !$("#aj-idioma").options.length) {
    aplicarIdioma(e.idioma);
    textosAjustes();
    ui.firma = "";
  }
  document.body.classList.remove("cargando");

  // Banner de nueva versión del launcher
  const nueva = e.hayNueva && e.propia && ui.bannerOculto !== e.propia.version;
  $("#banner").hidden = !nueva;
  if (nueva) {
    $("#banner-texto").textContent = t(e.tareaPropia ? "banner.actualizando" : "banner.disponible", { version: e.propia.version });
    $("#banner-actualizar").disabled = !!e.tareaPropia;
  }

  // Cuentas de los filtros
  const cuentas = {
    todas: e.apps.length,
    instaladas: e.apps.filter((a) => a.instalada).length,
    actualizaciones: e.apps.filter((a) => a.estado === "actualizacion").length,
    "no-instaladas": e.apps.filter((a) => !a.instalada).length,
  };
  $$("#filtros button").forEach((b) => {
    b.querySelector(".cuenta").textContent = cuentas[b.dataset.filtro];
    b.classList.toggle("activo", b.dataset.filtro === ui.filtro);
  });
  $$(".segmento-vista button").forEach((b) => b.classList.toggle("activo", b.dataset.modo === ui.modo));
  document.body.classList.toggle("modo-tabla", ui.modo === "tabla");

  // Solo se vuelve a dibujar el contenido si algo cambió (no se pierde el foco ni el hover).
  const firma = JSON.stringify([e.apps, ui.filtro, ui.modo, ui.orden, ui.busqueda, idiomaActual]);
  if (firma !== ui.firma) {
    ui.firma = firma;
    pintarContenido();
  }

  // Pie
  $("#btn-revisar").classList.toggle("girando", e.revisando);
  const pausa = e.limiteHasta && new Date(e.limiteHasta) > new Date()
    ? new Date(e.limiteHasta).toLocaleTimeString(localeActual(), { hour: "2-digit", minute: "2-digit" }) : "";
  $("#pie-revisado").textContent = e.revisando ? t("pie.buscando")
    : pausa ? t("pie.pausa", { hora: pausa })
    : e.revisado ? t("pie.revisado", { cuando: haceCuanto(e.revisado) }) : t("pie.sinRevisar");
  $("#pie-plataforma").textContent = t("pie.plataforma", { plataforma: e.plataforma });

  // Ajustes
  $("#aj-auto").checked = e.ajustes.revisarAuto;
  // El interruptor y el selector dicen lo mismo: activado = cada 24 horas, apagado = manualmente.
  $("#aj-intervalo").value = e.ajustes.revisarAuto ? "24" : "manual";
  $("#aj-idioma").value = e.ajustes.idioma || "auto";
  $("#aj-bandeja").checked = e.ajustes.bandeja;
  $("#aj-inicio").checked = e.ajustes.iniciarSesion;
  $("#aj-version").textContent = e.version;
  $("#aj-carpeta").textContent = e.carpeta;
  $("#aj-plataforma").textContent = e.plataforma;

  if (e.aviso.id && e.aviso.id !== ui.ultimoAviso) {
    if (ui.ultimoAviso !== 0 || e.aviso.error) avisar(tm(e.aviso.mensaje), e.aviso.error);
    ui.ultimoAviso = e.aviso.id;
  }
}

/* ---------------- Ciclo de refresco ---------------- */

let temporizador;
async function refrescar() {
  clearTimeout(temporizador);
  let ocupado = false;
  try {
    ui.estado = await api("/api/estado");
    if (ui.ultimoAviso === 0) ui.ultimoAviso = ui.estado.aviso.id || -1;
    pintar();
    ocupado = ui.estado.revisando || !!ui.estado.tareaPropia || ui.estado.apps.some((a) => a.tarea);
  } catch (e) {
    avisar(t("errorEstado", { detalle: e.message }), true);
    document.body.classList.remove("cargando");
  }
  temporizador = setTimeout(refrescar, ocupado ? 700 : 10000);
}

/* ---------------- Menú «…» ---------------- */

function cerrarMenu() { $("#menu").hidden = true; }

function abrirMenu(a, boton) {
  const menu = $("#menu");
  const opciones = [];
  if (a.ultima?.notas) opciones.push([t("menu.notas"), () => abrirEnlace(a.ultima.notas)]);
  if (a.urlRepo) opciones.push([t("menu.github"), () => abrirEnlace(a.urlRepo)]);
  if (a.web) opciones.push([t("menu.sitio"), () => abrirEnlace(a.web)]);
  if (a.instalada && !a.tarea) {
    if (a.ultima?.instalador) opciones.push([t("menu.reinstalar"), () => api("/api/instalar", { id: a.id }).then(refrescar, (e) => avisar(e.message, true))]);
    opciones.push([t("menu.desinstalar"), () => api("/api/desinstalar", { id: a.id }).then(refrescar, (e) => avisar(e.message, true)), true]);
  }
  if (!opciones.length) opciones.push([t("menu.sinOpciones"), null]);
  menu.replaceChildren(...opciones.map(([texto, f, peligro]) =>
    el("button", { type: "button", role: "menuitem", class: peligro ? "peligro" : null, text: texto, disabled: !f, onclick: () => { cerrarMenu(); f && f(); } })));
  menu.hidden = false;
  const r = boton.getBoundingClientRect();
  const ancho = menu.offsetWidth, alto = menu.offsetHeight;
  menu.style.left = `${Math.min(r.right - ancho, innerWidth - ancho - 12)}px`;
  menu.style.top = `${r.bottom + 6 + alto > innerHeight ? r.top - alto - 6 : r.bottom + 6}px`;
  menu.querySelector("button:not(:disabled)")?.focus();
}

document.addEventListener("click", (ev) => {
  if (!$("#menu").hidden && !ev.target.closest("#menu, .boton-cuadro")) cerrarMenu();
});
document.addEventListener("keydown", (ev) => {
  if (ev.key === "Escape") { cerrarMenu(); cerrarAjustes(); }
});

/* ---------------- Acciones generales ---------------- */

function abrirEnlace(url) {
  api("/api/enlace", { url }).catch((e) => avisar(e.message, true));
}

let temporizadorAviso;
function avisar(texto, esError = false) {
  const a = $("#aviso");
  a.textContent = texto;
  a.classList.toggle("error", esError);
  a.hidden = false;
  clearTimeout(temporizadorAviso);
  temporizadorAviso = setTimeout(() => { a.hidden = true; }, esError ? 7000 : 4500);
}

$$(".segmento-vista button").forEach((b) => b.addEventListener("click", () => { ui.modo = b.dataset.modo; guardar("ca.modo", ui.modo); pintar(); }));
$$("#filtros button").forEach((b) => b.addEventListener("click", () => { ui.filtro = b.dataset.filtro; guardar("ca.filtro", ui.filtro); pintar(); }));
$("#orden").value = ui.orden;
$("#orden").addEventListener("change", (ev) => { ui.orden = ev.target.value; guardar("ca.orden", ui.orden); pintar(); });
$("#buscar").addEventListener("input", (ev) => { ui.busqueda = ev.target.value; pintar(); });

$("#btn-revisar").addEventListener("click", async () => {
  try { await api("/api/revisar", {}); } catch (e) { avisar(e.message, true); }
  setTimeout(refrescar, 150);
});

$("#banner-cerrar").addEventListener("click", () => { ui.bannerOculto = ui.estado?.propia?.version || ""; pintar(); });
$("#banner-novedades").addEventListener("click", () => ui.estado?.propia && abrirEnlace(ui.estado.propia.notas));
$("#banner-actualizar").addEventListener("click", async () => {
  try { await api("/api/launcher/actualizar", {}); } catch (e) { avisar(e.message, true); }
  refrescar();
});

/* Ajustes */
function abrirAjustes() { $("#velo").hidden = false; $("#ajustes").hidden = false; $("#aj-auto").focus(); }
function cerrarAjustes() { $("#velo").hidden = true; $("#ajustes").hidden = true; }
$("#btn-ajustes").addEventListener("click", abrirAjustes);
$("#ajustes-cerrar").addEventListener("click", cerrarAjustes);
$("#velo").addEventListener("click", cerrarAjustes);
async function guardarAjustes() {
  const a = {
    revisarAuto: $("#aj-auto").checked, intervaloHoras: 24,
    idioma: $("#aj-idioma").value, bandeja: $("#aj-bandeja").checked, iniciarSesion: $("#aj-inicio").checked,
  };
  try { await api("/api/ajustes", a); } catch (e) { avisar(e.message, true); }
  refrescar();
}
// Elegir «manualmente» apaga el interruptor y «cada 24 horas» lo enciende (y al revés).
$("#aj-intervalo").addEventListener("change", (ev) => { $("#aj-auto").checked = ev.target.value === "24"; guardarAjustes(); });
$("#aj-auto").addEventListener("change", () => { $("#aj-intervalo").value = $("#aj-auto").checked ? "24" : "manual"; guardarAjustes(); });
["#aj-idioma", "#aj-bandeja", "#aj-inicio"].forEach((s) => $(s).addEventListener("change", guardarAjustes));

/* ---------------- Ventana (solo dentro del launcher, no en un navegador) ---------------- */

if (typeof window.caArrastrar === "function") {
  $("#controles").hidden = false;
  const barra = $("#barra");
  const interactivo = "button, a, input, select, nav button";
  barra.addEventListener("mousedown", (ev) => {
    if (ev.button !== 0 || ev.detail > 1 || ev.target.closest(interactivo)) return;
    window.caArrastrar();
  });
  barra.addEventListener("dblclick", (ev) => {
    if (!ev.target.closest(interactivo)) window.caMaximizar().then(actualizarMaximizada);
  });
  $$(".redim").forEach((r) => r.addEventListener("mousedown", (ev) => {
    if (ev.button === 0) window.caRedimensionar(r.dataset.borde);
  }));
  $("#btn-min").addEventListener("click", () => window.caMinimizar());
  $("#btn-max").addEventListener("click", () => window.caMaximizar().then(actualizarMaximizada));
  $("#btn-cerrar").addEventListener("click", () => window.caCerrar());
  const actualizarMaximizada = async () => {
    const max = await window.caMaximizada();
    document.body.classList.toggle("maximizada", max);
    $("#btn-max").setAttribute("aria-label", t(max ? "restaurar" : "maximizar"));
  };
  window.addEventListener("resize", actualizarMaximizada);
  actualizarMaximizada();
} else {
  $$(".redim").forEach((r) => r.remove());
}

refrescar();
