// Package server provides the Research Loop HTTP API and htmx dashboard.
// Start with: research-loop dashboard [--port 4321]
package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/research-loop/research-loop/internal/config"
	"github.com/research-loop/research-loop/internal/persistence"
)

// Server is the Research Loop dashboard + API server.
type Server struct {
	workspace string
	port      int
	cfg       *config.Config
	mux       *http.ServeMux
}

// New creates a new HTTP server for the given workspace.
func New(workspaceRoot string, port int) (*Server, error) {
	cfg, err := config.Load(workspaceRoot)
	if err != nil {
		return nil, err
	}

	s := &Server{
		workspace: workspaceRoot,
		port:      port,
		cfg:       cfg,
		mux:       http.NewServeMux(),
	}

	s.registerRoutes()
	return s, nil
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("\033[32m✓\033[0m  Dashboard live at \033[1mhttp://localhost:%d\033[0m\n", s.port)
	fmt.Printf("   Claude Code MCP: claude mcp add research-loop -- %s mcp serve\n", executablePath())
	fmt.Println()
	return http.ListenAndServe(addr, s.mux)
}

// registerRoutes wires all HTTP routes.
func (s *Server) registerRoutes() {
	// Dashboard
	s.mux.HandleFunc("GET /", s.handleDashboard)

	// htmx partials
	s.mux.HandleFunc("GET /partials/sessions", s.handlePartialSessions)
	s.mux.HandleFunc("GET /partials/session/{id}", s.handlePartialSessionDetail)
	s.mux.HandleFunc("GET /partials/kg/{id}", s.handlePartialKG)
	s.mux.HandleFunc("GET /partials/notebook/{id}", s.handlePartialNotebook)

	// JSON API (for Agent SDK runner)
	s.mux.HandleFunc("GET /api/sessions", s.apiSessions)
	s.mux.HandleFunc("GET /api/session/{id}", s.apiSession)
	s.mux.HandleFunc("POST /api/session/{id}/kg", s.apiUpdateKG)
	s.mux.HandleFunc("GET /api/health", s.apiHealth)
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	sessions := s.loadSessions()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, renderDashboard(sessions, s.workspace))
}

// ─── htmx partials ───────────────────────────────────────────────────────────

func (s *Server) handlePartialSessions(w http.ResponseWriter, r *http.Request) {
	sessions := s.loadSessions()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, renderSessionList(sessions))
}

func (s *Server) handlePartialSessionDetail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	sess := s.loadSession(id)
	if sess == nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, renderSessionDetail(sess))
}

func (s *Server) handlePartialKG(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	content := s.readSessionFile(id, "knowledge_graph.md")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, renderMarkdownPane("Knowledge Graph", content))
}

func (s *Server) handlePartialNotebook(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	content := s.readSessionFile(id, "lab_notebook.md")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, renderMarkdownPane("Lab Notebook", content))
}

// ─── JSON API ────────────────────────────────────────────────────────────────

func (s *Server) apiHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{
		"status":    "ok",
		"workspace": s.workspace,
		"version":   "0.1.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) apiSessions(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.loadSessions())
}

func (s *Server) apiSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	sess := s.loadSession(id)
	if sess == nil {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, sess)
}

func (s *Server) apiUpdateKG(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var body struct {
		NodeTitle  string `json:"node_title"`
		Result     string `json:"result"`
		Annotation string `json:"annotation"`
		Status     string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", 400)
		return
	}

	sess, err := persistence.LoadSession(s.workspace, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	kgPath := filepath.Join(sess.Root, "knowledge_graph.md")
	existing, _ := os.ReadFile(kgPath)

	statusEmoji := map[string]string{
		"improvement": "✅", "regression": "❌",
		"crash_failed": "💥", "checks_failed": "⚠️", "explored": "🔍",
	}
	emoji := statusEmoji[body.Status]
	if emoji == "" {
		emoji = "•"
	}

	entry := fmt.Sprintf("\n### %s %s\n\n- **Result**: %s\n- **Status**: %s\n- **Causal annotation**: %s\n",
		emoji, body.NodeTitle, body.Result, body.Status, body.Annotation)

	updated := string(existing) + entry
	_ = os.WriteFile(kgPath, []byte(updated), 0644)
	_ = sess.AppendJSONL(map[string]interface{}{
		"event": "kg_update_via_api", "node": body.NodeTitle,
		"result": body.Result, "status": body.Status,
	})

	writeJSON(w, map[string]string{"status": "ok", "session_id": id})
}

// ─── Data helpers ────────────────────────────────────────────────────────────

type SessionSummary struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	RunCount  int    `json:"run_count"`
	BestMetric string `json:"best_metric"`
	CreatedAt string `json:"created_at"`
}

type SessionDetail struct {
	SessionSummary
	Hypothesis    string `json:"hypothesis"`
	KnowledgeGraph string `json:"knowledge_graph"`
	LabNotebook   string `json:"lab_notebook"`
}

func (s *Server) loadSessions() []SessionSummary {
	dir := filepath.Join(s.workspace, ".research-loop", "sessions")
	entries, _ := os.ReadDir(dir)

	var result []SessionSummary
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id := e.Name()
		info, _ := e.Info()
		created := ""
		if info != nil {
			created = info.ModTime().Format("2006-01-02")
		}
		title := s.readFirstHeading(filepath.Join(dir, id, "hypothesis.md"))
		runs := s.countLines(filepath.Join(dir, id, "autoresearch.jsonl"))
		result = append(result, SessionSummary{
			ID: id, Title: title, RunCount: runs, CreatedAt: created,
		})
	}
	return result
}

func (s *Server) loadSession(id string) *SessionDetail {
	dir := filepath.Join(s.workspace, ".research-loop", "sessions", id)
	if _, err := os.Stat(dir); err != nil {
		return nil
	}
	runs := s.countLines(filepath.Join(dir, "autoresearch.jsonl"))
	return &SessionDetail{
		SessionSummary: SessionSummary{
			ID: id, Title: s.readFirstHeading(filepath.Join(dir, "hypothesis.md")),
			RunCount: runs,
		},
		Hypothesis:     readFile(filepath.Join(dir, "hypothesis.md")),
		KnowledgeGraph: readFile(filepath.Join(dir, "knowledge_graph.md")),
		LabNotebook:    readFile(filepath.Join(dir, "lab_notebook.md")),
	}
}

func (s *Server) readSessionFile(id, name string) string {
	return readFile(filepath.Join(s.workspace, ".research-loop", "sessions", id, name))
}

func (s *Server) readFirstHeading(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "(no hypothesis)"
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return "(untitled)"
}

func (s *Server) countLines(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	return strings.Count(string(data), "\n")
}

// ─── Util ─────────────────────────────────────────────────────────────────────

func readFile(path string) string {
	data, _ := os.ReadFile(path)
	return string(data)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func executablePath() string {
	path, err := os.Executable()
	if err != nil {
		return "research-loop"
	}
	return path
}
