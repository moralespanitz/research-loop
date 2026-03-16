/**
 * Research Loop — Agent SDK Runner
 *
 * Drives Claude Code programmatically using @anthropic-ai/claude-agent-sdk.
 * This is how Research Loop talks to Claude Code without a human at the terminal.
 *
 * Usage:
 *   npx ts-node runner.ts ingest "https://arxiv.org/abs/2403.05821"
 *   npx ts-node runner.ts experiment <session-id>
 *   npx ts-node runner.ts annotate <session-id> "node-title" "result" "causal annotation"
 */

import { query, type ClaudeAgentOptions } from "@anthropic-ai/claude-agent-sdk";
import * as fs from "fs";
import * as path from "path";

// ─── Workspace helpers ────────────────────────────────────────────────────────

function findWorkspaceRoot(start: string = process.cwd()): string {
  let dir = start;
  while (true) {
    if (fs.existsSync(path.join(dir, ".research-loop"))) return dir;
    const parent = path.dirname(dir);
    if (parent === dir) return start;
    dir = parent;
  }
}

function latestSession(workspace: string): string | null {
  const sessDir = path.join(workspace, ".research-loop", "sessions");
  if (!fs.existsSync(sessDir)) return null;
  const entries = fs.readdirSync(sessDir)
    .filter(e => fs.statSync(path.join(sessDir, e)).isDirectory())
    .sort()
    .reverse();
  return entries[0] ?? null;
}

function readSessionFile(workspace: string, sessionId: string, filename: string): string {
  const p = path.join(workspace, ".research-loop", "sessions", sessionId, filename);
  return fs.existsSync(p) ? fs.readFileSync(p, "utf8") : "";
}

// ─── Run a Claude Code agent task ────────────────────────────────────────────

async function runAgent(prompt: string, options: Partial<ClaudeAgentOptions> = {}) {
  const defaults: ClaudeAgentOptions = {
    allowedTools: ["Read", "Write", "Edit", "Bash", "Glob", "Grep"],
    permissionMode: "acceptEdits",
  };

  const merged = { ...defaults, ...options };

  console.error(`\n[research-loop agent] Starting task…\n`);
  let result = "";

  for await (const message of query({ prompt, options: merged })) {
    // Stream assistant text to stderr (not stdout — stdout is for results)
    if ("type" in message && message.type === "assistant") {
      const text = (message as any).content
        ?.filter((b: any) => b.type === "text")
        ?.map((b: any) => b.text)
        ?.join("") ?? "";
      if (text) process.stderr.write(text);
    }
    // Capture final result
    if ("result" in message && message.result) {
      result = message.result as string;
    }
  }

  return result;
}

// ─── Commands ────────────────────────────────────────────────────────────────

/** Ingest a paper and extract a hypothesis via the Claude Code agent. */
async function cmdIngest(url: string) {
  const workspace = findWorkspaceRoot();

  const prompt = `You are the Epistemic Agent in Research Loop.

Your task: ingest this paper and extract a structured research hypothesis.

Paper: ${url}
Workspace: ${workspace}

Steps:
1. Use the research_ingest_paper MCP tool to fetch and process the paper.
   If the MCP tool is not available, use WebFetch to download the abstract from
   https://export.arxiv.org/api/query?id_list=<id> and extract the hypothesis manually.
2. Write the structured hypothesis to ${workspace}/.research-loop/sessions/<session-id>/hypothesis.md
3. Report: session ID, core claim, proposed experiment, baseline repo, metric.

Follow the idea-selection skill from ${workspace}/skills/idea-selection/SKILL.md before finalizing the hypothesis.`;

  const result = await runAgent(prompt, {
    allowedTools: ["Read", "Write", "Edit", "Bash", "Glob", "Grep", "WebFetch"],
  });

  if (result) console.log(result);
}

/** Run one experiment loop iteration for a session. */
async function cmdExperiment(sessionId?: string) {
  const workspace = findWorkspaceRoot();
  const sid = sessionId ?? latestSession(workspace);
  if (!sid) {
    console.error("No sessions found. Run: research-loop start <url>");
    process.exit(1);
  }

  const hypothesis = readSessionFile(workspace, sid, "hypothesis.md");
  const kg = readSessionFile(workspace, sid, "knowledge_graph.md");

  const prompt = `You are the Empirical Agent in Research Loop.

Session: ${sid}
Workspace: ${workspace}

Current hypothesis:
${hypothesis}

Knowledge graph (what has been tried):
${kg}

Your task:
1. Read the knowledge graph. Do NOT repeat any experiment that has already been tried.
2. Propose ONE new code mutation that directly tests the core claim from the hypothesis.
3. If a baseline repo is specified, clone or find it. If not available, create a minimal
   Python training script that can be benchmarked.
4. Apply the mutation.
5. Run the benchmark command and capture the metric output (look for lines like METRIC name=value).
6. Log the result to ${workspace}/.research-loop/sessions/${sid}/autoresearch.jsonl
7. Update the knowledge graph with a causal annotation explaining WHY this worked or failed.

Follow the execution skill from ${workspace}/skills/execution/SKILL.md.
Output: the metric result and your causal annotation.`;

  const result = await runAgent(prompt);
  if (result) console.log(result);
}

/** Add a causal annotation to the knowledge graph. */
async function cmdAnnotate(sessionId: string, nodeTitle: string, result: string, annotation: string) {
  const workspace = findWorkspaceRoot();
  const sid = sessionId ?? latestSession(workspace);
  if (!sid) { console.error("No session found"); process.exit(1); }

  // POST to dashboard API if it's running
  try {
    const res = await fetch(`http://localhost:4321/api/session/${sid}/kg`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ node_title: nodeTitle, result, annotation, status: "explored" }),
      signal: AbortSignal.timeout(2000),
    });
    if (res.ok) {
      console.log(`Knowledge graph updated via dashboard API.`);
      return;
    }
  } catch {
    // Dashboard not running — write directly
  }

  const kgPath = path.join(workspace, ".research-loop", "sessions", sid, "knowledge_graph.md");
  const existing = fs.existsSync(kgPath) ? fs.readFileSync(kgPath, "utf8") : "";
  const entry = `\n### 🔍 ${nodeTitle}\n\n- **Result**: ${result}\n- **Causal annotation**: ${annotation}\n`;
  fs.writeFileSync(kgPath, existing + entry);
  console.log(`Knowledge graph updated: ${kgPath}`);
}

// ─── CLI dispatch ─────────────────────────────────────────────────────────────

const [, , cmd, ...args] = process.argv;

switch (cmd) {
  case "ingest":
    if (!args[0]) { console.error("Usage: runner.ts ingest <arxiv-url>"); process.exit(1); }
    cmdIngest(args[0]).catch(e => { console.error(e); process.exit(1); });
    break;

  case "experiment":
    cmdExperiment(args[0]).catch(e => { console.error(e); process.exit(1); });
    break;

  case "annotate":
    if (args.length < 3) {
      console.error("Usage: runner.ts annotate <session-id> <node-title> <result> <annotation>");
      process.exit(1);
    }
    cmdAnnotate(args[0], args[1], args[2], args[3] ?? "").catch(e => { console.error(e); process.exit(1); });
    break;

  default:
    console.log(`Research Loop Agent SDK Runner

Commands:
  runner.ts ingest <arxiv-url>                         Ingest a paper via Claude Code agent
  runner.ts experiment [session-id]                    Run one experiment loop iteration
  runner.ts annotate <session-id> <node> <result> <why> Add causal annotation to knowledge graph

Requires: ANTHROPIC_API_KEY environment variable`);
}
