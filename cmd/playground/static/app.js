import { SAMPLES, MONACO_LANGUAGE } from "./samples.js";

// ── Editor abstraction ───────────────────────────────────────────────────────

let baseView = null;
let headView = null;
let usingFallback = false;

async function initEditors(type) {
  const baseHost = document.getElementById("base-cm");
  const headHost = document.getElementById("head-cm");
  const baseTa   = document.getElementById("base-ta");
  const headTa   = document.getElementById("head-ta");

  try {
    const [
      { basicSetup, EditorView },
      { EditorState },
      { oneDark },
    ] = await Promise.all([
      import("https://esm.sh/codemirror@6.0.1"),
      import("https://esm.sh/@codemirror/state@6.4.1"),
      import("https://esm.sh/@codemirror/theme-one-dark@6.1.2"),
    ]);

    let langExt = [];
    try {
      if (type === "graphql") {
        const { graphql } = await import("https://esm.sh/cm6-graphql@0.0.12");
        langExt = [graphql()];
      } else if (type === "openapi") {
        const { yaml } = await import("https://esm.sh/@codemirror/lang-yaml@6.1.1");
        langExt = [yaml()];
      }
    } catch (_) { /* language extension unavailable, continue without */ }

    const makeState = (doc) =>
      EditorState.create({ doc, extensions: [basicSetup, oneDark, ...langExt] });

    if (baseView) baseView.destroy();
    if (headView) headView.destroy();

    baseView = new EditorView({ state: makeState(SAMPLES[type].base), parent: baseHost });
    headView = new EditorView({ state: makeState(SAMPLES[type].head), parent: headHost });

    usingFallback = false;
    baseHost.style.display = "";
    headHost.style.display = "";
    baseTa.style.display   = "none";
    headTa.style.display   = "none";
  } catch (_) {
    // CDN unavailable — fall back to plain textareas
    usingFallback = true;
    baseHost.style.display = "none";
    headHost.style.display = "none";
    baseTa.style.display   = "";
    headTa.style.display   = "";
    baseTa.value = SAMPLES[type].base;
    headTa.value = SAMPLES[type].head;
  }
}

function getContent(which) {
  if (usingFallback) {
    return document.getElementById(which === "base" ? "base-ta" : "head-ta").value;
  }
  const view = which === "base" ? baseView : headView;
  return view ? view.state.doc.toString() : "";
}

function setContent(which, content) {
  if (usingFallback) {
    document.getElementById(which === "base" ? "base-ta" : "head-ta").value = content;
    return;
  }
  const view = which === "base" ? baseView : headView;
  if (!view) return;
  view.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: content } });
}

// ── Tab switching ────────────────────────────────────────────────────────────

let currentType = "openapi";

document.querySelectorAll(".tab").forEach((btn) => {
  btn.addEventListener("click", async () => {
    currentType = btn.dataset.type;
    document.querySelectorAll(".tab").forEach((b) => b.classList.remove("active"));
    btn.classList.add("active");
    await initEditors(currentType);
    clearResults();
  });
});

// ── Compare ──────────────────────────────────────────────────────────────────

const compareBtn  = document.getElementById("compare-btn");
const errorBanner = document.getElementById("error-banner");
const resultsEl   = document.getElementById("results");

function clearResults() {
  resultsEl.hidden = true;
  resultsEl.innerHTML = "";
  errorBanner.hidden = true;
  errorBanner.textContent = "";
}

compareBtn.addEventListener("click", async () => {
  clearResults();
  compareBtn.disabled = true;
  compareBtn.innerHTML = '<span class="spinner"></span>Comparing…';

  try {
    const resp = await fetch("/api/compare", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        schema_type:  currentType,
        base_content: getContent("base"),
        head_content: getContent("head"),
      }),
    });

    const data = await resp.json();

    if (!resp.ok) {
      errorBanner.textContent = data.error || "Unknown error";
      errorBanner.hidden = false;
      return;
    }

    renderResults(data);
  } catch (err) {
    errorBanner.textContent = "Network error: " + err.message;
    errorBanner.hidden = false;
  } finally {
    compareBtn.disabled = false;
    compareBtn.textContent = "Compare";
  }
});

// ── Render results ───────────────────────────────────────────────────────────

function renderResults(data) {
  const { summary, changes } = data;

  const summaryHtml = `
    <div class="summary-bar">
      <h2>Results</h2>
      <span class="pill pill-total">${summary.total} total</span>
      ${summary.breaking    > 0 ? `<span class="pill pill-breaking">${summary.breaking} breaking</span>` : ""}
      ${summary.non_breaking > 0 ? `<span class="pill pill-non">${summary.non_breaking} non-breaking</span>` : ""}
      ${summary.info        > 0 ? `<span class="pill pill-info">${summary.info} info</span>` : ""}
    </div>`;

  if (!changes || changes.length === 0) {
    resultsEl.innerHTML = summaryHtml + `<div class="no-changes">No changes detected</div>`;
    resultsEl.hidden = false;
    return;
  }

  const rows = changes.map((c) => {
    const sevClass = c.severity === "breaking"     ? "sev-breaking"
                   : c.severity === "non-breaking" ? "sev-non-breaking"
                   : "sev-info";
    const rowClass = c.severity === "breaking"     ? "row-breaking"
                   : c.severity === "non-breaking" ? "row-non-breaking"
                   : "row-info";
    const method = c.method
      ? `<span class="td-method">${esc(c.method)}</span> `
      : "";
    const beforeAfter = (c.before || c.after)
      ? `<div class="td-before-after">${c.before ? `<s>${esc(c.before)}</s> ` : ""}${c.after ? `→ ${esc(c.after)}` : ""}</div>`
      : "";

    return `
      <tr class="${rowClass}">
        <td><span class="severity-badge ${sevClass}">${esc(c.severity).toUpperCase()}</span></td>
        <td class="td-type">${esc(String(c.type))}</td>
        <td class="td-path">${method}${esc(c.path)}</td>
        <td class="td-desc">${esc(c.description)}${beforeAfter}</td>
      </tr>`;
  }).join("");

  resultsEl.innerHTML = summaryHtml + `
    <div class="changes-table-wrap">
      <table>
        <thead>
          <tr>
            <th>Severity</th>
            <th>Type</th>
            <th>Path</th>
            <th>Description</th>
          </tr>
        </thead>
        <tbody>${rows}</tbody>
      </table>
    </div>`;

  resultsEl.hidden = false;
}

function esc(str) {
  return String(str)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

// ── Init ─────────────────────────────────────────────────────────────────────

initEditors(currentType);
