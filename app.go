package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	os.MkdirAll(a.getDataDir(), 0755)
	a.StartProxyServer()
}

func (a *App) StartProxyServer() {
	http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		targetURL := r.Header.Get("X-Target-Url")
		if targetURL == "" {
			http.Error(w, "Missing X-Target-Url", http.StatusBadRequest)
			return
		}

		req, err := http.NewRequest(r.Method, targetURL, r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for name, values := range r.Header {
			if name != "X-Target-Url" && name != "Origin" && name != "Referer" {
				for _, v := range values {
					req.Header.Add(name, v)
				}
			}
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.Header().Set("Cache-Control", "no-cache")
		
		// Copy response chunk by chunk for streaming
		buf := make([]byte, 4096)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				w.Write(buf[:n])
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			}
			if err != nil {
				break
			}
		}
	})

	go http.ListenAndServe("127.0.0.1:45543", nil)
}

func (a *App) getDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "diagrams" // fallback
	}
	return filepath.Join(home, ".smart-mermaid", "diagrams")
}

type Diagram struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Content   string `json:"content"`
	Synced    bool   `json:"synced"`
	UpdatedAt string `json:"updatedAt"`
	Sha       string                   `json:"sha,omitempty"`
	Model     string                   `json:"model,omitempty"`
	History   []map[string]interface{} `json:"history,omitempty"`
}

// GetDiagrams returns list of diagrams
func (a *App) GetDiagrams() ([]Diagram, error) {
	dir := a.getDataDir()
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var diagrams []Diagram
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
			content, err := os.ReadFile(filepath.Join(dir, f.Name()))
			if err != nil {
				continue
			}
			var d Diagram
			if err := json.Unmarshal(content, &d); err == nil {
				diagrams = append(diagrams, d)
			}
		}
	}
	return diagrams, nil
}

// SaveDiagram saves diagram locally
func (a *App) SaveDiagram(diagram Diagram) error {
	if diagram.ID == "" {
		diagram.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	diagram.UpdatedAt = time.Now().Format(time.RFC3339)
	
	data, err := json.MarshalIndent(diagram, "", "  ")
	if err != nil {
		return err
	}
	
	filename := filepath.Join(a.getDataDir(), diagram.ID+".json")
	return os.WriteFile(filename, data, 0644)
}

// DeleteDiagram deletes diagram locally
func (a *App) DeleteDiagram(id string) error {
	filename := filepath.Join(a.getDataDir(), id+".json")
	return os.Remove(filename)
}

// SyncToGitHub pushes to github
func (a *App) SyncToGitHub(token, repo, id string) (string, error) {
	filename := filepath.Join(a.getDataDir(), id+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	var d Diagram
	if err := json.Unmarshal(data, &d); err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s.mmd", repo, d.Name)
	
	reqBody := map[string]interface{}{
		"message": "Sync diagram " + d.Name + " from Smart Mermaid",
		"content": base64.StdEncoding.EncodeToString([]byte(d.Content)),
	}
	if d.Sha != "" {
		reqBody["sha"] = d.Sha
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub API Error: %s", string(bodyBytes))
	}

	var respData struct {
		Content struct {
			Sha string `json:"sha"`
		} `json:"content"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&respData); err == nil {
		d.Sha = respData.Content.Sha
		d.Synced = true
		a.SaveDiagram(d)
	}

	return "Synced successfully", nil
}

// GetGitHubTokenFromCLI fetches the GitHub token using the gh cli
func (a *App) GetGitHubTokenFromCLI() (string, error) {
	cmd := exec.Command("gh", "auth", "token")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("GitHub CLI not authenticated or not installed. Please run 'gh auth login' in your terminal")
	}
	token := strings.TrimSpace(out.String())
	if token == "" {
		return "", fmt.Errorf("GitHub CLI returned empty token")
	}
	return token, nil
}

// GetGitHubRepos fetches user repositories
func (a *App) GetGitHubRepos(token string) ([]string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/repos?per_page=100&sort=updated", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("GitHub API Error: %d", resp.StatusCode)
	}

	var repos []struct {
		FullName string `json:"full_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	var repoNames []string
	for _, r := range repos {
		repoNames = append(repoNames, r.FullName)
	}
	return repoNames, nil
}

// TestAIConnection tests the connection to the given AI base URL
func (a *App) TestAIConnection(baseUrl, apiKey string) (string, error) {
	req, err := http.NewRequest("GET", baseUrl+"/models", nil)
	if err != nil {
		return "", err
	}
	
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Failed to connect: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("API returned %d", resp.StatusCode)
	}
	
	return "success", nil
}

// FetchDiagramsFromGitHub fetches .mmd files from repo and syncs locally
func (a *App) FetchDiagramsFromGitHub(token, repo string) (int, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/contents", repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("GitHub API Error: %s", string(bodyBytes))
	}

	var contents []struct {
		Name        string `json:"name"`
		DownloadURL string `json:"download_url"`
		Sha         string `json:"sha"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return 0, err
	}

	localDiagrams, _ := a.GetDiagrams()
	syncedCount := 0

	for _, item := range contents {
		if strings.HasSuffix(item.Name, ".mmd") && item.DownloadURL != "" {
			// Download content
			dlReq, err := http.NewRequest("GET", item.DownloadURL, nil)
			if err != nil {
				continue
			}
			dlResp, err := client.Do(dlReq)
			if err != nil || dlResp.StatusCode != 200 {
				if dlResp != nil {
					dlResp.Body.Close()
				}
				continue
			}
			contentBytes, _ := io.ReadAll(dlResp.Body)
			dlResp.Body.Close()

			diagramName := strings.TrimSuffix(item.Name, ".mmd")
			
			// Find existing local diagram
			var targetDiagram Diagram
			found := false
			for _, d := range localDiagrams {
				if d.Name == diagramName {
					targetDiagram = d
					found = true
					break
				}
			}

			if !found {
				targetDiagram = Diagram{
					ID:   fmt.Sprintf("%d", time.Now().UnixNano()),
					Name: diagramName,
				}
			}

			targetDiagram.Content = string(contentBytes)
			targetDiagram.Sha = item.Sha
			targetDiagram.Synced = true
			targetDiagram.UpdatedAt = time.Now().Format(time.RFC3339)

			a.SaveDiagram(targetDiagram)
			syncedCount++
		}
	}

	return syncedCount, nil
}

// DeleteFromGitHub deletes a diagram from GitHub repository
func (a *App) DeleteFromGitHub(token, repo, name, sha string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s.mmd", repo, name)
	
	reqBody := map[string]interface{}{
		"message": "Delete diagram " + name + " via Smart Mermaid",
		"sha":     sha,
	}
	jsonBody, _ := json.Marshal(reqBody)
	
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API Error: %s", string(bodyBytes))
	}

	return nil
}
