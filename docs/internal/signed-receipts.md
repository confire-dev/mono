# Signed Provenance Receipts — Design v2

**Status:** Proposal  
**Branch:** feat/signed-provenance-receipts  
**Algorithm:** Ed25519  
**Philosophy:** local-first, additive, zero latency impact on the hook path

---

## Table of Contents

1. [Existing Architecture Summary](#1-existing-architecture-summary)
2. [Design Goals](#2-design-goals)
3. [Receipt Data Model](#3-receipt-data-model)
4. [Signing Architecture](#4-signing-architecture)
5. [Verification](#5-verification)
6. [Storage & Chaining](#6-storage--chaining)
7. [Integration Points](#7-integration-points)
8. [Security & Threat Model](#8-security--threat-model)
9. [Implementation Roadmap](#9-implementation-roadmap)

---

## 1. Existing Architecture Summary

### Core Data Models

**`InterceptEvent`** (`intercept/types.go`) — canonical event envelope, hook → CLI → daemon:

```
Host, Strategy, Phase (tool.pre / tool.post / session.start / …)
Session { ID, CWD, TranscriptPath }
Tool    { Name, Input (any), Output (any), UseID, IsMCP, MCPServer, DurationMs }
```

**`ProvenanceLabel`** (`provenance/labels.go`) — metadata record emitted after PostToolUse:

```
ContextID         string     — orphaned field, never populated (repurposed below)
SessionID         string
Client            string
SourceTool        string
MCPServer         string
OriginDomain      string
TrustLevel        enum       — trusted_local | workspace | external_untrusted |
                               mcp_unknown | mcp_trusted | sensitive | agent_generated
Flags             []string   — secret_redacted | prompt_injection_detected |
                               hidden_unicode_stripped | mutation_result | sanitized | …
RedactionsCount   int
SanitizationCount int
RiskScore         int
CreatedAt         time.Time
```

**`InterceptResult`** — engine response to the hook adapter:

```
Kind: block | review | warn | sanitize | passthrough | replace-output | add-context
ToolOutput, ToolInput, Context, SystemMessage, Reason, Stats
```

**`policy.Rule`** — one firewall rule:

```
ID, Name, Enabled, Phase (PreToolUse / PostToolUse), Action, Severity,
Match (ToolName, CommandRegex, OutputSecretScan, …), Message, MinMode, Source, Group
```

**`policy.CachedPolicy`** — on-disk custom rule cache (`~/.confire/policies/cache.json`):

```
FetchedAt, Version, Rules []Rule, GroupOverrides map[string]bool
```
No signature today — see §4.3 for the bundle receipt.

### Existing Event Flows

**PreToolUse** (`daemon.go::handleConn`):
1. `ConsumeBypassNext()` → passthrough if flag set
2. `firewall.CheckFlowRules(recentLabels, event, mode)` — cross-tool chain detection (ring buffer of 10 labels)
3. `guardrail.Run(event)` — policy engine, per-rule evaluation
4. `mcpRisk.Run(event)` — MCP risk classifier
5. Result serialized to socket → hook adapter → agent

**PostToolUse** (`daemon.go::handleConn`):
1. `sanitizeOutput()` / `runMCPSanitizePipeline()` — secret redaction + injection sanitization
2. `provenance.Classify(event, report)` — builds `ProvenanceLabel` from sanitized event
3. `writeProvenanceLabel(label)` → `~/.confire/sessions/<session_id>.jsonl` (metadata only)
4. `postProvenanceEvent(label)` → telemetry to Worker (fire-and-forget)
5. Returns `ResultSanitize` with clean output if findings, else `ResultPassthrough`

**Existing signing** (`transport/signer.go`): HMAC-SHA256 per request to Worker
(`X-Confire-Sig`). Transport-level authentication only. No event-level provenance signing exists.

**Existing storage:**

| Path | Contents | Tamper-protected? |
|---|---|---|
| `~/.confire/sessions/<session_id>.jsonl` | ProvenanceLabels | No |
| `~/.confire/policies/cache.json` | Custom rule cache | No |
| `~/.confire/stats.db` (SQLite) | Event counts only | No |

---

## 2. Design Goals

| Goal | Mechanism |
|---|---|
| Self-contained verifiable event records | Signed receipt per event |
| Offline verification | Ed25519 public key embedded in every receipt |
| Tamper-evident history | `previous_payload_hash` chain (renamed from `previous_receipt_hash`) |
| Pre/post correlation | `context_id` links `tool.pre` ↔ `tool.post` for the same call |
| Local-first | Daemon holds signing key; no network required |
| Remote policy trust | Worker co-signature on custom-rule decisions (Phase 2) |
| Policy bundle integrity | Dedicated receipt when a custom bundle is loaded (Phase 1) |
| Third-party verifiability | Exported bundles + `confire receipts verify` |
| Daemon restart resilience | Chain tip persisted in `.meta` file |

---

## 3. Receipt Data Model

### 3.1 Updated JSON Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12",
  "$id": "https://confire.dev/schemas/receipt/v1",
  "title": "ConFireReceipt",
  "type": "object",
  "required": [
    "version", "receipt_id", "previous_payload_hash",
    "payload_hash", "phase", "event", "timestamp",
    "confire_version", "signers"
  ],
  "properties": {

    "version": {
      "type": "string",
      "const": "1"
    },

    "confire_version": {
      "type": "string",
      "description": "CLI binary version that produced this receipt, e.g. '1.4.2'. Present on every receipt type."
    },

    "receipt_id": {
      "type": "string",
      "description": "UUID v4. Unique identifier for this receipt."
    },

    "previous_payload_hash": {
      "type": "string",
      "description": "SHA-256 hex of the payload_hash field of the immediately preceding receipt in this session's chain. '000...000' (64 zeros) for the genesis receipt. Named 'previous_payload_hash' (not 'previous_receipt_hash') because it stores the payload_hash of the prior receipt, not a hash of the entire prior receipt object."
    },

    "payload_hash": {
      "type": "string",
      "description": "SHA-256 hex of the canonical body bytes (everything except signers[]). The chain tip stored in .meta after each write."
    },

    "phase": {
      "type": "string",
      "enum": [
        "tool.pre",
        "tool.post",
        "session.start",
        "session.end",
        "policy.bundle.loaded"
      ],
      "description": "'policy.bundle.loaded' is a synthetic phase emitted when the daemon verifies a custom policy bundle at startup."
    },

    "parent_session_id": {
      "type": "string",
      "description": "Optional. Reserved for multi-device or forked session support. Identifies a prior session this one was forked from. Omit when not applicable."
    },

    "event": {
      "type": "object",
      "required": ["session_id", "client"],
      "description": "Present on all phases. Fields not applicable to a phase are omitted.",
      "properties": {

        "session_id": {
          "type": "string"
        },

        "client": {
          "type": "string",
          "description": "claude_code | cursor | vscode | windsurf | codex"
        },

        "context_id": {
          "type": "string",
          "description": "Shared UUID v4 linking a tool.pre receipt with its corresponding tool.post receipt for the same tool invocation. Set from Tool.UseID when available; otherwise generated by the daemon at PreToolUse time and stored until the matching PostToolUse arrives. Absent on session.start, session.end, and policy.bundle.loaded receipts."
        },

        "tool_name": {
          "type": "string"
        },

        "is_mcp": {
          "type": "boolean"
        },

        "mcp_server": {
          "type": "string"
        },

        "duration_ms": {
          "type": "integer",
          "description": "Tool execution duration as reported by the host. Only meaningful on tool.post."
        },

        "input_hash": {
          "type": "string",
          "description": "SHA-256 hex of canonical JSON of tool.Input. Present on tool.pre and tool.post. Allows a verifier who holds the original input to prove the daemon processed it."
        },

        "output_hash": {
          "type": "string",
          "description": "SHA-256 hex of canonical JSON of the SANITIZED tool output — after secret redaction and prompt-injection removal have been applied. This is what the model actually saw. Present on tool.post only. A verifier who holds the sanitized output can prove this receipt covers it. Raw (pre-sanitization) output is never hashed into a receipt to avoid inadvertent secret commitment."
        },

        "cwd_hash": {
          "type": "string",
          "description": "SHA-256 hex of the session CWD path string. The path itself is not stored (privacy). Lets a verifier who knows the path confirm it matches."
        }
      }
    },

    "decision": {
      "type": "object",
      "description": "Present on tool.pre receipts only.",
      "properties": {
        "action": {
          "type": "string",
          "enum": ["block", "review", "warn", "allow", "passthrough"]
        },
        "rule_id":     { "type": "string" },
        "rule_name":   { "type": "string" },
        "rule_source": {
          "type": "string",
          "enum": ["builtin", "custom", "flow", "mcp_risk"],
          "description": "'custom' means the firing rule came from a cloud-fetched policy bundle."
        },
        "policy_mode": {
          "type": "string",
          "enum": ["observe", "balanced", "strict", "bypass"]
        },
        "bypass_next": {
          "type": "boolean",
          "description": "true if a bypass-next flag was consumed for this event instead of running the firewall."
        }
      }
    },

    "provenance": {
      "type": "object",
      "description": "Present on tool.post receipts only. Mirrors ProvenanceLabel fields.",
      "properties": {
        "trust_level":        { "type": "string" },
        "flags":              { "type": "array", "items": { "type": "string" } },
        "redactions_count":   { "type": "integer" },
        "sanitization_count": { "type": "integer" },
        "risk_score":         { "type": "integer" },
        "origin_domain":      { "type": "string" }
      }
    },

    "bundle": {
      "type": "object",
      "description": "Present on policy.bundle.loaded receipts only.",
      "required": ["bundle_version", "fetched_at", "rule_count", "verification_result"],
      "properties": {
        "bundle_version": {
          "type": "string",
          "description": "Version string from the CachedPolicy.Version field."
        },
        "fetched_at": {
          "type": "string",
          "format": "date-time",
          "description": "When the bundle was fetched from the Worker (CachedPolicy.FetchedAt)."
        },
        "rule_count": {
          "type": "integer",
          "description": "Number of custom rules in the bundle."
        },
        "group_override_count": {
          "type": "integer",
          "description": "Number of group overrides applied."
        },
        "verification_result": {
          "type": "string",
          "enum": ["verified", "no_signature", "invalid_signature", "fallback_to_builtin"],
          "description": "'verified': bundle sig checked against embedded cloud key and passed. 'no_signature': bundle predates signing (pre-Phase2), loaded without verification. 'invalid_signature': sig check failed, daemon fell back to built-in rules only. 'fallback_to_builtin': bundle absent or corrupt."
        },
        "signer_id": {
          "type": "string",
          "description": "The signer_id from the bundle's BundleSig, if present. e.g. 'remote:confire-cloud'."
        }
      }
    },

    "timestamp": {
      "type": "string",
      "format": "date-time",
      "description": "RFC 3339 with nanoseconds. Wall clock at daemon event-processing time."
    },

    "signers": {
      "type": "array",
      "minItems": 1,
      "description": "Canonical order: local signer at index 0, remote co-signer at index 1 (if present). Verifiers must enforce this order: fail if a remote signer appears before the local signer.",
      "items": {
        "type": "object",
        "required": ["signer_id", "pubkey", "algorithm", "signature"],
        "properties": {
          "signer_id": {
            "type": "string",
            "description": "'local:<device_id>' for the device key. 'remote:confire-cloud' for the Worker co-signature."
          },
          "pubkey": {
            "type": "string",
            "description": "base64url-encoded 32-byte Ed25519 public key. Embedded in each receipt so receipts are self-contained."
          },
          "algorithm": {
            "type": "string",
            "const": "Ed25519"
          },
          "signature": {
            "type": "string",
            "description": "base64url-encoded 64-byte Ed25519 signature over canonical_body_bytes. Each signer signs the same canonical body bytes (the receipt body with signers[] omitted entirely). A remote co-signer signs the same bytes as the local signer."
          }
        }
      }
    }
  }
}
```

### 3.2 `context_id` — Linking Pre and Post Receipts

`context_id` was an orphaned field on `ProvenanceLabel`. It is now a first-class correlation key.

**Population rules:**

1. The host provides `Tool.UseID` in the `InterceptEvent` envelope. When non-empty, use it directly as `context_id`. Claude Code populates this; other hosts may not.

2. When `Tool.UseID` is empty, the daemon generates a UUID v4 at `PhaseToolPre` time and stores it in `daemonState.pendingContextIDs[session_id+tool_use_id_fallback]`. When the matching `PhaseToolPost` arrives (correlated by `session_id` + `tool.Name` + call order), the stored UUID is attached and the entry is removed.

3. Session-level receipts (`session.start`, `session.end`) and `policy.bundle.loaded` receipts omit `context_id`.

**Why this matters:** A verifier can now reconstruct a complete tool call timeline — the pre-decision receipt and the post-sanitization receipt are joined by a shared `context_id`, giving a full view of "what was about to run, what was decided, and what the model ultimately saw."

### 3.3 `output_hash` — Sanitized Output Only

`event.output_hash` commits to the **sanitized** output (post-redaction, post-injection-removal) — the bytes the model context actually received.

**Rationale:**
- The purpose of a receipt is to prove what the agent saw, not what was originally returned by the tool. The pre-sanitization output may contain live secrets or injection payloads — hashing it into a permanent record would create an indirect secret commitment that could be exploited if receipts are exported or synced.
- A verifier who wants to confirm that a specific sanitized output was processed can SHA-256 the cleaned text and compare it against `output_hash`. A verifier who holds both the raw and sanitized outputs can independently verify that the redaction pipeline ran correctly.
- The `provenance.redactions_count` and `provenance.flags` fields already document what the sanitization pipeline changed.

The pre-sanitization output is **never hashed** into any receipt field.

---

## 4. Signing Architecture

### 4.1 Local Signing Flow

**Key layout:**

```
~/.confire/keys/               # 0700
  local.ed25519.seed           # 32 raw bytes — private key seed; 0600; never transmitted
  local.ed25519.pub            # base64url line + "# confire-local <device_id>" header; 0644
  revoked.json                 # list of revoked key fingerprints (future rotation)
```

**Key generation (daemon startup):**

```go
// cli/receipt/keystore.go

func LoadOrGenerateKey(deviceID string) (ed25519.PrivateKey, error) {
    path := seedPath()
    if seed, err := os.ReadFile(path); err == nil && len(seed) == 32 {
        return ed25519.NewKeyFromSeed(seed), nil
    }
    _, priv, err := ed25519.GenerateKey(rand.Reader)
    if err != nil {
        return nil, err
    }
    seed := priv.Seed()
    if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
        return nil, err
    }
    if err := os.WriteFile(path, seed, 0600); err != nil {
        return nil, err
    }
    pub := priv.Public().(ed25519.PublicKey)
    _ = writePubKey(pub, deviceID) // best-effort
    return priv, nil
}
```

**Signing each receipt:**

The signed payload is the canonical body bytes — deterministic JSON serialization (RFC 8785 / JCS) of the receipt with the `signers` field omitted entirely.

```
canonical_body_bytes = jcs.Marshal(receipt_body_without_signers)
payload_hash         = hex(SHA256(canonical_body_bytes))
signature            = Ed25519.Sign(private_key, canonical_body_bytes)
```

Ed25519 in Go's `crypto/ed25519` is deterministic — no per-signature randomness, no `k` reuse risk.

Each signer signs the same `canonical_body_bytes`. A remote co-signer (Phase 2) receives `canonical_body_bytes` from the daemon and signs it independently.

**Signer array order (canonical):**

```
signers[0]  — local signer  (always present)
signers[1]  — remote co-signer (present only when Phase 2 co-signing is enabled)
```

Verifiers enforce this order. A receipt with a remote signer at index 0 is rejected.

**In `handleConn`, signing is synchronous** (10μs per Ed25519 sign on modern hardware) and happens immediately after the decision or provenance label is produced. The file write is goroutined:

```go
// PreToolUse
receipt := state.buildPreToolReceipt(event, result, matchResult)
go state.writeAndChain(receipt)

// PostToolUse
receipt := state.buildPostToolReceipt(event, sanitizedOutput, label)
go state.writeAndChain(receipt)
```

### 4.2 Daemon Restart Resilience

The chain tip (last `payload_hash` per session) is persisted in a `.meta` file alongside the JSONL receipt log. This ensures a daemon restart in the middle of a long session does not break the chain.

**`.meta` file format** (`~/.confire/receipts/<session_id>.meta`):

```json
{
  "session_id": "abc123",
  "chain_tip": "a1b2c3d4...ff",
  "receipt_count": 23,
  "genesis_hash": "0000000000000000000000000000000000000000000000000000000000000000",
  "local_pubkey": "base64url…",
  "confire_version": "1.4.2",
  "started_at": "2026-06-16T10:00:01.234Z",
  "updated_at": "2026-06-16T10:22:11.456Z"
}
```

**Write protocol:**

1. After `writeReceipt()` succeeds (the JSONL append is fsync'd), atomically update the `.meta` file.
2. Use a write-to-tmp-then-rename pattern to ensure the `.meta` file is never partially written.

```go
func (s *receiptStore) writeAndChain(receipt Receipt) {
    data, _ := json.Marshal(receipt)
    // 1. Append to JSONL (O_APPEND is atomic for small writes on local filesystems)
    f, _ := os.OpenFile(s.jsonlPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
    f.Write(append(data, '\n'))
    f.Sync()
    f.Close()
    // 2. Atomically update .meta
    s.mu.Lock()
    s.chainTip = receipt.PayloadHash
    s.receiptCount++
    s.mu.Unlock()
    s.flushMeta() // write-tmp-rename
}
```

**On daemon startup**, for each in-progress session detected (by checking `~/.confire/receipts/`), the daemon reads `.meta` to restore `sessionChains[session_id] = chain_tip`. This means the next receipt emitted after a restart correctly sets `previous_payload_hash` to the tip of the pre-restart chain.

**`daemonState` additions:**

```go
type daemonState struct {
    // ... existing fields ...

    localSignKey      ed25519.PrivateKey           // loaded at daemon startup
    localPubkeyB64    string                       // base64url, cached for receipt construction
    receiptStore      *receipt.Store               // manages JSONL + .meta per session
    pendingContextIDs map[string]string            // key: sessionID+"::"+toolName+"::"+callSeq → context_id UUID
    contextIDMu       sync.Mutex
}
```

### 4.3 Policy Bundle Receipt

A `policy.bundle.loaded` receipt is emitted **at daemon startup**, immediately after the bundle is loaded and its signature verified (Phase 2) or skipped (Phase 1 / pre-signing bundles).

This receipt is the **genesis receipt** for the daemon's policy state. It precedes any session receipts and anchors the rule set in use.

**When it is generated:** inside `runDaemon()`, after `policy.LoadRules()` returns and before the Unix socket listener is opened. There is no active session at this point, so:

- `event.session_id` is set to a synthetic daemon-scope ID: `"daemon::" + daemonStartupID` (a UUID generated once per daemon process).
- `previous_payload_hash` is the 64-zero genesis value (no prior receipt in this chain).
- The bundle receipt is written to `~/.confire/receipts/daemon-policy.jsonl` (separate from session receipt files).

**Example `bundle` block:**

```json
{
  "bundle_version": "2026-06-14-v3",
  "fetched_at": "2026-06-14T09:00:00Z",
  "rule_count": 42,
  "group_override_count": 3,
  "verification_result": "verified",
  "signer_id": "remote:confire-cloud"
}
```

**Verification results and daemon behavior:**

| `verification_result` | Bundle loaded? | Logged? |
|---|---|---|
| `verified` | Yes | stderr info |
| `no_signature` | Yes (with warning) | stderr warn |
| `invalid_signature` | No — fallback to built-ins | stderr error |
| `fallback_to_builtin` | No — built-ins only | stderr info |

A receipt is written in all cases, including `invalid_signature` — the audit trail records the failure, not just successes.

### 4.4 Remote Co-signing (Phase 2)

When a **custom rule** fires (`rule.Source == "custom"`), the daemon can optionally request a co-signature from the Worker:

```
POST /v1/receipts/cosign
Authorization: Bearer <api_key>
X-Confire-Sig: <HMAC-SHA256 transport sig>

{
  "payload_hash": "<hex>",
  "canonical_body_b64": "<base64url of canonical_body_bytes>",
  "receipt_id": "<uuid>"
}

→ 200 OK
{
  "signer_id": "remote:confire-cloud",
  "pubkey":    "<base64url>",
  "signature": "<base64url>"
}
```

The Worker verifies the transport signature, confirms the user has an active plan with custom rules, signs `base64url.decode(canonical_body_b64)` with the Confire Cloud Ed25519 key, and returns the signer entry.

The daemon appends this as `signers[1]`. If the Worker is offline or returns an error, the receipt stands with the local signer only — co-signing is **additive, never required**.

**Confire Cloud public key distribution:**

- Embedded at compile time in the CLI binary (hardcoded constant, like a CT log key).
- Published at `https://keys.confire.dev/v1/signing-key` with its own self-signed receipt.
- Rotated annually with a minimum 90-day key-overlap window.

### 4.5 Policy Bundle Signing (Phase 2)

`policy.CachedPolicy` gains a `BundleSig` field:

```go
type CachedPolicy struct {
    FetchedAt      time.Time       `json:"fetched_at"`
    Version        string          `json:"version,omitempty"`
    Rules          []Rule          `json:"rules"`
    GroupOverrides map[string]bool `json:"group_overrides,omitempty"`
    BundleSig      *BundleSig      `json:"bundle_sig,omitempty"` // nil for pre-Phase2 bundles
}

type BundleSig struct {
    SignerID   string `json:"signer_id"`
    Pubkey     string `json:"pubkey"`
    Algorithm  string `json:"algorithm"`
    Signature  string `json:"signature"` // Ed25519.Sign(cloud_key, SHA256(canonical_json(rules+overrides+version+fetched_at)))
}
```

On load, if `BundleSig != nil`, the daemon verifies the signature against the embedded cloud public key. Failure sets `verification_result = "invalid_signature"` and falls back to built-in rules.

---

## 5. Verification

### 5.1 Algorithm

```
1. Read receipts from JSONL file in order.
2. Set expected_previous = "0000...0000" (64 zeros).
3. For each receipt:
   a. Deserialize receipt.
   b. Recompute canonical_body_bytes = jcs.Marshal(receipt_without_signers).
   c. Assert hex(SHA256(canonical_body_bytes)) == receipt.payload_hash.
   d. Assert receipt.previous_payload_hash == expected_previous.
   e. For each signer in receipt.signers (enforce index order):
      - pubkey  = base64url.decode(signer.pubkey)
      - sig     = base64url.decode(signer.signature)
      - assert ed25519.Verify(pubkey, canonical_body_bytes, sig) == true
      - if signer.signer_id starts with "remote:":
          - assert pubkey == embedded_cloud_pubkey  (or matches keys.confire.dev)
   f. If receipt has signers[0] with signer_id starting "remote:": REJECT (order violation).
   g. Set expected_previous = receipt.payload_hash.
4. Report results.
```

### 5.2 CLI Commands

```bash
# Verify all receipts in a session
confire receipts verify --session <session_id>

# Verify an exported bundle file
confire receipts verify ./audit-2026-06-16.bundle.json

# Inspect one receipt by ID
confire receipts show --receipt-id <uuid>

# Export a session bundle
confire receipts export --session <session_id> [--out <file>]

# List recent sessions with receipt counts
confire receipts list [--limit N]

# Key management
confire receipts key show          # print local public key + fingerprint
confire receipts key fingerprint   # SHA256[:16] of local pubkey
confire receipts key rotate        # generate new keypair; emit key-rotation receipt
```

### 5.3 Updated Verify Output

```
Session: abc123...
Chain:   47 receipts, unbroken
         genesis → ... → receipt_47  ✓

Signers:
  [0] local:device_abc123  Ed25519  VERIFIED  (matches ~/.confire/keys/local.ed25519.pub)
  [1] remote:confire-cloud Ed25519  VERIFIED  (matches embedded cloud key)  ← 12 receipts

Receipt breakdown:
  session.start           1
  policy.bundle.loaded    1  (bundle: 2026-06-14-v3, result: verified ✓)
  tool.pre               22
  tool.post              22
  session.end             1

Cloud co-signatures:   12 / 44 tool receipts  (custom rules fired 12 times)
Policy bundle:         verified ✓  (42 custom rules, fetched 2026-06-14)

Timestamps:
  Earliest: 2026-06-16T10:00:01.234Z
  Latest:   2026-06-16T11:22:33.456Z
  Duration: 1h22m32s

Warnings:
  Receipt 23: timestamp gap of 4m12s (possible clock skew or daemon restart)
  Receipt 31: context_id present on tool.post but no matching tool.pre found in session
              (may indicate a receipt from before context_id was populated — not a chain error)
```

If cloud co-signatures are absent entirely:

```
Cloud co-signatures:  0 / 44 tool receipts  (no custom rules fired, or Phase 2 not enabled)
Policy bundle:        no_signature  (bundle predates signing — loaded without verification)
```

If verification fails:

```
FAIL  Receipt 24: payload_hash mismatch
      expected: d4e5f6...
      got:      aabbcc...
      Chain is broken after receipt 23. Receipts 24–47 cannot be trusted.
```

---

## 6. Storage & Chaining

### 6.1 File Layout

```
~/.confire/
  keys/
    local.ed25519.seed         # 0600 — 32-byte private key seed; never transmitted
    local.ed25519.pub          # 0644 — base64url pubkey + signer_id header
    revoked.json               # future: list of revoked fingerprints

  sessions/
    <session_id>.jsonl         # existing: ProvenanceLabels (unchanged)

  receipts/
    daemon-policy.jsonl        # policy.bundle.loaded receipts (one per daemon start)
    <session_id>.jsonl         # signed receipts for this session (one per line)
    <session_id>.meta          # chain tip + session metadata (atomic write-rename)
```

### 6.2 `.meta` File

```json
{
  "session_id": "abc123",
  "chain_tip": "a1b2c3d4...ff",
  "receipt_count": 23,
  "genesis_hash": "0000000000000000000000000000000000000000000000000000000000000000",
  "local_pubkey": "base64url…",
  "confire_version": "1.4.2",
  "started_at": "2026-06-16T10:00:01.234Z",
  "updated_at": "2026-06-16T10:22:11.456Z"
}
```

Written atomically (write to `<session_id>.meta.tmp`, then `os.Rename`) after each JSONL append. On daemon restart, all `.meta` files in `~/.confire/receipts/` are read to restore `daemonState.receiptStore.chainTips`.

### 6.3 Chain Structure

```
daemon-policy.jsonl:
  Receipt 0  phase=policy.bundle.loaded
    previous_payload_hash = "0000...0000"
    payload_hash          = "bbcc11..."

<session_id>.jsonl:
  Receipt 0  phase=session.start
    previous_payload_hash = "0000...0000"
    payload_hash          = "a1b2c3..."

  Receipt 1  phase=tool.pre     context_id="uuid-A"
    previous_payload_hash = "a1b2c3..."
    payload_hash          = "d4e5f6..."

  Receipt 2  phase=tool.post    context_id="uuid-A"   ← same context_id as Receipt 1
    previous_payload_hash = "d4e5f6..."
    payload_hash          = "g7h8i9..."

  ...

  Receipt N  phase=session.end
    previous_payload_hash = <payload_hash of Receipt N-1>
    payload_hash          = <terminal hash>
```

The policy bundle chain (`daemon-policy.jsonl`) is independent from session chains. A session receipt does not chain to the bundle receipt — they are parallel chains. The session's `policy.bundle.loaded` receipt is visible in `confire receipts verify` output to show which bundle was active, but the chains are not linked by `previous_payload_hash`.

### 6.4 Session Bundle Export Format

```json
{
  "bundle_version": "1",
  "exported_at": "2026-06-16T11:22:33Z",
  "session_id": "abc123",
  "confire_version": "1.4.2",
  "genesis_hash": "0000...0000",
  "terminal_hash": "<payload_hash of last receipt>",
  "receipt_count": 47,
  "policy_bundle_receipt": { /* the policy.bundle.loaded receipt active during this session */ },
  "receipts": [ /* array of receipt objects in chain order */ ]
}
```

Including the `policy_bundle_receipt` allows a verifier to confirm the rule set in use without access to the daemon's separate `daemon-policy.jsonl`.

---

## 7. Integration Points

### 7.1 `handleConn` Changes (Summary)

**PreToolUse** — after guardrail + MCP risk evaluation:
```go
ctxID := state.resolveContextID(event)          // UseID or generated UUID
receipt := state.buildPreToolReceipt(event, result, matchResult, ctxID)
go state.receiptStore.WriteAndChain(receipt)
go state.maybeCosign(receipt)                   // Phase 2: async, non-blocking
json.NewEncoder(conn).Encode(result)            // unchanged
```

**PostToolUse** — after `sanitizeOutput()` + `provenance.Classify()`:
```go
ctxID := state.consumeContextID(event)          // retrieves and removes pending UUID
receipt := state.buildPostToolReceipt(event, sanitizedOutput, label, ctxID)
go state.receiptStore.WriteAndChain(receipt)
// existing: go writeProvenanceLabel(label) — unchanged
json.NewEncoder(conn).Encode(result)            // unchanged
```

**SessionStart:**
```go
state.onSessionStart(event)
go state.postSessionStart(event)
receipt := state.buildSessionStartReceipt(event)
go state.receiptStore.WriteAndChain(receipt)
```

**SessionEnd:**
```go
state.onSessionEnd(event)
receipt := state.buildSessionEndReceipt(event)
state.receiptStore.WriteAndChain(receipt)       // sync on session end to flush cleanly
```

### 7.2 `ProvenanceLabel.ContextID` — Now Populated

`provenance.Classify()` gains a `contextID string` parameter and sets `label.ContextID = contextID`. The JSONL label and the receipt now share this correlation key.

### 7.3 `daemonState` — Summary of Additions

```go
type daemonState struct {
    // existing fields unchanged …

    // Receipt signing
    localSignKey      ed25519.PrivateKey
    localPubkeyB64    string
    receiptStore      *receipt.Store      // manages JSONL + .meta per session

    // context_id correlation
    contextIDMu       sync.Mutex
    pendingContextIDs map[string]string   // key: sessionID+"::"+seq → context_id UUID
    contextIDSeq      map[string]int64    // per-session call counter for host-UseID-absent calls
}
```

### 7.4 Remote Rule Traceability

When `rule.Source == "custom"` and the bundle was `verified`, the `decision.rule_id` in the receipt pinpoints exactly which custom rule fired. Combined with the `policy_bundle_receipt` (which commits to the bundle version and verification result), a verifier can trace: "rule `custom/no-force-push` from bundle v2 fired, bundle was cloud-signed and verified."

---

## 8. Security & Threat Model

### 8.1 Key Management

| Concern | Mitigation |
|---|---|
| Key at rest | `local.ed25519.seed` — 0600, dir 0700 |
| Key in memory | Held in daemon process; acceptable for long-lived local process |
| Key backup | User advised to preserve `~/.confire/keys/local.ed25519.seed`; rotation preserves old receipt verifiability |
| Root compromise | Out of scope; local key is a user-scoped control, not a root-level boundary |

### 8.2 Key Compromise

1. Attacker can forge **new** receipts (future events) — cannot alter past chain retroactively.
2. User runs `confire receipts key rotate` → new keypair; key-rotation receipt signed by old key referencing new pubkey fingerprint.
3. Old receipts remain verifiable with the old pubkey (embedded per-receipt in `signers[0].pubkey`).
4. Dashboard-side: revocation tracker in `~/.confire/keys/revoked.json` (future).

### 8.3 Replay Protection

Each receipt has a unique `receipt_id` (UUID v4) and a `previous_payload_hash` that chains from a session-specific genesis. Replaying a receipt in a different session breaks `previous_payload_hash` continuity. Replaying the entire chain is a copy, not a forgery — it carries the original pubkey and can be distinguished from the original only if the verifier tracks session IDs.

### 8.4 `output_hash` and Secret Safety

The raw tool output is never hashed or stored in any receipt. Only the sanitized output hash is committed. This means:

- No live secrets can be recovered from a receipt file.
- A receipt can prove what the model saw without revealing what the tool originally returned.
- If secret redaction failed silently (a false negative), the `output_hash` reflects the un-redacted sanitized output — but it's a hash, not the content.

### 8.5 Canonical JSON / Signature Malleability

Use RFC 8785 (JCS) — JSON Canonicalization Scheme. All keys sorted lexicographically, no whitespace, UTF-8 strings. Go stdlib `encoding/json` marshals struct fields in declaration order (deterministic within a build, not across versions), so a purpose-built canonical serializer is required. A minimal JCS implementation is ~80 lines of Go with no external dependencies.

### 8.6 What Receipts Don't Protect Against

- A compromised daemon process (can forge receipts with the held private key)
- A compromised local key (see §8.2)
- Content confidentiality (hashes prove what was processed; they don't reveal it)
- Clock manipulation (timestamps are advisory, not security primitives)

Receipts are an **audit and tamper-evidence** mechanism. The security boundary is the daemon's firewall logic, not the receipt layer.

---

## 9. Implementation Roadmap

### Phase 1 (MVP — ~2–3 days): Local signing + bundle receipt

**New package: `cli/receipt/`**

| File | Purpose |
|---|---|
| `types.go` | `Receipt`, `ReceiptBody`, `ReceiptEvent`, `ReceiptDecision`, `ReceiptProvenance`, `ReceiptBundle`, `Signer` structs |
| `canon.go` | RFC 8785 JCS canonical JSON serializer (~80 lines, zero deps) |
| `keystore.go` | `LoadOrGenerateKey()`, `WritePubKey()`, key path helpers |
| `signer.go` | `Sign(privKey, body)`, `buildPreToolReceipt()`, `buildPostToolReceipt()`, `buildSessionReceipt()`, `buildBundleReceipt()` |
| `store.go` | `Store` type: `WriteAndChain()`, `.meta` atomic flush, `LoadChainTip()`, `RestoreFromMeta()` |
| `verify.go` | `VerifyChain()`, `VerifyReceipt()`, embedded cloud pubkey constant |

**Changes to existing files:**

| File | Change |
|---|---|
| `cli/cmd/daemon.go` | Add `localSignKey`, `localPubkeyB64`, `receiptStore`, `pendingContextIDs`, `contextIDSeq` to `daemonState`; load key at startup; emit receipts in `handleConn`; emit bundle receipt after `policy.LoadRules()` |
| `cli/cmd/daemon.go` | `runDaemon()`: call `receipt.Store.RestoreFromMeta()` at startup to reload chain tips for any in-progress sessions |
| `cli/provenance/labels.go` | `Classify()` gains `contextID string` parameter; sets `label.ContextID` |
| `cli/policy/cache.go` | Add `BundleSig *BundleSig` to `CachedPolicy`; verify on load; return `VerificationResult` string |

**New CLI file:**

| File | Purpose |
|---|---|
| `cli/cmd/receipts.go` | `confire receipts` subcommands: `verify`, `export`, `show`, `list`, `key show/fingerprint/rotate` |

**Phase 1 explicitly excludes:**
- Remote co-signing (Worker endpoint) — Phase 2
- Policy bundle `BundleSig` verification — Phase 2 (Phase 1 emits bundle receipt with `verification_result: "no_signature"`)
- Dashboard UI — Phase 3

**Phase 1 deliverables checklist:**

- [ ] RFC 8785 canonical JSON implementation with round-trip test
- [ ] Key generation + load at daemon startup
- [ ] Genesis receipt on `session.start`
- [ ] `tool.pre` receipt after every guardrail decision
- [ ] `tool.post` receipt after every provenance classify
- [ ] `session.end` receipt (sync write)
- [ ] `policy.bundle.loaded` receipt at daemon startup (with `verification_result: "no_signature"` for Phase 1)
- [ ] `.meta` atomic write after each receipt
- [ ] `RestoreFromMeta()` called at daemon startup
- [ ] `context_id` populated from `Tool.UseID` or generated UUID; `ProvenanceLabel.ContextID` set
- [ ] `output_hash` = SHA-256 of sanitized output only
- [ ] `confire_version` field on every receipt
- [ ] `confire receipts verify --session <id>` (local key only in Phase 1)
- [ ] `confire receipts export --session <id>`
- [ ] `confire receipts key show`

### Phase 2 (~1 week): Policy bundle signing + remote co-signing

- [ ] Worker endpoint `POST /v1/receipts/cosign`
- [ ] `BundleSig` added to `CachedPolicy`; Worker signs bundle on `policy pull`
- [ ] Daemon verifies bundle sig on load; sets `verification_result: "verified"` or `"invalid_signature"`
- [ ] Daemon requests co-signature for custom-rule decisions (goroutine, non-blocking, appended to receipt on response)
- [ ] `confire receipts key rotate` (emits key-rotation receipt)
- [ ] `confire receipts key fingerprint`
- [ ] Verify output: cloud co-signature count + policy bundle status

### Phase 3 (~2 weeks): Dashboard integration

- [ ] Session timeline: per-event receipt status badges (local / cloud-cosigned)
- [ ] "Verify session" upload → server-side chain verification
- [ ] Policy bundle history page with per-bundle verification status
- [ ] Receipt explorer: click any event → full receipt JSON
- [ ] `confire receipts list` with filtering by action / trust level

### Phase 4 (future)

- RFC 3161 TSA timestamp submission for compliance use cases
- WASM in-browser receipt verifier (client-side, no upload required)
- Multi-device key federation (multiple local signers in one session)
- Confire notary: public append-only log of session terminal hashes (transparency log)

---

## Appendix A: Rationale for Ed25519

| Alternative | Why not chosen |
|---|---|
| ECDSA P-256 | Requires per-signature randomness (k); reuse is catastrophic; slower |
| RSA-2048 | 256-byte signatures vs. 64 bytes; 100× larger keys; padding oracle surface |
| HMAC-SHA256 (existing) | Symmetric — verifier needs the secret key; no third-party verifiability |
| Ed25519ph (pre-hash variant) | Designed for large messages; our canonical bodies are always < 1KB; unnecessary |

Ed25519 is in Go's standard library (`crypto/ed25519`). No external cryptography dependencies introduced.

## Appendix B: `previous_payload_hash` Naming Rationale

The field was `previous_receipt_hash` in v1. The rename to `previous_payload_hash` is more precise: it stores the `payload_hash` field of the prior receipt — not a hash of the entire prior receipt object (which would include `signers[]` and produce a different value). The name makes the chaining invariant unambiguous: `receipt[n].previous_payload_hash == receipt[n-1].payload_hash`.
