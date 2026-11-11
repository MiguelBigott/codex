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

// Versión arriba a la izquierda, en la cabecera de la tarjeta: solo el número.
function versionChip(a) {
  if (!a.ultima) return el("span", { class: "version-chip" });
  // En actualización se muestra la versión que se va a instalar; el estado ya dice que hay una nueva.
  if (a.estado === "actualizacion") return el("span", { class: "version-chip", text: `v${a.ultima.version}` });
  return el("span", { class: "version-chip", text: `v${a.instalada ? a.versionInstalada : a.ultima.version}` });
}

// Debajo de la descripción solo queda la fecha (o el error / el aviso de que no hay versión).
function lineaVersion(a) {
  if (a.error && !a.ultima) return el("p", { class: "version" }, el("span", { class: "error", text: tm(a.error) }));
  if (!a.ultima) return el("p", { class: "version", text: t(a.estado === "proximamente" ? "version.aunSin" : "version.sin") });
  // Las apps que se actualizan solas pueden ir por delante de la versión que conoce Codex:
  // instaladas, no se muestra la fecha de otra.
  if (a.seActualizaSola && a.instalada) return el("p", { class: "version" });
  return el("p", { class: "version", text: a.ultima.fecha ? formatoFecha(a.ultima.fecha) : "" });
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
    el("span", { class: "puntos", "aria-hidden": "true" }, el("i", {}), el("i", {}), el("i", {})));
}

/* ---------------- Vistas ---------------- */

function tarjeta(a) {
  const enlaces = el("div", { class: "tarjeta-enlaces" },
    a.web && el("button", { type: "button", text: t("sitioWeb"), onclick: () => abrirEnlace(a.web) }),
    a.urlRepo && el("button", { type: "button", text: t("github"), onclick: () => abrirEnlace(a.urlRepo) }));
  const estilo = { "--color-app": a.color };
  if (a.categoria === "artcraft") estilo["--color-barra"] = a.color; // apps de ArtCraft: barra de su color
  // Un clic en la tarjeta (fuera de sus botones) abre el panel de detalles; con teclado, Enter.
  return el("article", {
    class: "tarjeta", style: estilo, "data-id": a.id, tabindex: "0",
    "aria-label": t("det.ver", { app: a.nombre + a.sufijo }),
    onclick: (ev) => { if (!ev.target.closest("button, a")) abrirDetalles(a.id); },
    onkeydown: (ev) => { if ((ev.key === "Enter" || ev.key === " ") && ev.target === ev.currentTarget) { ev.preventDefault(); abrirDetalles(a.id); } },
  },
    el("div", { class: "tarjeta-cabecera" }, versionChip(a), enlaces),
    // Placa: icono con el nombre a su derecha, texto debajo, estado y fecha en una línea y una
    // barra de botones al pie.
    el("div", { class: "tarjeta-cuerpo" },
      el("div", { class: "tarjeta-titulo" }, iconoApp(a), nombreApp(a)),
      el("div", { class: "tarjeta-texto" }, el("p", { class: "descripcion", text: descripcion(a) }), similar(a) && el("p", { class: "similar", text: similar(a) }), credito(a)),
      el("div", { class: "tarjeta-estado" }, pastillaEstado(a), lineaVersion(a))),
    el("div", { class: "acciones" }, botonPrincipal(a), botonMenu(a)));
}

// indice: posición de la fila en la tabla, desde la primera (ClipVault) hacia abajo.
function filaTabla(a, indice) {
  const instalada = a.instalada ? a.versionInstalada || "—" : "—";
  const ultima = a.ultima ? a.ultima.version : "—";
  const fecha = a.ultima?.fecha ? formatoFecha(a.ultima.fecha) : "—";
  return el("tr", { style: { "--color-app": a.color }, class: "fila-app", onclick: (ev) => { if (!ev.target.closest("button, a")) abrirDetalles(a.id); } },
    el("td", {}, el("div", { class: "celda-app" }, el("span", { class: "hueco-numero", text: String(indice + 1).padStart(2, "0") }), iconoApp(a),
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

  pintarDetalles();
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
  if (ev.key === "Escape") {
    if (!$("#menu").hidden) return cerrarMenu();
    if (ui.detalle) return cerrarDetalles();
    cerrarAjustes();
  }
  // En el panel de detalles, las flechas pasan a la app anterior o a la siguiente.
  if (ui.detalle && $("#menu").hidden && !ev.target.closest("input, select, textarea")) {
    if (ev.key === "ArrowLeft") navegarDetalles(-1);
    if (ev.key === "ArrowRight") navegarDetalles(1);
  }
});

/* ---------------- Panel de detalles ----------------
   Al hacer clic en una app (fuera de sus botones) se abre su página: titular, descripción,
   funciones y plataformas (de /api/detalles, que se pide al abrirla) y, de la última versión,
   número, fecha, tamaño y notas. El estado sigue llegando por /api/estado, así que los botones
   del panel (instalar, actualizar, abrir...) se ven al día. */

const cacheDetalles = new Map(); // "id@tag" → respuesta de /api/detalles (null mientras carga)

const claveDetalles = (a) => `${a.id}@${a.ultima?.tag || ""}`;
const appsEnOrden = () => grupos(ui.estado.apps).flatMap((g) => g.apps);

async function cargarDetalles(a) {
  const clave = claveDetalles(a);
  if (cacheDetalles.has(clave)) return;
  cacheDetalles.set(clave, null);
  try {
    cacheDetalles.set(clave, await api(`/api/detalles?id=${encodeURIComponent(a.id)}`));
  } catch (e) {
    cacheDetalles.set(clave, { error: e.message });
  }
  if (ui.detalle?.id === a.id) pintarDetalles();
}

function abrirDetalles(id) {
  cerrarMenu();
  const enfocar = !ui.detalle;
  ui.detalle = { id, firma: "" };
  pintarDetalles();
  const panel = $("#detalles .detalles");
  if (panel) panel.querySelector(".det-cuerpo").scrollTop = 0;
  if (enfocar) $("#detalles .det-cerrar")?.focus();
}

function cerrarDetalles() {
  const id = ui.detalle?.id;
  ui.detalle = null;
  $("#detalles")?.remove();
  document.querySelector(`.tarjeta[data-id="${CSS.escape(id || "")}"]`)?.focus();
}

function navegarDetalles(paso) {
  const lista = appsEnOrden();
  const i = lista.findIndex((x) => x.id === ui.detalle?.id);
  if (i < 0) return;
  abrirDetalles(lista[(i + paso + lista.length) % lista.length].id);
}

function botonNav(trazo, clase, texto, f) {
  const b = el("button", { type: "button", class: `det-nav ${clase}`, "aria-label": texto, title: texto, onclick: f });
  b.innerHTML = `<svg viewBox="0 0 24 24" aria-hidden="true"><path d="${trazo}"/></svg>`;
  return b;
}

// limpiarMd deja una línea de las notas de GitHub en texto plano: enlaces como su texto, los
// pedidos de cambio como #número y sin negritas ni etiquetas.
function limpiarMd(s) {
  return s
    .replace(/\[([^\]]+)\]\([^)]*\)/g, "$1")
    .replace(/\s+by @[\w-]+ in https?:\/\/github\.com\/\S+?\/(?:pull|issues)\/(\d+)/g, " #$1")
    .replace(/https?:\/\/github\.com\/\S+?\/(?:pull|issues)\/(\d+)/g, "#$1")
    .replace(/<[^>]+>/g, "")
    .replace(/\*\*|__|`/g, "")
    .trim();
}

// entradasMd convierte las notas en títulos, viñetas y párrafos.
function entradasMd(md) {
  const entradas = [];
  for (const crudo of (md || "").split("\n")) {
    const linea = crudo.trim();
    if (!linea || /^(-{3,}|<!--.*)$/.test(linea) || /^\**full changelog/i.test(linea)) continue;
    let m;
    if ((m = linea.match(/^#{1,6}\s+(.*)$/))) entradas.push({ tipo: "h", texto: limpiarMd(m[1]) });
    else if ((m = linea.match(/^[-*+]\s+(.*)$/))) entradas.push({ tipo: "li", texto: limpiarMd(m[1]) });
    else entradas.push({ tipo: "p", texto: limpiarMd(linea) });
  }
  // Codex no nombra al sitio de videos ni a su motor de descarga, tampoco en notas de terceros.
  return entradas.filter((e) => e.texto && !/youtube|yt-dlp/i.test(e.texto));
}

// bloquesMd arma los nodos: las viñetas seguidas van en una misma lista.
function bloquesMd(entradas) {
  const nodos = [];
  let lista = null;
  for (const e of entradas) {
    if (e.tipo === "li") {
      if (!lista) nodos.push(lista = el("ul"));
      lista.append(el("li", { text: e.texto }));
      continue;
    }
    lista = null;
    nodos.push(el(e.tipo === "h" ? "h4" : "p", { text: e.texto }));
  }
  return nodos;
}

const VINETAS_VISIBLES = 6;

function notasDeVersion(md, abiertas) {
  const entradas = entradasMd(md);
  if (!entradas.length) return el("p", { class: "det-vacio", text: t("det.sinNotas") });
  // Se ven las primeras viñetas; el resto queda tras «Mostrar N cambios más».
  let vistas = 0, corte = entradas.length;
  for (let i = 0; i < entradas.length; i++) {
    if (entradas[i].tipo === "li" && ++vistas > VINETAS_VISIBLES) { corte = i; break; }
  }
  const resto = entradas.slice(corte);
  const restantes = resto.filter((e) => e.tipo === "li").length;
  return el("div", { class: "det-notas" }, bloquesMd(entradas.slice(0, corte)),
    resto.length ? el("details", { open: abiertas || null },
      el("summary", { text: t("det.masCambios", { n: restantes || resto.length }) }), bloquesMd(resto)) : null);
}

// idiomaDetalles: el texto del panel sale en español si Codex está en español (por defecto, el
// idioma del PC) y en inglés en los demás; el botón ES / EN lo cambia mientras Codex esté abierto.
const idiomaDetalles = () => ui.idiomaDetalles || (idiomaActual === "es" ? "es" : "en");

// detalleEnIdioma devuelve la página en ese idioma: la original o su traducción. Las plataformas
// son las mismas en los dos.
function detalleEnIdioma(d) {
  if (!d) return null;
  const idioma = idiomaDetalles();
  const x = d.idioma === idioma ? d : d.traducciones?.[idioma];
  return x ? { ...d, ...x, plataformas: x.plataformas || d.plataformas } : d;
}

function selectorIdioma() {
  const actual = idiomaDetalles();
  return el("div", { class: "det-idioma", role: "group", "aria-label": t("det.idioma") },
    ["es", "en"].map((idioma) => el("button", {
      type: "button", lang: idioma, text: idioma.toUpperCase(), class: idioma === actual ? "activo" : null,
      "aria-pressed": String(idioma === actual),
      onclick: () => { ui.idiomaDetalles = idioma; pintarDetalles(); },
    })));
}

function pintarDetalles() {
  if (!ui.detalle || !ui.estado) return;
  const a = ui.estado.apps.find((x) => x.id === ui.detalle.id);
  if (!a) return cerrarDetalles();
  const clave = claveDetalles(a);
  if (!cacheDetalles.has(clave)) cargarDetalles(a);
  const datos = cacheDetalles.get(clave);
  // Solo se vuelve a dibujar si algo cambió: así no se pierde el desplazamiento ni el foco.
  const firma = JSON.stringify([a, datos ? Object.keys(datos).length : 0, datos?.cambios?.length, idiomaActual, idiomaDetalles()]);
  if (firma === ui.detalle.firma) return;
  ui.detalle.firma = firma;

  const viejo = $("#detalles");
  const desplazado = viejo?.querySelector(".det-cuerpo")?.scrollTop || 0;
  const notasAbiertas = !!viejo?.querySelector(".det-notas details")?.open;
  const enfocado = viejo?.contains(document.activeElement) ? document.activeElement.className : "";

  const d = detalleEnIdioma(datos?.detalle);
  const ultima = a.ultima;
  const fecha = ultima?.fecha && !(a.seActualizaSola && a.instalada) ? formatoFecha(ultima.fecha) : "";
  const lineaVer = ultima ? [t("version", { version: ultima.version }), fecha].filter(Boolean).join(" · ") : t("version.sin");
  // Sin página propia, la descripción ya va como titular: abajo queda solo el lema.
  const parrafos = d ? d.acerca.split(/\n\s*\n/) : [similar(a)].filter(Boolean);
  const tamano = ultima?.instalador?.tamano ? `${Math.max(1, Math.round(ultima.instalador.tamano / 1048576))} MB` : "—";
  const enlaces = [
    a.web && [t("sitioWeb"), a.web],
    a.urlRepo && [t("github"), a.urlRepo],
    a.urlRepo && [t("det.versiones"), `${a.urlRepo}/releases`],
  ].filter(Boolean);

  const notas = !datos ? el("p", { class: "det-vacio", text: t("det.cargando") })
    : datos.error ? el("p", { class: "det-vacio", text: datos.error })
    : notasDeVersion(datos.cambios, notasAbiertas);

  const panel = el("div", { id: "detalles" },
    el("div", { class: "velo-detalles", onclick: cerrarDetalles }),
    el("section", { class: "detalles", role: "dialog", "aria-modal": "true", "aria-labelledby": "det-nombre", style: { "--color-app": a.color } },
      el("header", { class: "det-cabecera" },
        el("div", { class: "det-barra" },
          el("span", { class: "det-etiqueta", text: t("det.etiqueta") }),
          el("div", { class: "det-navegacion" },
            d ? selectorIdioma() : null,
            botonNav("M15 6l-6 6 6 6", "det-anterior", t("det.anterior"), () => navegarDetalles(-1)),
            botonNav("M9 6l6 6-6 6", "det-siguiente", t("det.siguiente"), () => navegarDetalles(1)),
            botonNav("M7 7l10 10M17 7L7 17", "det-cerrar", t("det.cerrar"), cerrarDetalles))),
        el("div", { class: "det-identidad" },
          iconoApp(a),
          el("div", { class: "det-titulos" },
            Object.assign(nombreApp(a, "h2"), { id: "det-nombre" }),
            el("p", { class: "det-titular", text: d?.titular || descripcion(a) }),
            el("p", { class: "det-version", text: lineaVer }),
            a.autor ? el("p", { class: "det-autor", text: t("de", { autor: a.autor }) }) : null),
          el("div", { class: "det-acciones acciones" }, botonPrincipal(a), botonMenu(a))),
        enlaces.length ? el("div", { class: "det-enlaces" },
          enlaces.map(([texto, url]) => el("button", { type: "button", text: texto, onclick: () => abrirEnlace(url) }))) : null),
      el("div", { class: "det-cuerpo" },
        el("div", { class: "det-izq" },
          el("div", { class: "det-acerca" }, parrafos.map((p) => el("p", { text: p }))),
          d?.funciones?.length ? el("h3", { class: "det-h", text: t("det.funciones") }) : null,
          d?.funciones?.length ? el("ol", { class: "det-funciones" }, d.funciones.map((f, i) =>
            el("li", {}, el("span", { class: "num", text: String(i + 1).padStart(2, "0") }), el("b", { text: f.titulo }), el("p", { text: f.texto })))) : null),
        el("div", { class: "det-der" },
          el("h3", { class: "det-h", text: t("det.ultima") }),
          el("dl", { class: "det-datos" }, [
            [t("det.version"), ultima?.version || "—"],
            [t("det.publicada"), ultima?.fecha ? formatoFecha(ultima.fecha) : "—"],
            [t("det.descarga"), tamano],
            [t("det.plataformas"), d?.plataformas || "Windows"],
          ].map(([k, v]) => el("div", {}, el("dt", { text: k }), el("dd", { text: v })))),
          el("h3", { class: "det-h", text: t("det.novedades") }),
          notas,
          ultima?.notas ? el("button", { type: "button", class: "det-registro", text: t("det.registro"), onclick: () => abrirEnlace(ultima.notas) }) : null))));

  if (viejo) viejo.replaceWith(panel);
  else document.body.append(panel);
  panel.querySelector(".det-cuerpo").scrollTop = desplazado;
  if (enfocado) panel.querySelector(`.${enfocado.split(" ").filter(Boolean).join(".")}`)?.focus();
}

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
