// Command updatechecksums recomputes SHA-256 checksums for curl_bash agents.
//
// Usage (from repo root):
//
//	go run ./scripts/updatechecksums
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alebak/squad-ai/internal/registry"
)

const registryRelPath = "registry/agents.json"

func main() {
	os.Exit(run())
}

func run() int {
	failedPath := failedFilePath()
	cat, err := loadCatalog(registryRelPath)
	if err != nil {
		return failLocal(failedPath, err.Error())
	}

	today := time.Now().UTC().Format("2006-01-02")
	client := &http.Client{Timeout: 60 * time.Second}
	changedIDs, summary, failed := refreshChecksums(cat, client, today)

	if len(changedIDs) > 0 {
		if err := writeCatalog(registryRelPath, cat, today); err != nil {
			return failLocal(failedPath, err.Error())
		}
		fmt.Printf("updated registry for: %s\n", strings.Join(changedIDs, ", "))
	} else {
		fmt.Println("no checksum content changes")
	}

	if err := writeFailedFile(failedPath, failed); err != nil {
		fmt.Fprintf(os.Stderr, "warning: writing failure file: %v\n", err)
	}

	emitSummary(len(changedIDs) > 0, len(failed) > 0, changedIDs, summary)
	if len(failed) > 0 {
		return 1
	}
	return 0
}

func failedFilePath() string {
	if p := os.Getenv("SQUAD_CHECKSUM_FAILED_PATH"); p != "" {
		return p
	}
	return filepath.Join(os.TempDir(), "checksums-failed.txt")
}

func loadCatalog(path string) (*registry.Catalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s not readable: %w", path, err)
	}
	var cat registry.Catalog
	if err := json.Unmarshal(data, &cat); err != nil {
		return nil, fmt.Errorf("cannot parse registry: %w", err)
	}
	if len(cat.Agents) == 0 {
		return nil, errors.New("registry missing agents list")
	}
	return &cat, nil
}

func refreshChecksums(cat *registry.Catalog, client *http.Client, today string) (changedIDs, summary, failed []string) {
	for i := range cat.Agents {
		changed, line, fail := refreshAgent(&cat.Agents[i], client, today)
		if fail != "" {
			failed = append(failed, fail)
			continue
		}
		if changed {
			changedIDs = append(changedIDs, cat.Agents[i].ID)
			summary = append(summary, line)
		}
	}
	return changedIDs, summary, failed
}

func refreshAgent(a *registry.Agent, client *http.Client, today string) (changed bool, summaryLine, fail string) {
	if a.Install.URL == "" {
		return false, "", ""
	}
	if err := validateHTTPSURL(a.Install.URL); err != nil {
		msg := fmt.Sprintf("FAILED|%s|%s|%v", a.ID, a.Install.URL, err)
		fmt.Fprintf(os.Stderr, "warning: %s\n", msg)
		return false, "", msg
	}

	newHash, err := downloadSHA256(client, a.Install.URL)
	if err != nil {
		msg := fmt.Sprintf("FAILED|%s|%s|%v", a.ID, a.Install.URL, err)
		fmt.Fprintf(os.Stderr, "warning: %s\n", msg)
		return false, "", msg
	}

	current := ""
	if a.Checksum != nil {
		current = a.Checksum.SHA256
	}
	if current == newHash {
		fmt.Printf("ok: %s\n", a.ID)
		return false, "", ""
	}

	fmt.Printf("changed: %s: %s → %s\n", a.ID, orNone(current), newHash)
	a.Checksum = &registry.Checksum{
		SHA256:           newHash,
		ContentChangedAt: today,
		VerifiedAt:       today,
	}
	line := fmt.Sprintf("%s: `%s` → `%s` (%s)", a.ID, orNone(current), newHash, a.Install.URL)
	return true, line, ""
}

func writeCatalog(path string, cat *registry.Catalog, today string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read registry for patch: %w", err)
	}
	doc, agents, err := parseRegistryDocument(raw)
	if err != nil {
		return err
	}
	if err := patchAgentChecksums(agents, cat); err != nil {
		return err
	}
	return writeRegistryDocument(path, doc, agents, today)
}

func parseRegistryDocument(raw []byte) (map[string]json.RawMessage, []map[string]json.RawMessage, error) {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, nil, fmt.Errorf("parse registry document: %w", err)
	}
	var agents []map[string]json.RawMessage
	if err := json.Unmarshal(doc["agents"], &agents); err != nil {
		return nil, nil, fmt.Errorf("parse agents array: %w", err)
	}
	return doc, agents, nil
}

func patchAgentChecksums(agents []map[string]json.RawMessage, cat *registry.Catalog) error {
	byID := make(map[string]*registry.Checksum, len(cat.Agents))
	for i := range cat.Agents {
		byID[cat.Agents[i].ID] = cat.Agents[i].Checksum
	}
	for i := range agents {
		var id string
		if err := json.Unmarshal(agents[i]["id"], &id); err != nil {
			return fmt.Errorf("parse agent id: %w", err)
		}
		ck, ok := byID[id]
		if !ok || ck == nil {
			continue
		}
		ckRaw, err := json.Marshal(ck)
		if err != nil {
			return fmt.Errorf("marshal checksum for %s: %w", id, err)
		}
		agents[i]["checksum"] = ckRaw
	}
	return nil
}

func writeRegistryDocument(path string, doc map[string]json.RawMessage, agents []map[string]json.RawMessage, today string) error {
	agentsRaw, err := json.Marshal(agents)
	if err != nil {
		return fmt.Errorf("marshal agents array: %w", err)
	}
	doc["agents"] = agentsRaw
	updatedRaw, err := json.Marshal(today)
	if err != nil {
		return fmt.Errorf("marshal updated_at: %w", err)
	}
	doc["updated_at"] = updatedRaw
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}
	out = append(out, 10) // newline
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return fmt.Errorf("write registry: %w", err)
	}
	return nil
}

func writeFailedFile(path string, failed []string) error {
	if len(failed) == 0 {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}
	return os.WriteFile(path, []byte(strings.Join(failed, "\n")+"\n"), 0o644)
}

func downloadSHA256(client *http.Client, rawURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "squad-ai-checksum-updater/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: closing response body: %v\n", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}
	h := sha256.New()
	if _, err := io.Copy(h, resp.Body); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func validateHTTPSURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "https" {
		return errors.New("URL must use HTTPS scheme")
	}
	if u.Host == "" {
		return errors.New("URL must have a host")
	}
	host := strings.ToLower(u.Hostname())
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return errors.New("URL must not target localhost")
	}
	if strings.ContainsAny(rawURL, "\r\n") {
		return errors.New("URL must not contain newlines")
	}
	return nil
}

func failLocal(failedPath, message string) int {
	fmt.Fprintf(os.Stderr, "error: %s\n", message)
	if err := os.WriteFile(failedPath, []byte("LOCAL|"+message+"\n"), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "warning: writing failure file: %v\n", err)
	}
	emitSummary(false, true, nil, nil)
	return 2
}

func emitSummary(changed, failed bool, changedIDs, summary []string) {
	fmt.Printf("CHANGED=%t\n", changed)
	fmt.Printf("FAILED=%t\n", failed)
	fmt.Printf("CHANGED_IDS=%s\n", strings.Join(changedIDs, ","))
	fmt.Printf("SUMMARY=%s\n", strings.Join(summary, " | "))
}

func orNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}
