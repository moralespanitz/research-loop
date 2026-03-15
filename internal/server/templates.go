package server

import (
	"fmt"
	"html"
	"strings"
)

// ─── Full dashboard page ─────────────────────────────────────────────────────

func renderDashboard(sessions []SessionSummary, workspace string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Research Loop</title>
  <script src="https://unpkg.com/htmx.org@2.0.0"></script>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #0f0f0f; color: #e0e0e0; min-height: 100vh; }
    header { background: #1a1a1a; border-bottom: 1px solid #2a2a2a; padding: 16px 24px; display: flex; align-items: center; gap: 12px; }
    header h1 { font-size: 18px; font-weight: 600; color: #fff; }
    header .badge { background: #2a2a2a; border-radius: 4px; padding: 2px 8px; font-size: 11px; color: #888; }
    .workspace-path { font-size: 12px; color: #555; margin-left: auto; font-family: monospace; }
    .layout { display: grid; grid-template-columns: 280px 1fr; height: calc(100vh - 57px); }
    .sidebar { background: #141414; border-right: 1px solid #2a2a2a; overflow-y: auto; }
    .sidebar-header { padding: 16px; font-size: 11px; font-weight: 600; text-transform: uppercase; color: #555; letter-spacing: 0.08em; }
    .session-item { padding: 12px 16px; border-bottom: 1px solid #1e1e1e; cursor: pointer; transition: background 0.1s; }
    .session-item:hover { background: #1e1e1e; }
    .session-item.active { background: #1e2a3a; border-left: 2px solid #3b82f6; }
    .session-title { font-size: 13px; font-weight: 500; color: #ccc; margin-bottom: 4px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
    .session-meta { font-size: 11px; color: #555; }
    .main { overflow-y: auto; padding: 24px; }
    .empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100%%; gap: 12px; color: #444; }
    .empty-state code { background: #1a1a1a; padding: 8px 16px; border-radius: 6px; font-size: 13px; color: #888; }
    .card { background: #1a1a1a; border: 1px solid #2a2a2a; border-radius: 8px; margin-bottom: 16px; overflow: hidden; }
    .card-header { padding: 12px 16px; background: #1e1e1e; border-bottom: 1px solid #2a2a2a; font-size: 12px; font-weight: 600; text-transform: uppercase; color: #666; letter-spacing: 0.06em; }
    .card-body { padding: 16px; }
    pre { background: #111; border-radius: 4px; padding: 12px; font-size: 12px; color: #aaa; overflow-x: auto; white-space: pre-wrap; word-break: break-word; line-height: 1.6; }
    .stat-row { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; margin-bottom: 16px; }
    .stat { background: #1a1a1a; border: 1px solid #2a2a2a; border-radius: 8px; padding: 16px; text-align: center; }
    .stat-value { font-size: 28px; font-weight: 700; color: #fff; }
    .stat-label { font-size: 11px; color: #555; margin-top: 4px; text-transform: uppercase; letter-spacing: 0.06em; }
    .tabs { display: flex; gap: 4px; margin-bottom: 16px; }
    .tab { padding: 6px 14px; border-radius: 6px; font-size: 13px; cursor: pointer; color: #666; border: 1px solid transparent; }
    .tab.active, .tab:hover { background: #1e1e1e; border-color: #2a2a2a; color: #ccc; }
    .connect-bar { background: #0d1f0d; border: 1px solid #1a3a1a; border-radius: 8px; padding: 12px 16px; margin-bottom: 16px; font-size: 12px; color: #4ade80; }
    .connect-bar code { background: #111; padding: 2px 6px; border-radius: 3px; }
    .htmx-indicator { opacity: 0; transition: opacity 0.2s; }
    .htmx-request .htmx-indicator { opacity: 1; }
  </style>
</head>
<body>
<header>
  <span>🔬</span>
  <h1>Research Loop</h1>
  <span class="badge">v0.1.0</span>
  <span class="workspace-path">%s</span>
</header>
<div class="layout">
  <div class="sidebar">
    <div class="sidebar-header">Sessions</div>
    <div id="session-list" hx-get="/partials/sessions" hx-trigger="load, every 30s">
      %s
    </div>
  </div>
  <div class="main" id="main-panel">
    %s
  </div>
</div>
</body>
</html>`,
		html.EscapeString(workspace),
		renderSessionList(sessions),
		renderMainDefault(),
	)
}

func renderMainDefault() string {
	return `<div class="empty-state">
  <div style="font-size:48px">🔬</div>
  <div style="font-size:16px;color:#666">Select a session or start a new investigation</div>
  <code>research-loop start "https://arxiv.org/abs/2403.05821"</code>
</div>`
}

// ─── Session list (sidebar) ───────────────────────────────────────────────────

func renderSessionList(sessions []SessionSummary) string {
	if len(sessions) == 0 {
		return `<div style="padding:16px;font-size:12px;color:#444">
      No sessions yet.<br><br>
      <code style="font-size:11px;color:#666">research-loop start &lt;arxiv-url&gt;</code>
    </div>`
	}

	var sb strings.Builder
	for _, s := range sessions {
		sb.WriteString(fmt.Sprintf(`<div class="session-item"
      hx-get="/partials/session/%s"
      hx-target="#main-panel"
      hx-swap="innerHTML">
      <div class="session-title">%s</div>
      <div class="session-meta">%d runs · %s</div>
    </div>`,
			html.EscapeString(s.ID),
			html.EscapeString(truncate(s.Title, 36)),
			s.RunCount,
			html.EscapeString(s.CreatedAt),
		))
	}
	return sb.String()
}

// ─── Session detail (main panel) ─────────────────────────────────────────────

func renderSessionDetail(s *SessionDetail) string {
	connectCmd := fmt.Sprintf("claude mcp add research-loop -- $(which research-loop) mcp serve")

	hyp := strings.TrimSpace(s.Hypothesis)
	if hyp == "" {
		hyp = "(no hypothesis yet)"
	}

	return fmt.Sprintf(`
<div class="connect-bar">
  Connected to Claude Code via MCP &nbsp;·&nbsp;
  <code>%s</code>
</div>

<div class="stat-row">
  <div class="stat"><div class="stat-value">%d</div><div class="stat-label">Total Runs</div></div>
  <div class="stat"><div class="stat-value">—</div><div class="stat-label">Best Metric</div></div>
  <div class="stat"><div class="stat-value">—</div><div class="stat-label">Improvements</div></div>
</div>

<div class="card">
  <div class="card-header">Hypothesis</div>
  <div class="card-body">
    <pre>%s</pre>
  </div>
</div>

<div class="tabs">
  <div class="tab active"
    hx-get="/partials/kg/%s"
    hx-target="#tab-content"
    hx-swap="innerHTML">Knowledge Graph</div>
  <div class="tab"
    hx-get="/partials/notebook/%s"
    hx-target="#tab-content"
    hx-swap="innerHTML">Lab Notebook</div>
</div>

<div id="tab-content" hx-get="/partials/kg/%s" hx-trigger="load">
  <div class="htmx-indicator">Loading…</div>
</div>`,
		html.EscapeString(connectCmd),
		s.RunCount,
		html.EscapeString(hyp),
		html.EscapeString(s.ID),
		html.EscapeString(s.ID),
		html.EscapeString(s.ID),
	)
}

// ─── Markdown pane ────────────────────────────────────────────────────────────

func renderMarkdownPane(title, content string) string {
	if content == "" {
		content = "(empty)"
	}
	return fmt.Sprintf(`<div class="card">
  <div class="card-header">%s</div>
  <div class="card-body">
    <pre>%s</pre>
  </div>
</div>`,
		html.EscapeString(title),
		html.EscapeString(content),
	)
}

// ─── Util ─────────────────────────────────────────────────────────────────────

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
