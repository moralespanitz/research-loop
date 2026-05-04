# Security Workflow — $research-loop security

Autonomous security auditing that iteratively discovers, validates, and
reports vulnerabilities. Combines STRIDE threat modeling, OWASP Top 10
sweeps, and red-team adversarial analysis.

**Output:** Severity-ranked security report with threat model, findings,
mitigations, and iteration log.

## Trigger

- User invokes `$research-loop security`
- User says "security audit", "threat model", "find vulnerabilities"
- User says "red-team this app", "OWASP audit", "STRIDE analysis"

## Loop Support

```
# Unlimited — keep finding vulnerabilities until interrupted
$research-loop security

# Bounded — run exactly N iterations
$research-loop security
Iterations: 10

# With target scope
$research-loop security
Scope: src/api/**/*.ts
Focus: authentication and authorization flows
```

## PREREQUISITE: Interactive Setup

**CRITICAL — BLOCKING PREREQUISITE:** If invoked without `--diff`, scope,
or focus, scan the codebase first, then gather input.

**Single batched call — all 3 questions at once:**

| # | Header | Question | Options |
|---|--------|----------|---------|
| 1 | `Scope` | "What should I audit?" | "Entire codebase", "API routes + middleware", "Auth + authorization", "External-facing code" |
| 2 | `Depth` | "How thorough?" | "Quick scan (5)", "Standard (15)", "Deep (30+)", "Unlimited" |
| 3 | `Action` | "What should I do with vulnerabilities?" | "Report only", "Report + auto-fix Critical/High", "Report + CI gate" |

## Architecture

```
┌────────────────────────────────────────────────────────────┐
│                  SETUP PHASE (once)                        │
│  1. Scan codebase → tech stack, frameworks, APIs           │
│  2. Map assets → data stores, auth, external services      │
│  3. Map trust boundaries → client/server, API/DB           │
│  4. Generate STRIDE threat model                           │
│  5. Build attack surface map                               │
│  6. Create security-audit-results.tsv log                  │
├────────────────────────────────────────────────────────────┤
│                  AUTONOMOUS LOOP                           │
│  Each iteration: pick ONE attack vector, find/validate     │
│  the vulnerability, log, repeat.                           │
└────────────────────────────────────────────────────────────┘
```

## Setup: Threat Model Generation

### Step 1: Codebase Reconnaissance
Read: package.json, config files, Dockerfile, API routes, auth/middleware,
DB schemas, CI/CD configs.

### Step 2: Asset Identification

| Asset Type | Examples | Priority |
|------------|----------|----------|
| Data stores | Database, Redis, file storage | Critical |
| Authentication | Login, OAuth, JWT, sessions, API keys | Critical |
| API endpoints | REST routes, GraphQL resolvers, webhooks | High |
| External services | Payment APIs, email providers | High |
| User input surfaces | Forms, URL params, headers, file uploads | High |
| Configuration | Environment variables, CORS settings | Medium |

### Step 3: Trust Boundary Mapping
```
Trust Boundaries:
  ├── Browser ←→ Server
  ├── Server ←→ Database
  ├── Server ←→ External APIs
  ├── Public routes ←→ Authenticated routes
  ├── User role ←→ Admin role
  └── CI/CD ←→ Production
```

### Step 4: STRIDE Threat Model

| Threat | Question | Example |
|--------|----------|---------|
| Spoofing | Can attacker impersonate user/service? | Weak auth, missing CSRF |
| Tampering | Can data be modified? | SQL injection, prototype pollution |
| Repudiation | Can actions be denied? | Missing audit logs |
| Information Disclosure | Can sensitive data leak? | Error messages expose internals |
| Denial of Service | Can service be disrupted? | Missing rate limiting |
| Elevation of Privilege | Can user gain unauthorized access? | IDOR, broken access control |

### Step 5: Attack Surface Map
```
Attack Surface:
  ├── Entry Points
  │   ├── GET /api/users/:id        → IDOR risk
  │   ├── POST /api/auth/login      → Brute force
  │   └── POST /api/upload          → Path traversal
  ├── Data Flows
  │   ├── User input → DB query     → Injection risk
  │   └── JWT → route handler       → Token validation
  └── Abuse Paths
      ├── Rate limit bypass → account takeover
      └── IDOR chain → data exfiltration
```

## The Security Loop

### Phase 1: Select Attack Vector

1. Critical STRIDE threats not yet tested
2. OWASP Top 10 categories not yet covered
3. High-severity attack paths from surface map
4. Dependency vulnerabilities (supply chain)
5. Configuration weaknesses (headers, CORS, CSP)
6. Business logic flaws (race conditions)
7. Information disclosure (error handling, debug modes)

### Phase 2: Analyze (Deep Dive)

1. Read all relevant code files
2. Trace data flow from entry point to data store
3. Identify missing validation, sanitization, or access checks
4. Look for known vulnerability patterns

### Phase 3: Validate (Proof Construction)

**Finding Proof Structure:**
```
  ├── Vulnerable code location (file:line)
  ├── Attack scenario (step-by-step)
  ├── Input that triggers the vulnerability
  ├── Expected vs actual behavior
  ├── Impact assessment
  └── Confidence level (Confirmed / Likely / Possible)
```

**Credential hygiene:** Always mask secrets in finding output — use
`<REDACTED_TOKEN>`, reference env var names, never live values.

### Phase 4: Classify

**Severity (CVSS-inspired):**
| Severity | Criteria |
|----------|----------|
| Critical | RCE, auth bypass, SQL injection, admin takeover |
| High | XSS (stored), SSRF, privilege escalation |
| Medium | CSRF, open redirect, info disclosure |
| Low | Missing headers, verbose errors |
| Info | Best practice suggestions |

**OWASP Top 10 (2021):** A01-A10 mapping.
**STRIDE mapping:** Tag each finding with STRIDE category.

### Phase 5: Log

```tsv
iteration	vector	severity	owasp	stride	confidence	location	description
0	-	-	-	-	-	-	baseline — 3 npm audit warnings
1	IDOR	High	A01	EoP	Confirmed	src/api/users.ts:42	No ownership check on GET /api/users/:id
```

### Phase 6: Repeat

Unbounded: Keep finding vulnerabilities. Never stop.
Bounded: After N iterations, generate final report.

**Coverage summary every 5 iterations:**
```
=== Security Audit Progress (iteration 10) ===
STRIDE Coverage: S[✓] T[✓] R[✗] I[✓] D[✓] E[✓] — 5/6
OWASP Coverage: A01[✓] A02[✗] A03[✓] ... — 5/10
Findings: 4 Critical, 2 High, 3 Medium, 1 Low
```

## Final Report Structure

- Executive Summary (date, scope, iterations, findings count)
- Threat Model (assets, trust boundaries, STRIDE analysis, attack surface map)
- Findings (descending severity, each with OWASP, STRIDE, location, evidence, mitigation)
- Coverage Matrix (OWASP + STRIDE tested/found)
- Dependency Audit
- Security Headers Check
- Recommendations (priority order)
- Iteration Log (full TSV)

## Red-Team Adversarial Lenses

- **Security Adversary:** "Hacker trying to breach this system"
- **Supply Chain Attacker:** "Compromising dependencies or build pipeline"
- **Insider Threat:** "Malicious employee or compromised account"
- **Infrastructure Attacker:** "Attacking deployment, not code"

## Flags

| Flag | Purpose |
|------|---------|
| `--diff` | Delta mode — only audit changed files |
| `--fix` | After audit, auto-fix Critical/High findings |
| `--fail-on {severity}` | Non-zero exit for CI/CD gating |

## Composite Metric

```
security_score = (owasp_tested/10)*50 + (stride_tested/6)*30 + min(findings, 20)
```

Higher = more thorough coverage. Incentivizes breadth across both taxonomies.
