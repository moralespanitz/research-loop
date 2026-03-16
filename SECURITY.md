# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| 0.2.x   | Yes       |
| 0.1.x   | No        |

## Data & Privacy

Research Loop is designed with a local-first architecture:

- **All session data stays on your machine.** Nothing is sent to any server unless you explicitly export a bundle or publish to the registry.
- **LLM API calls** transmit paper text and code snippets to your configured provider (Anthropic, OpenAI, Ollama, etc.). If you require full on-device privacy, use Ollama or LM Studio as your backend.
- **API keys** are read from environment variables only. Never stored in plain-text config files or committed to git.
- **Paper library** stores PDFs and extracted text locally. No institutional credentials or paywalled content is accessed automatically.
- **`.research` bundles** contain paper metadata but not full PDFs. Review bundles before public sharing if working on proprietary code or unpublished results.

## Threat Model

Research Loop is a local developer tool, not a multi-user server. The primary risks are:

1. **Accidental secret exposure** — API keys or proprietary code in exported bundles or shared JSONL logs.
2. **Malicious `.research` bundles** — A bundle from an untrusted source could contain a `benchmark_command` that executes arbitrary shell commands. Only open bundles from sources you trust.
3. **LLM prompt injection** — A malicious paper could attempt to inject instructions into the LLM pipeline. The system does not have write access to your broader filesystem by default, but exercise caution with untrusted PDFs.

## Reporting a Vulnerability

If you discover a security vulnerability, please **do not open a public GitHub issue**.

Instead, email: **security@research-loop.dev**

Include:
- A description of the vulnerability
- Steps to reproduce
- Potential impact
- Any suggested mitigations

You will receive acknowledgment within 48 hours and a resolution timeline within 7 days. We follow responsible disclosure — we ask that you give us 90 days to address the issue before public disclosure.

## Security Best Practices

- Pin your `config.toml` to a specific model version to avoid unexpected behavior from model updates.
- Review auto-generated `hypothesis.md` before running experiments — the Empirical Agent executes the benchmark command as written.
- Do not run Research Loop with elevated privileges (sudo/root).
- Use `.gitignore` to exclude `.research-loop/credentials.toml` and any local PDF files from version control.
