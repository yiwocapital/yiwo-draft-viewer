package app

import (
	"context"
	"path/filepath"
	"time"

	"github.com/yiwocapital/yiwo-draft-viewer/internal/config"
	"github.com/yiwocapital/yiwo-draft-viewer/internal/diff"
	"github.com/yiwocapital/yiwo-draft-viewer/internal/file"
	"github.com/yiwocapital/yiwo-draft-viewer/internal/git"
	"github.com/yiwocapital/yiwo-draft-viewer/internal/model"
	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Service struct {
	ctx         context.Context
	cfg         config.Config
	configDir   string
	currentPath string
	content     string
	title       string
	summary     string
	body        string
	hasFm       bool
	repo        *git.Repo
	watcher     *fsnotify.Watcher
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Startup(ctx context.Context) {
	s.ctx = ctx
	s.SetConfigDir(".")
}

func (s *Service) Ctx() context.Context {
	return s.ctx
}

func (s *Service) SetConfigDir(dir string) {
	s.configDir = dir
	if cfg, err := config.Load(dir); err == nil {
		s.cfg = cfg
	}
}

func (s *Service) OpenFile(path string) model.Result {
	content, err := file.Load(path)
	if err != nil {
		return model.Result{Ok: false, Error: err.Error()}
	}
	fm, body := file.ParseFrontmatter(content)
	repo, _ := git.Open(path)
	s.currentPath = path
	s.content = content
	s.title = fm.Title
	s.summary = fm.Summary
	s.body = body
	s.hasFm = fm.HasFrontmatter
	s.repo = repo
	s.startWatcher(path)
	return model.Result{Ok: true, Data: map[string]interface{}{
		"path": path, "content": content,
		"title": fm.Title, "summary": fm.Summary,
		"hasFrontmatter": fm.HasFrontmatter,
		"charCount":      file.CountChars(content),
		"hasGit":         repo != nil,
	}}
}

func (s *Service) ListCommits() model.Result {
	if s.repo == nil {
		return model.Result{Ok: true, Data: map[string]interface{}{"items": []interface{}{}}}
	}
	rel, _ := filepath.Rel(s.repo.WorkTreePath(), s.currentPath)
	commits := s.repo.Commits(rel)
	items := []map[string]interface{}{{
		"hash": "WORKING", "shortHash": "未提交",
		"message": "未提交的修改", "firstLine": "未提交",
		"hasMore": false, "timestamp": time.Now().Unix(),
		"isUnstaged": true,
	}}
	for _, c := range commits {
		items = append(items, map[string]interface{}{
			"hash":      c.Hash,
			"shortHash": c.ShortHash,
			"message":   c.Message,
			"firstLine": c.FirstLine,
			"hasMore":   c.HasMore,
			"timestamp": c.Timestamp,
		})
	}
	return model.Result{Ok: true, Data: map[string]interface{}{"items": items}}
}

func (s *Service) GetDiff(left, right string) model.Result {
	rel, _ := filepath.Rel(s.repo.WorkTreePath(), s.currentPath)
	leftContent := s.resolveContent(left, rel)
	rightContent := s.resolveContent(right, rel)
	if rightContent == "" {
		return model.Result{Ok: true, Data: map[string]interface{}{
			"segments": []interface{}{}, "charCount": 0, "static": true,
		}}
	}
	segs := diff.Compute(leftContent, rightContent)
	flat := make([]map[string]interface{}, 0, len(segs))
	for _, sg := range segs {
		flat = append(flat, map[string]interface{}{"op": int(sg.Op), "text": sg.Text})
	}
	return model.Result{Ok: true, Data: map[string]interface{}{
		"segments":  flat,
		"charCount": file.CountChars(rightContent),
		"static":    false,
	}}
}

func (s *Service) resolveContent(ref, rel string) string {
	if ref == "WORKING" {
		return s.content
	}
	if ref == "" {
		return ""
	}
	return s.repo.Blob(ref, rel)
}

func (s *Service) CopySection(kind string) model.Result {
	var text string
	switch kind {
	case "title":
		if !s.hasFm {
			return model.Result{Ok: false, Error: "no frontmatter"}
		}
		text = s.title
	case "summary":
		if !s.hasFm {
			return model.Result{Ok: false, Error: "no frontmatter"}
		}
		text = s.summary
	case "body":
		text = s.body
	case "all":
		text = s.content
	default:
		return model.Result{Ok: false, Error: "unknown section"}
	}
	return model.Result{Ok: true, Data: map[string]interface{}{"text": text}}
}

func (s *Service) Reload() model.Result {
	if s.currentPath == "" {
		return model.Result{Ok: false, Error: "no file open"}
	}
	return s.OpenFile(s.currentPath)
}

func (s *Service) startWatcher(path string) {
	if s.watcher != nil {
		s.watcher.Close()
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	s.watcher = w
	dir := filepath.Dir(path)
	w.Add(dir)
	if s.repo != nil {
		w.Add(filepath.Join(s.repo.WorkTreePath(), ".git"))
	}
	go func() {
		var debounce *time.Timer
		for {
			select {
			case _, ok := <-w.Events:
				if !ok {
					return
				}
				if debounce != nil {
					debounce.Stop()
				}
				debounce = time.AfterFunc(200*time.Millisecond, func() {
					r := s.OpenFile(s.currentPath)
					if r.Ok && s.ctx != nil {
						runtime.EventsEmit(s.ctx, "reloaded", r.Data)
					}
				})
			case _, ok := <-w.Errors:
				if !ok {
					return
				}
			}
		}
	}()
}

// silence unused-import warning when fmt not referenced elsewhere.