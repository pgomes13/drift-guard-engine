import json

log  = json.load(open('/tmp/drift-log-existing.json'))
meta = json.load(open('/tmp/new-entry.json'))
html_body = open('/tmp/entry-body.html').read()

entry = {
    "pr":        meta["pr"],
    "url":       meta["url"],
    "timestamp": meta["timestamp"],
    "summary": {
        "total":        meta["total"],
        "breaking":     meta["breaking"],
        "non_breaking": meta["non_breaking"],
        "info":         meta["info"],
    },
    "html_body": html_body,
}

log = [e for e in log if e.get("pr") != entry["pr"]]
log.insert(0, entry)

with open('/tmp/drift-log-new.json', 'w') as f:
    json.dump(log, f, indent=2)

# ── Stats ─────────────────────────────────────────────────────────────────────
total_prs         = len(log)
total_breaking    = sum(e["summary"]["breaking"]     for e in log)
total_nonbreaking = sum(e["summary"]["non_breaking"] for e in log)
prs_with_breaking = sum(1 for e in log if e["summary"]["breaking"] > 0)

# ── Chart data (last 20, oldest → newest) ────────────────────────────────────
chart_entries     = list(reversed(log[:20]))
chart_labels      = json.dumps([f"PR #{e['pr']}"              for e in chart_entries])
chart_breaking    = json.dumps([e["summary"]["breaking"]      for e in chart_entries])
chart_nonbreaking = json.dumps([e["summary"]["non_breaking"]  for e in chart_entries])

# ── Entry HTML ────────────────────────────────────────────────────────────────
def entry_html(e):
    has_cls  = " has-breaking" if e["summary"]["breaking"] > 0 else ""
    b_badge  = (f'<span class="badge-breaking">{e["summary"]["breaking"]} breaking</span>'
                if e["summary"]["breaking"] > 0 else "")
    ok_badge = (f'<span class="badge-ok">{e["summary"]["non_breaking"]} non-breaking</span>'
                if e["summary"]["non_breaking"] > 0 else "")
    body = e.get("html_body") or '<p class="no-drift">No details available.</p>'
    return (
        f'<div class="entry{has_cls}" data-pr="{e["pr"]}">'
        f'<div class="entry-header">'
        f'<a href="{e["url"]}">PR #{e["pr"]}</a>'
        f'<span class="entry-badges">{b_badge}{ok_badge}</span>'
        f'<span class="entry-time" data-utc="{e["timestamp"]}"></span>'
        f'</div>'
        f'<div class="entry-body">{body}</div>'
        f'</div>'
    )

entries_html = "\n".join(entry_html(e) for e in log)

# ── CSS ───────────────────────────────────────────────────────────────────────
css = (
    "*{box-sizing:border-box}"
    "body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;"
    "max-width:1100px;margin:48px auto;padding:0 24px;color:#24292e;background:#f6f8fa}"
    "h1{font-size:1.6rem;border-bottom:1px solid #e1e4e8;padding-bottom:12px;margin-bottom:24px}"
    ".stats{display:grid;grid-template-columns:repeat(auto-fit,minmax(180px,1fr));gap:16px;margin-bottom:28px}"
    ".stat-card{background:#fff;border:1px solid #e1e4e8;border-radius:8px;padding:20px 24px}"
    ".stat-card.breaking{border-left:4px solid #cb2431}"
    ".stat-value{font-size:2rem;font-weight:700;line-height:1}"
    ".stat-label{font-size:.8rem;color:#6a737d;margin-top:6px;text-transform:uppercase;letter-spacing:.04em}"
    ".chart-wrap{background:#fff;border:1px solid #e1e4e8;border-radius:8px;padding:20px 24px;margin-bottom:28px}"
    ".chart-wrap h2{font-size:1rem;margin:0 0 16px;color:#586069}"
    ".chart-wrap canvas{max-height:220px}"
    ".filter-bar{margin-bottom:16px;display:flex;align-items:center;gap:16px}"
    ".filter-bar label{font-size:.875rem;color:#586069;display:flex;align-items:center;gap:6px;cursor:pointer}"
    ".entry{background:#fff;border:1px solid #e1e4e8;border-radius:8px;margin-bottom:16px;overflow:hidden}"
    ".entry.hidden{display:none}"
    "tr.hidden{display:none}"
    ".entry-header{display:flex;justify-content:space-between;align-items:center;"
    "padding:10px 16px;background:#f6f8fa;border-bottom:1px solid #e1e4e8;flex-wrap:wrap;gap:8px}"
    ".entry-header a{font-weight:600;color:#0366d6;text-decoration:none}"
    ".entry-badges{display:flex;gap:6px}"
    ".badge-breaking{background:#ffeef0;color:#cb2431;font-size:.75rem;font-weight:600;"
    "padding:2px 8px;border-radius:12px;border:1px solid #ffc0c5}"
    ".badge-ok{background:#e6ffed;color:#22863a;font-size:.75rem;font-weight:600;"
    "padding:2px 8px;border-radius:12px;border:1px solid #b4e8bc}"
    ".entry-time{color:#6a737d;font-size:.8rem}"
    ".entry-body{padding:16px}"
    "table{width:100%;border-collapse:collapse;font-size:.875rem;border:1px solid #d0d7de}"
    "th{text-align:left;padding:10px 14px;background:#f6f8fa;border:1px solid #d0d7de;font-weight:600}"
    "td{padding:10px 14px;border:1px solid #d0d7de;vertical-align:top}"
    ".breaking td:first-child{color:#cb2431;font-weight:600}"
    ".nonbreaking td:first-child{color:#22863a}"
    ".summary{margin:0 0 12px;color:#586069;font-size:.875rem}"
    ".no-drift{color:#6a737d;font-style:italic}"
)

html = f"""<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>API Drift Log</title>
<style>{css}</style>
</head>
<body>
<h1>API Drift Log</h1>

<div class="stats">
  <div class="stat-card"><div class="stat-value">{total_prs}</div><div class="stat-label">PRs Checked</div></div>
  <div class="stat-card breaking"><div class="stat-value">{total_breaking}</div><div class="stat-label">Breaking Changes</div></div>
  <div class="stat-card"><div class="stat-value">{total_nonbreaking}</div><div class="stat-label">Non-breaking Changes</div></div>
  <div class="stat-card"><div class="stat-value">{prs_with_breaking}</div><div class="stat-label">PRs with Breaking Changes</div></div>
</div>

<div class="chart-wrap">
  <h2>Breaking changes per PR (last 20)</h2>
  <canvas id="driftChart"></canvas>
</div>

<div class="filter-bar">
  <label><input type="checkbox" id="filterBreaking"> Show breaking changes only</label>
</div>

<!-- ENTRIES -->
{entries_html}
<!-- ENTRIES -->

<script src="https://cdn.jsdelivr.net/npm/chart.js@4/dist/chart.umd.min.js"></script>
<script>
document.addEventListener("DOMContentLoaded", function() {{
  document.querySelectorAll(".entry-time[data-utc]").forEach(function(el) {{
    var d = new Date(el.dataset.utc), h = d.getHours(), m = d.getMinutes(), ap = h >= 12 ? "pm" : "am";
    h = h % 12 || 12;
    el.textContent = d.toLocaleDateString(undefined, {{weekday:"long",day:"numeric",month:"long",year:"numeric"}})
      + " " + h + ":" + (m < 10 ? "0" + m : m) + ap;
  }});

  document.getElementById("filterBreaking").addEventListener("change", function() {{
    document.querySelectorAll("tr.nonbreaking").forEach(function(row) {{
      row.classList.toggle("hidden", this.checked);
    }}.bind(this));
  }});

  new Chart(document.getElementById("driftChart").getContext("2d"), {{
    type: "bar",
    data: {{
      labels: {chart_labels},
      datasets: [
        {{label:"Breaking",     data:{chart_breaking},    backgroundColor:"rgba(203,36,49,0.7)", borderColor:"#cb2431", borderWidth:1}},
        {{label:"Non-breaking", data:{chart_nonbreaking}, backgroundColor:"rgba(34,134,58,0.5)", borderColor:"#22863a", borderWidth:1}}
      ]
    }},
    options: {{
      responsive: true,
      maintainAspectRatio: true,
      plugins: {{legend: {{position: "top"}}}},
      scales:  {{y: {{beginAtZero: true, ticks: {{stepSize: 1}}}}}}
    }}
  }});
}});
</script>
</body>
</html>"""

with open('/tmp/new-index.html', 'w') as f:
    f.write(html)
