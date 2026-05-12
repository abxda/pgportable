// ─── strings (i18n stub) ──────────────────────────────────────
const STR = {
  status: {
    stopped:  "Detenido",
    starting: "Iniciando…",
    running:  "Corriendo",
    stopping: "Deteniendo…",
    error:    "Error",
  },
  toast: {
    copied:    "Copiado al portapapeles",
    started:   "PostgreSQL listo",
    stopped:   "PostgreSQL detenido",
    noBinaries:"Falta la carpeta pgsql/ con los binarios",
  },
  tree: {
    emptyStopped: "Arranca el entorno para explorar.",
    emptyRunning: "Cargando…",
    noDatabases:  "No hay bases de datos visibles.",
    noSchemas:    "Sin esquemas.",
    noTables:     "Sin tablas.",
  },
};

// ─── icons (Lucide-style inline SVG) ─────────────────────────
const SVG = (path) =>
  `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">${path}</svg>`;

const ICONS = {
  database: SVG(`<ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M3 5v14a9 3 0 0 0 18 0V5"/><path d="M3 12a9 3 0 0 0 18 0"/>`),
  play:     SVG(`<polygon points="6 3 20 12 6 21 6 3"/>`),
  stop:     SVG(`<rect x="6" y="6" width="12" height="12" rx="1.5"/>`),
  copy:     SVG(`<rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>`),
  key:      SVG(`<circle cx="7.5" cy="15.5" r="4.5"/><path d="m10.5 12.5 9-9"/><path d="m17 5 3 3"/><path d="m14 8 3 3"/>`),
  terminal: SVG(`<polyline points="4 17 10 11 4 5"/><line x1="12" x2="20" y1="19" y2="19"/>`),
  disk:     SVG(`<rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="12" cy="12" r="3"/><line x1="12" x2="12" y1="3" y2="6"/>`),
  folder:   SVG(`<path d="M20 19a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h7a2 2 0 0 1 2 2z"/>`),
  tree:     SVG(`<line x1="6" x2="20" y1="6" y2="6"/><line x1="10" x2="20" y1="12" y2="12"/><line x1="10" x2="20" y1="18" y2="18"/><circle cx="3.5" cy="6" r="1.5"/><circle cx="6.5" cy="12" r="1.5"/><circle cx="6.5" cy="18" r="1.5"/>`),
  table:    SVG(`<rect x="3" y="3" width="18" height="18" rx="2"/><line x1="3" x2="21" y1="9" y2="9"/><line x1="3" x2="21" y1="15" y2="15"/><line x1="9" x2="9" y1="3" y2="21"/>`),
  schema:   SVG(`<path d="M2 9 12 4 22 9 12 14 2 9z"/><path d="M2 15 12 20 22 15"/>`),
};

const $ = (id) => document.getElementById(id);

let lastStatus = null;

// ─── boot ─────────────────────────────────────────────────────
window.addEventListener("DOMContentLoaded", () => {
  paintIcons();
  bindUI();
  refreshStatus();
  refreshStorage();
  attachRuntimeEvents();
});

function paintIcons() {
  $("brandIcon").innerHTML = ICONS.database;
  $("iconStart").innerHTML = ICONS.play;
  $("iconStop").innerHTML  = ICONS.stop;
  $("iconKey").innerHTML   = ICONS.key;
  $("iconDisk").innerHTML  = ICONS.disk;
  $("iconTree").innerHTML  = ICONS.tree;
  $("iconTerm").innerHTML  = ICONS.terminal;
  $("iconFolder").innerHTML= ICONS.folder;
  document.querySelectorAll(".c-copy, .c-btn--icon .ico").forEach(el => el.innerHTML = ICONS.copy);
}

function bindUI() {
  $("btnStart").addEventListener("click", onStart);
  $("btnStop").addEventListener("click",  onStop);
  $("btnClearLogs").addEventListener("click", () => { $("logBox").textContent = ""; });
  $("btnOpenFolder").addEventListener("click", () => window.go.main.App.OpenDataFolder());
  $("btnRefreshStorage").addEventListener("click", refreshStorage);
  $("btnRefreshTree").addEventListener("click", () => loadTree(true));

  // copia por campo
  document.querySelectorAll(".c-copy[data-copy]").forEach(b => {
    b.addEventListener("click", () => {
      if (!lastStatus) return;
      const k = b.dataset.copy;
      copy(String(lastStatus[k] ?? ""));
    });
  });

  // copia de inputs/textarea por id
  document.querySelectorAll("[data-copy-input]").forEach(b => {
    b.addEventListener("click", () => {
      const el = $(b.dataset.copyInput);
      if (el) copy(el.value);
    });
  });
}

function attachRuntimeEvents() {
  window.runtime.EventsOn("pg:status", (s) => {
    applyStatus(s);
    // si pasa a running, dispara primera carga del árbol
    if (s && s.state === "running") {
      loadTree(true);
      refreshStorage();
    }
    if (s && s.state === "stopped") {
      $("tree").innerHTML = "";
    }
  });
  window.runtime.EventsOn("pg:log",   appendLog);
  window.runtime.EventsOn("pg:error", (msg) => toast(msg, true));
}

// ─── actions ─────────────────────────────────────────────────
async function onStart() {
  setBusy(true);
  try { await window.go.main.App.Start(); }
  catch (e) { toast(String(e), true); }
  finally   { setBusy(false); }
}
async function onStop() {
  setBusy(true);
  try { await window.go.main.App.Stop(); }
  catch (e) { toast(String(e), true); }
  finally   { setBusy(false); }
}

async function refreshStatus() {
  try {
    const s = await window.go.main.App.GetStatus();
    applyStatus(s);
    if (s.state === "running") loadTree(true);
  } catch (e) { console.error("GetStatus", e); }
}

function applyStatus(s) {
  if (!s) return;
  lastStatus = s;
  const pill  = $("statusPill");
  const label = $("statusLabel");
  pill.dataset.state = s.state;
  label.textContent = (s.state === "running")
    ? `${STR.status.running} · puerto ${s.port}`
    : (STR.status[s.state] || s.state);

  $("fHost").textContent = s.host;
  $("fPort").textContent = s.port;
  $("fUser").textContent = s.user;
  $("fPass").textContent = s.password;
  $("fDb").textContent   = s.database;
  $("connUri").value     = s.connectionUri;
  $("jdbcUrl").value     = s.jdbcUrl;
  $("psqlCmd").value     = s.psqlCommand;
  $("envBlock").value    = s.envBlock;

  const startable = (s.state === "stopped" || s.state === "error");
  $("btnStart").disabled = !startable || !s.binariesPresent;
  $("btnStop").disabled  = (s.state !== "running");

  if (!s.binariesPresent) toast(STR.toast.noBinaries, true);
}

function setBusy(busy) {
  $("btnStart").disabled = busy;
  $("btnStop").disabled  = busy;
}

// ─── storage ─────────────────────────────────────────────────
async function refreshStorage() {
  try {
    const i = await window.go.main.App.GetStorageInfo();
    $("sDataDir").textContent = i.dataDir;
    $("sBinDir").textContent  = i.binDir;
    $("sLogFile").textContent = i.logFile;
    $("sSize").textContent    = i.exists ? i.sizeHuman : "—";
    $("sFiles").textContent   = i.exists ? i.fileCount : "(no inicializado)";
  } catch (e) { console.error("GetStorageInfo", e); }
}

// ─── tree explorer ───────────────────────────────────────────
async function loadTree(force = false) {
  const root = $("tree");
  if (!lastStatus || lastStatus.state !== "running") {
    root.innerHTML = "";
    root.dataset.empty = STR.tree.emptyStopped;
    return;
  }
  if (!force && root.children.length > 0) return;
  root.innerHTML = "";
  root.dataset.empty = STR.tree.emptyRunning;

  try {
    const dbs = await window.go.main.App.ListDatabases();
    if (!dbs || dbs.length === 0) {
      root.dataset.empty = STR.tree.noDatabases;
      return;
    }
    const ul = document.createElement("ul");
    dbs.forEach(db => ul.appendChild(makeDbNode(db)));
    root.appendChild(ul);
  } catch (e) {
    root.innerHTML = "";
    root.dataset.empty = "Error: " + e;
  }
}

function makeDbNode(db) {
  const li = document.createElement("li");
  const node = document.createElement("div");
  node.className = "node" + (db.current ? " node--current" : "");
  node.dataset.loaded = "false";
  node.innerHTML = `
    <span class="chev"></span>
    ${ICONS.database}
    <span>${escapeHtml(db.name)}</span>
    <span class="meta">${escapeHtml(db.size)} · ${escapeHtml(db.owner)}</span>`;
  node.addEventListener("click", async () => {
    if (node.dataset.loaded === "true") {
      // collapse
      const next = li.querySelector(":scope > ul");
      if (next) next.remove();
      node.dataset.loaded = "false";
      return;
    }
    const child = document.createElement("ul");
    child.innerHTML = `<li class="muted">Cargando esquemas…</li>`;
    li.appendChild(child);
    try {
      const schemas = await window.go.main.App.ListSchemas(db.name);
      child.innerHTML = "";
      if (!schemas || schemas.length === 0) {
        child.innerHTML = `<li><span class="meta">${STR.tree.noSchemas}</span></li>`;
      } else {
        schemas.forEach(s => child.appendChild(makeSchemaNode(db.name, s)));
      }
      node.dataset.loaded = "true";
    } catch (e) {
      child.innerHTML = `<li><span class="meta">Error: ${escapeHtml(String(e))}</span></li>`;
    }
  });
  li.appendChild(node);
  return li;
}

function makeSchemaNode(dbName, s) {
  const li = document.createElement("li");
  const node = document.createElement("div");
  node.className = "node";
  node.dataset.loaded = "false";
  node.innerHTML = `
    <span class="chev"></span>
    ${ICONS.schema}
    <span>${escapeHtml(s.name)}</span>`;
  node.addEventListener("click", async () => {
    if (node.dataset.loaded === "true") {
      const next = li.querySelector(":scope > ul");
      if (next) next.remove();
      node.dataset.loaded = "false";
      return;
    }
    const child = document.createElement("ul");
    child.innerHTML = `<li class="muted">Cargando tablas…</li>`;
    li.appendChild(child);
    try {
      const tables = await window.go.main.App.ListTables(dbName, s.name);
      child.innerHTML = "";
      if (!tables || tables.length === 0) {
        child.innerHTML = `<li><span class="meta">${STR.tree.noTables}</span></li>`;
      } else {
        tables.forEach(t => child.appendChild(makeTableNode(t)));
      }
      node.dataset.loaded = "true";
    } catch (e) {
      child.innerHTML = `<li><span class="meta">Error: ${escapeHtml(String(e))}</span></li>`;
    }
  });
  li.appendChild(node);
  return li;
}

function makeTableNode(t) {
  const li = document.createElement("li");
  const node = document.createElement("div");
  node.className = "node";
  node.dataset.loadable = "false";
  const rows = t.rows >= 0 ? `~${t.rows.toLocaleString()} filas` : "—";
  const type = t.type === "VIEW" ? "VIEW" : "TABLA";
  node.innerHTML = `
    <span class="chev"></span>
    ${ICONS.table}
    <span>${escapeHtml(t.name)}</span>
    <span class="meta">${type} · ${rows}</span>`;
  li.appendChild(node);
  return li;
}

function escapeHtml(s) {
  return String(s).replace(/[&<>"']/g, c => ({
    "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;"
  }[c]));
}

// ─── console ─────────────────────────────────────────────────
function appendLog(line) {
  const box = $("logBox");
  box.textContent += line + "\n";
  const lines = box.textContent.split("\n");
  if (lines.length > 600) box.textContent = lines.slice(-500).join("\n");
  box.scrollTop = box.scrollHeight;
}

// ─── clipboard + toast ───────────────────────────────────────
async function copy(text) {
  try {
    await navigator.clipboard.writeText(text);
    toast(STR.toast.copied);
  } catch {
    const ta = document.createElement("textarea");
    ta.value = text;
    document.body.appendChild(ta);
    ta.select();
    document.execCommand("copy");
    ta.remove();
    toast(STR.toast.copied);
  }
}

let toastTimer = null;
function toast(msg, isError = false) {
  const el = $("toast");
  el.textContent = msg;
  el.classList.toggle("is-error", isError);
  el.classList.add("is-visible");
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => el.classList.remove("is-visible"), 2400);
}
