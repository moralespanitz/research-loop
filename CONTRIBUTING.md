# Contributing to Research Loop

Research Loop is an open-source Agent OS for scientific researchers. We welcome contributions from researchers, engineers, and anyone who wants to make science faster.

## Ways to Contribute

- **Bug reports** — open an issue with steps to reproduce and your environment details
- **Feature requests** — open an issue describing the user problem, not just the solution
- **Skills** — add or improve a skill in `.claude/skills/` following the guide below
- **Fellows** — build a new Fellow (autonomous capability package) and submit it to the registry
- **Go core** — contribute to the binary: TUI panes, LLM drivers, ingestion pipeline, loop state machine
- **Documentation** — improve the README, add examples, write tutorials
- **Research validation** — test the tool on real papers and report where it breaks

## Development Setup

```bash
# Prerequisites: Go 1.22+, git
git clone https://github.com/moralespanitz/research-loop
cd research-loop

# Build the binary
go build -o research-loop ./cmd/...

# Run tests
go test ./...

# Install locally
go install ./cmd/research-loop
```

## Project Structure

```
research-loop/
├── cmd/                  # CLI entry points
├── internal/             # Go core: loop, discover, explore, ingestion
├── fellows/                # Built-in Fellow definitions (FELLOW.toml + prompts)
├── skills/               # Duplicate of .claude/skills/ (source of truth)
├── .claude/
│   ├── skills/           # Claude Code skills (loaded via Skill tool)
│   ├── commands/         # Slash commands
│   └── hooks/            # SessionStart hooks
├── CLAUDE.md             # Workspace system prompt
└── .research-loop/       # Runtime state (sessions, config, library)
```

## Writing a Skill

Skills are Markdown files that tell Claude Code how to behave. They are the primary UX layer of Research Loop.

**File location:** `.claude/skills/<skill-name>/SKILL.md`  
**Also mirror to:** `skills/<skill-name>/SKILL.md`

### Skill structure

```markdown
---
name: skill-name
description: Use when [specific triggering conditions — NOT a summary of what the skill does]
---

<SUBAGENT-STOP>
If you were dispatched as a subagent to execute a specific task, skip this skill.
Do the task and return structured results immediately.
</SUBAGENT-STOP>

<HARD-GATE>        <!-- optional: for non-negotiable rules -->
[Rule that must never be broken]
</HARD-GATE>

# Skill Name

[Content]
```

### Key rules for skills

1. **Description = triggering conditions only.** Never summarize the skill's workflow in the description — Claude will use it as a shortcut and skip reading the skill body.
2. **Always include `<SUBAGENT-STOP>`** so subagents dispatched for search/execution tasks don't enter the advisor flow.
3. **One thing at a time.** Skills that interact with the user must present one question, option, or result at a time.
4. **Use `<HARD-GATE>` for non-negotiable rules** (e.g., no searches before topic confirmation, one Carlini question at a time).
5. **Subagent prompts must be task-shaped.** When a skill spawns subagents, each prompt must start with "You are a research search agent. Do not ask questions. Search and return structured results immediately."

See `skills/writing-skills/` for a complete guide (inspired by [obra/superpowers](https://github.com/obra/superpowers)).

## Writing a Fellow

A Fellow is an autonomous capability package — a scheduled agent that works on the researcher's behalf.

```
fellows/
  my-hand/
    FELLOW.toml       # Manifest: name, description, schedule, tools
    system.md       # System prompt for this Fellow
    SKILL.md        # Domain expertise (optional)
```

**FELLOW.toml schema:**
```toml
name = "my-hand"
description = "One sentence: what does this Fellow do autonomously?"
schedule = "daily"          # cron expression or: hourly, daily, on-paper-add
tools = ["web_search", "read_file", "write_file"]
max_runtime_minutes = 30
requires_approval = false   # true = pauses for researcher sign-off
```

## Pull Request Process

1. **Fork** the repository and create a branch: `git checkout -b feat/my-feature`
2. **Test your change** — for skills, manually test with Claude Code; for Go code, run `go test ./...`
3. **Update docs** — if you add a new skill or Fellow, update the README table
4. **Open a PR** against `main` with:
   - A clear description of what changed and why
   - A reference to the issue it addresses (if any)
   - For skills: a description of the scenario you tested it against

## Code Style

- **Go**: follow standard `gofmt` formatting; use `golangci-lint`
- **Markdown**: use ATX headings (`#`), fenced code blocks, and plain prose — no HTML
- **TOML**: use lowercase keys, group related keys together
- **No emojis** in code or documentation unless the user explicitly requests them

## Commit Messages

Follow the conventional commits format:

```
feat: add Deep Learner Fellow with 5-phase corpus pipeline
fix: subagent interception when research-loop skill loads in worker context
docs: add CONTRIBUTING.md and ROADMAP.md
refactor: unify LLM driver interface across providers
```

## Community

- **Issues**: https://github.com/moralespanitz/research-loop/issues
- **Discussions**: https://github.com/moralespanitz/research-loop/discussions
- **Security**: security@research-loop.dev (see SECURITY.md)

## Code of Conduct

Be direct, be honest, be kind. We are all here to make science better.
Discrimination, harassment, and bad faith participation are not welcome.
