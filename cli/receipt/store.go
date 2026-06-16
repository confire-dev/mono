package receipt

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Meta holds per-session chain state persisted to a .meta file alongside
// the JSONL receipt log. It is updated atomically after every receipt write
// so the chain tip survives daemon restarts.
type Meta struct {
	SessionID      string    `json:"session_id"`
	ChainTip       string    `json:"chain_tip"`
	ReceiptCount   int       `json:"receipt_count"`
	GenesisHash    string    `json:"genesis_hash"`
	LocalPubkey    string    `json:"local_pubkey"`
	ConfireVersion string    `json:"confire_version"`
	StartedAt      time.Time `json:"started_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Store manages per-session receipt JSONL files and their .meta chain-tip files.
// All exported methods are safe for concurrent use.
type Store struct {
	dir     string
	version string
	pubkey  string
	mu      sync.Mutex
	metas   map[string]*Meta // session key → in-memory meta
}

// NewStore opens (or creates) the receipt storage directory and restores any
// existing session chain tips from .meta files. Call this once at daemon startup.
func NewStore(dir, confireVersion, localPubkey string) (*Store, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	s := &Store{
		dir:     dir,
		version: confireVersion,
		pubkey:  localPubkey,
		metas:   make(map[string]*Meta),
	}
	s.restoreAll()
	return s, nil
}

// ChainTip returns the payload_hash of the last receipt written for the given
// session key, or GenesisHash if no receipts have been written yet.
// Use BundleSessionKey for the policy bundle chain.
func (s *Store) ChainTip(key string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m, ok := s.metas[key]; ok && m.ChainTip != "" {
		return m.ChainTip
	}
	return GenesisHash
}

// Write appends r to its session JSONL file and atomically updates the .meta file.
// The session key is derived from the receipt's phase and event.session_id.
func (s *Store) Write(r Receipt) error {
	key := sessionKey(r)

	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	f, err := os.OpenFile(s.jsonlFile(key), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(data)
	syncErr := f.Sync()
	f.Close()
	if writeErr != nil {
		return writeErr
	}
	if syncErr != nil {
		return syncErr
	}

	s.mu.Lock()
	m, ok := s.metas[key]
	if !ok {
		m = &Meta{
			SessionID:      key,
			GenesisHash:    GenesisHash,
			LocalPubkey:    s.pubkey,
			ConfireVersion: s.version,
			StartedAt:      time.Now().UTC(),
		}
		s.metas[key] = m
	}
	m.ChainTip = r.PayloadHash
	m.ReceiptCount++
	m.UpdatedAt = time.Now().UTC()
	snap := *m
	s.mu.Unlock()

	return s.flushMeta(snap)
}

// ReadChain reads all receipts for the given session key from the JSONL file.
func (s *Store) ReadChain(key string) ([]Receipt, error) {
	return ReadChainFromFile(s.jsonlFile(key))
}

// ReadChainFromFile reads all receipts from a JSONL file at the given path.
func ReadChainFromFile(path string) ([]Receipt, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var receipts []Receipt
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var r Receipt
		if err := json.Unmarshal([]byte(line), &r); err != nil {
			return receipts, err
		}
		receipts = append(receipts, r)
	}
	return receipts, nil
}

// ── internal ──────────────────────────────────────────────────────────────────

// sessionKey maps a receipt to its storage key.
// Policy bundle receipts use BundleSessionKey so they go to daemon-policy.jsonl.
func sessionKey(r Receipt) string {
	if r.Phase == PhasePolicyBundleLoaded {
		return BundleSessionKey
	}
	if r.Event != nil && r.Event.SessionID != "" {
		return r.Event.SessionID
	}
	return "_unknown_"
}

func (s *Store) jsonlFile(key string) string {
	if key == BundleSessionKey {
		return filepath.Join(s.dir, "daemon-policy.jsonl")
	}
	return filepath.Join(s.dir, key+".jsonl")
}

func (s *Store) metaFile(key string) string {
	if key == BundleSessionKey {
		return filepath.Join(s.dir, "daemon-policy.meta")
	}
	return filepath.Join(s.dir, key+".meta")
}

// flushMeta writes m to disk atomically using a temp-file rename.
func (s *Store) flushMeta(m Meta) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.metaFile(m.SessionID) + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.metaFile(m.SessionID))
}

// restoreAll scans the store directory for .meta files and reloads chain tips.
// Called once at daemon startup to resume in-progress session chains.
func (s *Store) restoreAll() {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		name := e.Name()
		// Skip temp files and non-meta files.
		if !strings.HasSuffix(name, ".meta") || strings.HasSuffix(name, ".meta.tmp") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.dir, name))
		if err != nil {
			continue
		}
		var m Meta
		if err := json.Unmarshal(data, &m); err != nil || m.SessionID == "" {
			continue
		}
		s.mu.Lock()
		s.metas[m.SessionID] = &m
		s.mu.Unlock()
	}
}
