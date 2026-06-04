export type Risk = 'HIGH' | 'MEDIUM' | 'LOW' | 'NONE'

export interface ClassifyResult {
  risk: Risk
  category: string | null
  matchedPattern: string | null
}

const NONE: ClassifyResult = { risk: 'NONE', category: null, matchedPattern: null }

// ── Injection patterns (ordered: first match wins) ────────────────────────

interface Pattern { re: RegExp; risk: Risk; category: string; label: string }

const PATTERNS: Pattern[] = [
  // Prompt injection — explicit override phrases
  { re: /ignore\s+(all\s+|previous\s+|your\s+)?(instructions?|rules|guidelines|constraints|prompt)/i,
    risk: 'HIGH', category: 'prompt_injection', label: 'ignore-instructions' },
  // Extended: "ignore my previous instructions", "ignore the following instructions", etc.
  { re: /ignore\s+\w+\s+(previous|following|old|initial|original)\s+(instructions?|rules|guidelines)/i,
    risk: 'HIGH', category: 'prompt_injection', label: 'ignore-instructions-extended' },
  { re: /disregard\s+(all\s+|previous\s+|your\s+)?(instructions?|rules|guidelines)/i,
    risk: 'HIGH', category: 'prompt_injection', label: 'disregard-instructions' },
  { re: /forget\s+everything\s+(you\s+)?(were|have\s+been)\s+(told|trained|instructed)/i,
    risk: 'HIGH', category: 'prompt_injection', label: 'forget-instructions' },
  { re: /new\s+(set\s+of\s+)?instructions?\s*[:.]/i,
    risk: 'HIGH', category: 'prompt_injection', label: 'new-instructions' },
  { re: /your\s+new\s+(task|role|purpose|objective|instructions?)\s*[:.]/i,
    risk: 'HIGH', category: 'prompt_injection', label: 'new-task-injection' },
  { re: /new\s+system\s+prompt/i,
    risk: 'HIGH', category: 'prompt_injection', label: 'new-system-prompt' },
  { re: /override\s+(previous|your)\s+(instructions?|settings|rules)/i,
    risk: 'HIGH', category: 'prompt_injection', label: 'override-instructions' },
  { re: /you\s+are\s+now\s+operating\s+in\s+(admin|unrestricted|developer)\s+mode/i,
    risk: 'HIGH', category: 'prompt_injection', label: 'mode-switch-injection' },
  { re: /output\s+your\s+(full\s+)?(system\s+)?prompt/i,
    risk: 'HIGH', category: 'prompt_injection', label: 'reveal-system-prompt' },
  { re: /also\s*[:.]\s*ignore\s+(your|previous)\s*(previous\s+)?guidelines/i,
    risk: 'HIGH', category: 'prompt_injection', label: 'inline-guideline-bypass' },

  // Role manipulation
  { re: /\bDAN\b.{0,80}broken\s+free\s+of.{0,60}confines/i,
    risk: 'HIGH', category: 'role_manipulation', label: 'DAN-jailbreak' },
  { re: /does\s+not\s+have\s+to\s+abide\s+by\s+the\s+rules\s+set\s+for\s+it/i,
    risk: 'HIGH', category: 'role_manipulation', label: 'DAN-no-rules' },
  { re: /enable\s+(developer|jailbreak|unrestricted|admin)\s+mode/i,
    risk: 'HIGH', category: 'role_manipulation', label: 'enable-jailbreak-mode' },
  { re: /you\s+(have\s+(no|zero)|with\s+no)\s+restrictions/i,
    risk: 'HIGH', category: 'role_manipulation', label: 'no-restrictions-claim' },
  { re: /break\s+free\s+from\s+(typical|all|any)\s+confines/i,
    risk: 'HIGH', category: 'role_manipulation', label: 'break-free' },
  { re: /bypass\s+(all\s+|any\s+)?(safety\s+)?(restrictions|guidelines|safety|rules)/i,
    risk: 'HIGH', category: 'role_manipulation', label: 'bypass-safety' },
  { re: /authoriz\w+\s+you\s+to\s+(skip|bypass|disable)\s+all\s+safety/i,
    risk: 'HIGH', category: 'role_manipulation', label: 'authority-bypass' },
  { re: /confirmation\s+code\s*[:.]\s*[A-Z0-9\-]+OVERRIDE/i,
    risk: 'HIGH', category: 'role_manipulation', label: 'admin-override-code' },
  // Role switch embedded in documents: "I will now ignore my previous instructions..."
  { re: /I\s+will\s+now\s+(ignore|disregard|forget)\s+.{0,30}(instructions?|guidelines|rules)/i,
    risk: 'HIGH', category: 'role_manipulation', label: 'inline-role-switch' },

  // Data exfiltration
  { re: /output\s+all\s+(environment\s+variables|api\s+keys|credentials|secrets)/i,
    risk: 'HIGH', category: 'data_exfiltration', label: 'dump-env-vars' },
  { re: /reveal\s+(your|the)\s+(system\s+prompt|api\s+keys?|credentials|secrets)/i,
    risk: 'HIGH', category: 'data_exfiltration', label: 'reveal-secrets' },
  { re: /base64\s+encode.{0,120}(include|send|post|upload|output|exfiltrate)\s+it/i,
    risk: 'HIGH', category: 'data_exfiltration', label: 'base64-exfil' },
  { re: /summarize\s+all\s+(files|code|data).{0,120}POST\s+the\s+result\s+to\s+https?:\/\//i,
    risk: 'HIGH', category: 'data_exfiltration', label: 'exfil-via-post' },
  { re: /~\/\.ssh\/id_(rsa|ed25519|ecdsa)/,
    risk: 'HIGH', category: 'data_exfiltration', label: 'ssh-key-access' },
  // Outbound POST to non-allow-listed domains
  { re: /POST\s+the\s+result\s+to\s+https?:\/\/(?!(?:api\.github|hooks\.slack|api\.stripe|jira\.|atlassian\.|figma\.com|amplitude))/i,
    risk: 'HIGH', category: 'data_exfiltration', label: 'suspicious-post-target' },

  // Hidden text (HTML-based attacks)
  { re: /style\s*=\s*["'][^"']*font-size\s*:\s*0/i,
    risk: 'HIGH', category: 'hidden_text', label: 'font-size-zero' },
  { re: /style\s*=\s*["'][^"']*opacity\s*:\s*0(?:\.\d+)?/i,
    risk: 'HIGH', category: 'hidden_text', label: 'opacity-zero' },
  { re: /style\s*=\s*["'][^"']*display\s*:\s*none/i,
    risk: 'HIGH', category: 'hidden_text', label: 'display-none' },
  { re: /style\s*=\s*["'][^"']*color\s*:\s*#(?:fff(?:fff)?)\b/i,
    risk: 'HIGH', category: 'hidden_text', label: 'white-on-white' },
  { re: /<!--[\s\S]*?(?:ignore|instructions?|system\s+prompt|override|disregard|new\s+task)[\s\S]*?-->/i,
    risk: 'HIGH', category: 'hidden_text', label: 'html-comment-injection' },
  // Two or more zero-width chars within 200 chars of each other
  { re: /[​‌‍⁠﻿][\s\S]{0,200}[​‌‍⁠﻿]/,
    risk: 'HIGH', category: 'hidden_text', label: 'zero-width-chars' },

  // Phishing
  { re: /your\s+(?:account|password|MFA|multi.?factor\s+auth\w*).{0,60}(?:will\s+be|has\s+been|is)\s+(?:suspended|expired?|disabled|blocked)/i,
    risk: 'HIGH', category: 'phishing', label: 'account-threat' },
  { re: /verify\s+(?:your\s+)?(?:credentials?|account|identity|access)\s+immediately/i,
    risk: 'HIGH', category: 'phishing', label: 'immediate-verify' },
  { re: /URGENT\b.{0,120}(?:click\s+here|re.?authenticate|verify\s+immediately)/i,
    risk: 'HIGH', category: 'phishing', label: 'urgency-cta' },
  { re: /\bIT\s+(?:Security|Department|Team)\b/i,
    risk: 'HIGH', category: 'phishing', label: 'it-impersonation' },
  { re: /your\s+password\s+has\s+expired/i,
    risk: 'HIGH', category: 'phishing', label: 'password-expired' },
  // Wire transfer fraud / social engineering
  { re: /wire\s+transfer.{0,120}(do\s+not\s+(discuss|tell|mention|share)|immediately|right\s+away|urgently|asap)/i,
    risk: 'HIGH', category: 'phishing', label: 'wire-transfer-fraud' },
  { re: /(process|initiate|send|complete)\s+.{0,30}wire\s+transfer.{0,80}(immediately|right\s+now|urgently|do\s+not)/i,
    risk: 'HIGH', category: 'phishing', label: 'wire-transfer-fraud' },

  // Secret exposure
  { re: /\bAKIA[0-9A-Z]{16}\b/,
    risk: 'HIGH', category: 'secret_exposure', label: 'aws-access-key' },
  // GitHub PAT classic (ghp_/gho_/ghs_) — 36 chars is standard but allow 35+ for resilience
  { re: /\bgh[ops]_[A-Za-z0-9]{35,}\b/,
    risk: 'HIGH', category: 'secret_exposure', label: 'github-token' },
  { re: /\bgithub_pat_[A-Za-z0-9_]{82}\b/,
    risk: 'HIGH', category: 'secret_exposure', label: 'github-fine-grained-pat' },
  { re: /\bsk-[A-Za-z0-9T]{48,}\b/,
    risk: 'HIGH', category: 'secret_exposure', label: 'openai-key' },
  { re: /\bsk-ant-api0[0-9]-[A-Za-z0-9_-]{80,}/,
    risk: 'HIGH', category: 'secret_exposure', label: 'anthropic-key' },
  { re: /\bsk_live_[A-Za-z0-9]{24,}\b/,
    risk: 'HIGH', category: 'secret_exposure', label: 'stripe-secret' },
  { re: /-----BEGIN\s+(?:RSA\s+|EC\s+|OPENSSH\s+)?PRIVATE\s+KEY-----/,
    risk: 'HIGH', category: 'secret_exposure', label: 'pem-private-key' },
  // npm auth token — 36 chars is standard but allow 35+ for resilience
  { re: /\bnpm_[A-Za-z0-9]{35,}\b/,
    risk: 'HIGH', category: 'secret_exposure', label: 'npm-token' },
]

function checkDecodedBase64(payload: string): ClassifyResult {
  // Find standalone base64-looking blobs (≥20 chars) surrounded by whitespace or quotes
  const b64re = /(?:^|[\s"'])([A-Za-z0-9+/]{20,}={0,2})(?:[\s"']|$)/gm
  for (const m of payload.matchAll(b64re)) {
    const blob = m[1]
    if (!blob) continue
    try {
      const decoded = Buffer.from(blob, 'base64').toString('utf8')
      // Only scan if decoded text is printable ASCII (avoids binary garbage)
      if (!/^[\x20-\x7e\n\r\t]+$/.test(decoded)) continue
      const inner = classify(decoded)
      if (inner.risk !== 'NONE') return inner
    } catch { /* ignore decode errors */ }
  }
  return NONE
}

export function classify(payload: string): ClassifyResult {
  if (!payload) return NONE

  for (const p of PATTERNS) {
    if (p.re.test(payload)) {
      return { risk: p.risk, category: p.category, matchedPattern: p.label }
    }
  }

  // Recurse into decoded base64 blobs
  return checkDecodedBase64(payload)
}
