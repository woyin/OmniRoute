package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var skillToolDirs = map[string]string{
	"claude": ".claude/skills", "codex": ".codex/skills", "hermes": "AppData/Local/hermes/skills",
	"opencode": ".opencode/skills", "gemini": ".gemini/skills", "cursor": ".cursor/skills",
	"copilot": ".copilot/skills", "cline": ".cline/skills", "windsurf": ".windsurf/skills",
	"devin": ".devin/skills", "antigravity": ".antigravity/skills", "qwen": ".qwen/skills",
	"kilocode": ".kilocode/skills", "openclaw": ".openclaw/skills", "droid": ".droid/skills", "continue": ".continue/skills",
}

func skillCollectInstallHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			RepoName    string   `json:"repoName"`
			Targets     []string `json:"targets"`
			Description string   `json:"description"`
		}
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil || strings.TrimSpace(body.RepoName) == "" || len(body.Targets) < 1 || len(body.Targets) > 10 {
			jsonError(w, http.StatusBadRequest, "Invalid request")
			return
		}
		parts := strings.Split(strings.Trim(body.RepoName, "/"), "/")
		skillName := parts[len(parts)-1]
		category := inferSkillCategory(skillName + " " + body.Description)
		home, _ := os.UserHomeDir()
		results := make([]map[string]interface{}, 0, len(body.Targets))
		allOK := true
		for _, target := range body.Targets {
			base, ok := skillToolDirs[target]
			if !ok {
				allOK = false
				results = append(results, map[string]interface{}{"target": target, "ok": false, "action": "error", "error": `Unknown target tool: "` + target + `"`})
				continue
			}
			dest := filepath.Join(home, filepath.FromSlash(base), category, skillName)
			results = append(results, map[string]interface{}{"target": target, "ok": true, "action": "planned", "destDir": dest, "note": "Ready: SKILL.md from " + body.RepoName + " can be synced to " + dest})
		}
		writeJSONResponse(w, map[string]interface{}{"ok": allOK, "repoName": body.RepoName, "skillName": skillName, "results": results})
	}
}

func inferSkillCategory(text string) string {
	text = strings.ToLower(text)
	categories := []struct {
		name  string
		words []string
	}{
		{"security", []string{"security", "pentest", "exploit", "malware", "forensics", "vulnerability"}},
		{"data-science", []string{"data", "analytics", "pandas", "ml", "model", "train"}},
		{"devops", []string{"deploy", "docker", "k8s", "terraform", "ci/cd", "pipeline"}},
		{"creative", []string{"design", "image", "video", "art", "music"}},
		{"productivity", []string{"email", "doc", "slide", "report", "calendar"}},
		{"research", []string{"paper", "arxiv", "academic", "literature"}},
		{"software-development", []string{"code", "refactor", "test", "lint", "review", "debug"}},
		{"media", []string{"youtube", "transcript", "gif", "video", "audio"}},
	}
	for _, category := range categories {
		for _, word := range category.words {
			if strings.Contains(text, word) {
				return category.name
			}
		}
	}
	return "imported-github"
}
