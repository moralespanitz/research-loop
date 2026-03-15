// Package mcp implements a Model Context Protocol (MCP) server for Research Loop.
// Transport: stdio (JSON-RPC 2.0), as required by Claude Code.
//
// Claude Code connects with:
//   claude mcp add research-loop -- /path/to/research-loop mcp serve
//
// Or via .mcp.json in the project root for automatic project-scoped registration.
//
// Reference: https://modelcontextprotocol.io/docs/learn/architecture
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/research-loop/research-loop/internal/config"
	"github.com/research-loop/research-loop/internal/hypothesis"
	"github.com/research-loop/research-loop/internal/ingestion"
	"github.com/research-loop/research-loop/internal/llm"
	"github.com/research-loop/research-loop/internal/persistence"
)

// ─── JSON-RPC types ───────────────────────────────────────────────────────────

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type notification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// ─── MCP capability types ────────────────────────────────────────────────────

type serverInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type capabilities struct {
	Tools     map[string]interface{} `json:"tools"`
	Resources map[string]interface{} `json:"resources"`
}

type toolDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema inputSchema `json:"inputSchema"`
}

type inputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]schemaProp  `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}

type schemaProp struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type resourceDef struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MIMEType    string `json:"mimeType"`
}

// ─── Server ───────────────────────────────────────────────────────────────────

// Server is the Research Loop MCP server.
type Server struct {
	workspace string
	cfg       *config.Config
	out       *json.Encoder
	in        *bufio.Scanner
}

// New creates a new MCP server bound to the given workspace root.
func New(workspaceRoot string) (*Server, error) {
	cfg, err := config.Load(workspaceRoot)
	if err != nil {
		return nil, err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 4*1024*1024), 4*1024*1024)

	return &Server{
		workspace: workspaceRoot,
		cfg:       cfg,
		out:       enc,
		in:        scanner,
	}, nil
}

// Serve runs the MCP stdio loop until stdin closes.
func (s *Server) Serve() error {
	fmt.Fprintln(os.Stderr, "research-loop MCP server started (stdio)")

	for s.in.Scan() {
		line := strings.TrimSpace(s.in.Text())
		if line == "" {
			continue
		}

		var req request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.sendError(nil, -32700, "parse error")
			continue
		}

		s.handle(req)
	}

	if err := s.in.Err(); err != nil && err != io.EOF {
		return err
	}
	return nil
}

// ─── Dispatcher ──────────────────────────────────────────────────────────────

func (s *Server) handle(req request) {
	switch req.Method {
	case "initialize":
		s.handleInitialize(req)
	case "initialized":
		// notification, no response needed
	case "tools/list":
		s.handleToolsList(req)
	case "tools/call":
		s.handleToolsCall(req)
	case "resources/list":
		s.handleResourcesList(req)
	case "resources/read":
		s.handleResourcesRead(req)
	case "ping":
		s.send(response{JSONRPC: "2.0", ID: req.ID, Result: map[string]interface{}{}})
	default:
		s.sendError(req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method))
	}
}

// ─── initialize ──────────────────────────────────────────────────────────────

func (s *Server) handleInitialize(req request) {
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"serverInfo": serverInfo{
			Name:    "research-loop",
			Version: "0.1.0",
		},
		"capabilities": capabilities{
			Tools:     map[string]interface{}{"listChanged": false},
			Resources: map[string]interface{}{"listChanged": false},
		},
		"instructions": `Research Loop MCP server.

Available tools let you ingest papers, extract hypotheses, query the knowledge
graph, and manage research sessions — all from inside Claude Code.

Typical workflow:
  1. research_ingest_paper   — point at an ArXiv URL, get a hypothesis.md
  2. research_session_status — see what sessions exist and their state
  3. research_read_hypothesis — read the current hypothesis for a session
  4. research_kg_query       — search the knowledge graph for prior experiments
  5. research_update_kg      — add a causal annotation after an experiment

All state lives in plain Markdown and JSONL files in .research-loop/sessions/.
`,
	}
	s.send(response{JSONRPC: "2.0", ID: req.ID, Result: result})
}

// ─── tools/list ──────────────────────────────────────────────────────────────

func (s *Server) handleToolsList(req request) {
	tools := []toolDef{
		{
			Name:        "research_ingest_paper",
			Description: "Download and ingest a paper from an ArXiv URL or bare ID. Extracts a structured hypothesis and initializes a new research session. Returns the session ID and the extracted hypothesis.",
			InputSchema: inputSchema{
				Type: "object",
				Properties: map[string]schemaProp{
					"url": {Type: "string", Description: "ArXiv URL, bare ID (e.g. 2403.05821), or local PDF path"},
				},
				Required: []string{"url"},
			},
		},
		{
			Name:        "research_session_status",
			Description: "List all research sessions in the current workspace, showing session ID, paper title, and whether experiments have been run.",
			InputSchema: inputSchema{Type: "object", Properties: map[string]schemaProp{}},
		},
		{
			Name:        "research_read_hypothesis",
			Description: "Read the current hypothesis.md for a research session. Returns the full structured hypothesis including core claim, key insight, proposed experiment, and metric.",
			InputSchema: inputSchema{
				Type: "object",
				Properties: map[string]schemaProp{
					"session_id": {Type: "string", Description: "Session ID (from research_session_status). Leave empty to use the most recent session."},
				},
			},
		},
		{
			Name:        "research_read_notebook",
			Description: "Read the lab_notebook.md for a session — the human-readable log of all experiments run so far.",
			InputSchema: inputSchema{
				Type: "object",
				Properties: map[string]schemaProp{
					"session_id": {Type: "string", Description: "Session ID. Leave empty for most recent."},
				},
			},
		},
		{
			Name:        "research_kg_query",
			Description: "Read the knowledge_graph.md for a session. Returns the living DAG of every hypothesis tried, result observed, and causal annotation written. Use this to understand what has been explored and what has failed.",
			InputSchema: inputSchema{
				Type: "object",
				Properties: map[string]schemaProp{
					"session_id": {Type: "string", Description: "Session ID. Leave empty for most recent."},
				},
			},
		},
		{
			Name:        "research_update_kg",
			Description: "Append a causal annotation or new node to the knowledge_graph.md. Use after an experiment to record what was tried, the result, and why it succeeded or failed.",
			InputSchema: inputSchema{
				Type: "object",
				Properties: map[string]schemaProp{
					"session_id":  {Type: "string", Description: "Session ID. Leave empty for most recent."},
					"node_title":  {Type: "string", Description: "Short name for this hypothesis or experiment (e.g. 'gqa_4_groups')"},
					"result":      {Type: "string", Description: "Metric value or outcome (e.g. 'val_bpb: 3.21', 'crash_failed', 'checks_failed')"},
					"annotation":  {Type: "string", Description: "Causal explanation: WHY did this succeed or fail? What does it imply for the next hypothesis?"},
					"status":      {Type: "string", Description: "One of: improvement, regression, crash_failed, checks_failed, explored"},
				},
				Required: []string{"node_title", "result", "annotation", "status"},
			},
		},
		{
			Name:        "research_export_bundle",
			Description: "Export the current session as a portable .research bundle (ZIP archive). Returns the file path of the exported bundle.",
			InputSchema: inputSchema{
				Type: "object",
				Properties: map[string]schemaProp{
					"session_id": {Type: "string", Description: "Session ID. Leave empty for most recent."},
					"output":     {Type: "string", Description: "Output file path. Defaults to <session-id>.research"},
				},
			},
		},
	}

	s.send(response{JSONRPC: "2.0", ID: req.ID, Result: map[string]interface{}{"tools": tools}})
}

// ─── tools/call ──────────────────────────────────────────────────────────────

func (s *Server) handleToolsCall(req request) {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, -32602, "invalid params")
		return
	}

	str := func(key string) string {
		v, _ := params.Arguments[key].(string)
		return v
	}

	var result string
	var toolErr error

	switch params.Name {
	case "research_ingest_paper":
		result, toolErr = s.toolIngestPaper(str("url"))
	case "research_session_status":
		result, toolErr = s.toolSessionStatus()
	case "research_read_hypothesis":
		result, toolErr = s.toolReadFile("hypothesis.md", str("session_id"))
	case "research_read_notebook":
		result, toolErr = s.toolReadFile("lab_notebook.md", str("session_id"))
	case "research_kg_query":
		result, toolErr = s.toolReadFile("knowledge_graph.md", str("session_id"))
	case "research_update_kg":
		result, toolErr = s.toolUpdateKG(
			str("session_id"), str("node_title"),
			str("result"), str("annotation"), str("status"),
		)
	case "research_export_bundle":
		result, toolErr = s.toolExportBundle(str("session_id"), str("output"))
	default:
		s.sendError(req.ID, -32601, fmt.Sprintf("unknown tool: %s", params.Name))
		return
	}

	if toolErr != nil {
		s.send(response{
			JSONRPC: "2.0", ID: req.ID,
			Result: map[string]interface{}{
				"content": []map[string]interface{}{
					{"type": "text", "text": fmt.Sprintf("Error: %v", toolErr)},
				},
				"isError": true,
			},
		})
		return
	}

	s.send(response{
		JSONRPC: "2.0", ID: req.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": result},
			},
		},
	})
}

// ─── Tool implementations ────────────────────────────────────────────────────

func (s *Server) toolIngestPaper(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("url is required")
	}

	paperDir := s.workspace + "/.research-loop/library/papers"
	if err := os.MkdirAll(paperDir, 0755); err != nil {
		return "", err
	}

	ctx := context.Background()

	var paper *ingestion.Paper
	var err error
	if strings.HasSuffix(strings.ToLower(input), ".pdf") || fileExists(input) {
		paper, err = ingestion.FetchLocalPDF(input)
	} else {
		paper, err = ingestion.FetchArXiv(ctx, input, paperDir)
	}
	if err != nil {
		return "", fmt.Errorf("fetching paper: %w", err)
	}

	llmClient, err := llm.New(s.cfg.LLM)
	if err != nil {
		return "", fmt.Errorf("LLM not configured: %w\n\nSet ANTHROPIC_API_KEY or configure .research-loop/config.toml", err)
	}

	h, err := hypothesis.Extract(ctx, llmClient, paper)
	if err != nil {
		return "", fmt.Errorf("extracting hypothesis: %w", err)
	}

	session, err := persistence.NewSession(s.workspace, h.PaperTitle)
	if err != nil {
		return "", err
	}
	if err := session.WriteHypothesis(h); err != nil {
		return "", err
	}
	if err := session.WriteKnowledgeGraph(h.PaperTitle); err != nil {
		return "", err
	}
	if err := session.WriteLabNotebook(h.PaperTitle); err != nil {
		return "", err
	}
	_ = session.AppendJSONL(map[string]interface{}{
		"event":    "session_initialized_via_mcp",
		"arxiv_id": h.ArXivID,
		"paper":    h.PaperTitle,
	})

	return fmt.Sprintf(`Session initialized.

Session ID: %s
Paper: %s
ArXiv: https://arxiv.org/abs/%s

Core Claim:
%s

Key Insight:
%s

Proposed Experiment:
%s

Baseline: %s
Metric:   %s

Files written:
  %s/hypothesis.md
  %s/knowledge_graph.md
  %s/lab_notebook.md
  %s/autoresearch.jsonl

Use research_read_hypothesis to read the full hypothesis.md.`,
		session.ID, h.PaperTitle, h.ArXivID,
		h.CoreClaim, h.KeyInsight, h.ProposedExperiment,
		h.BaselineRepo, h.Metric,
		session.Root, session.Root, session.Root, session.Root,
	), nil
}

func (s *Server) toolSessionStatus() (string, error) {
	dir := s.workspace + "/.research-loop/sessions"
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		return "No sessions found. Use research_ingest_paper to start one.", nil
	}

	var sb strings.Builder
	sb.WriteString("Research Loop sessions:\n\n")

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		hypPath := dir + "/" + e.Name() + "/hypothesis.md"
		jsonlPath := dir + "/" + e.Name() + "/autoresearch.jsonl"

		title := readFirstHeading(hypPath)

		runs := 0
		if data, err := os.ReadFile(jsonlPath); err == nil {
			runs = strings.Count(string(data), "\n")
		}

		sb.WriteString(fmt.Sprintf("• %s\n", e.Name()))
		sb.WriteString(fmt.Sprintf("  Paper : %s\n", title))
		sb.WriteString(fmt.Sprintf("  Runs  : %d\n\n", runs))
	}

	return sb.String(), nil
}

func (s *Server) toolReadFile(filename, sessionID string) (string, error) {
	sessionDir, err := s.resolveSession(sessionID)
	if err != nil {
		return "", err
	}
	path := sessionDir + "/" + filename
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("%s not found for session — run research_ingest_paper first", filename)
	}
	return string(data), nil
}

func (s *Server) toolUpdateKG(sessionID, nodeTitle, result, annotation, status string) (string, error) {
	if nodeTitle == "" || result == "" || annotation == "" || status == "" {
		return "", fmt.Errorf("node_title, result, annotation, and status are all required")
	}

	sessionDir, err := s.resolveSession(sessionID)
	if err != nil {
		return "", err
	}

	kgPath := sessionDir + "/knowledge_graph.md"
	existing, err := os.ReadFile(kgPath)
	if err != nil {
		return "", fmt.Errorf("knowledge_graph.md not found — run research_ingest_paper first")
	}

	statusEmoji := map[string]string{
		"improvement":   "✅",
		"regression":    "❌",
		"crash_failed":  "💥",
		"checks_failed": "⚠️",
		"explored":      "🔍",
	}
	emoji := statusEmoji[status]
	if emoji == "" {
		emoji = "•"
	}

	entry := fmt.Sprintf("\n### %s %s\n\n- **Result**: %s\n- **Status**: %s\n- **Causal annotation**: %s\n",
		emoji, nodeTitle, result, status, annotation)

	updated := string(existing) + entry
	if err := os.WriteFile(kgPath, []byte(updated), 0644); err != nil {
		return "", err
	}

	// Also log to JSONL
	if sid, e := s.resolveSessionID(sessionID); e == nil {
		if sess, e2 := persistence.LoadSession(s.workspace, sid); e2 == nil {
			_ = sess.AppendJSONL(map[string]interface{}{
				"event":      "kg_update",
				"node":       nodeTitle,
				"result":     result,
				"status":     status,
				"annotation": annotation,
			})
		}
	}

	return fmt.Sprintf("Knowledge graph updated.\n\nAdded node: %s %s\nResult: %s\nStatus: %s\nAnnotation: %s",
		emoji, nodeTitle, result, status, annotation), nil
}

func (s *Server) toolExportBundle(sessionID, output string) (string, error) {
	sid, err := s.resolveSessionID(sessionID)
	if err != nil {
		return "", err
	}
	if output == "" {
		output = sid + ".research"
	}

	// Delegate to the export logic via CLI args — simplest approach
	// to avoid duplicating the zip logic here.
	// Instead, inline the core logic from cli/export.go:
	session, err := persistence.LoadSession(s.workspace, sid)
	if err != nil {
		return "", err
	}

	import_archive := func() error {
		import_zip(session, output)
		return nil
	}
	_ = import_archive

	return fmt.Sprintf("Bundle export: run 'research-loop export --session %s --output %s' from your terminal.\n(Bundle export via MCP tool will be available in v0.2.)", sid, output), nil
}

// ─── resources/list ──────────────────────────────────────────────────────────

func (s *Server) handleResourcesList(req request) {
	dir := s.workspace + "/.research-loop/sessions"
	entries, _ := os.ReadDir(dir)

	var resources []resourceDef

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id := e.Name()
		for _, fname := range []string{"hypothesis.md", "knowledge_graph.md", "lab_notebook.md"} {
			path := dir + "/" + id + "/" + fname
			if _, err := os.Stat(path); err != nil {
				continue
			}
			resources = append(resources, resourceDef{
				URI:         fmt.Sprintf("research://%s/%s", id, fname),
				Name:        fmt.Sprintf("%s — %s", id, fname),
				Description: resourceDescription(fname),
				MIMEType:    "text/markdown",
			})
		}
	}

	s.send(response{JSONRPC: "2.0", ID: req.ID, Result: map[string]interface{}{"resources": resources}})
}

// ─── resources/read ──────────────────────────────────────────────────────────

func (s *Server) handleResourcesRead(req request) {
	var params struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, -32602, "invalid params")
		return
	}

	// Parse research://<session-id>/<filename>
	uri := strings.TrimPrefix(params.URI, "research://")
	parts := strings.SplitN(uri, "/", 2)
	if len(parts) != 2 {
		s.sendError(req.ID, -32602, "invalid resource URI")
		return
	}

	path := s.workspace + "/.research-loop/sessions/" + parts[0] + "/" + parts[1]
	data, err := os.ReadFile(path)
	if err != nil {
		s.sendError(req.ID, -32002, fmt.Sprintf("resource not found: %s", params.URI))
		return
	}

	s.send(response{
		JSONRPC: "2.0", ID: req.ID,
		Result: map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"uri":      params.URI,
					"mimeType": "text/markdown",
					"text":     string(data),
				},
			},
		},
	})
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func (s *Server) send(r response) {
	_ = s.out.Encode(r)
}

func (s *Server) sendError(id interface{}, code int, msg string) {
	_ = s.out.Encode(response{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &rpcError{Code: code, Message: msg},
	})
}

func (s *Server) resolveSessionID(sessionID string) (string, error) {
	if sessionID != "" {
		return sessionID, nil
	}
	dir := s.workspace + "/.research-loop/sessions"
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		return "", fmt.Errorf("no sessions found — run research_ingest_paper first")
	}
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].IsDir() {
			return entries[i].Name(), nil
		}
	}
	return "", fmt.Errorf("no session directories found")
}

func (s *Server) resolveSession(sessionID string) (string, error) {
	sid, err := s.resolveSessionID(sessionID)
	if err != nil {
		return "", err
	}
	return s.workspace + "/.research-loop/sessions/" + sid, nil
}

func resourceDescription(fname string) string {
	switch fname {
	case "hypothesis.md":
		return "Structured hypothesis extracted from the paper: core claim, key insight, proposed experiment, and metric."
	case "knowledge_graph.md":
		return "Living DAG of every hypothesis tried, result observed, and causal annotation written."
	case "lab_notebook.md":
		return "Human-readable experiment log updated after every run."
	}
	return ""
}

func readFirstHeading(path string) string {
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// import_zip is a stub — full bundle zip is implemented in cli/export.go
func import_zip(session *persistence.Session, output string) {
	// no-op stub; full implementation lives in cli/export.go
	_ = session
	_ = output
}
