package app

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
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
	ctx             context.Context
	cfg             config.Config
	configDir       string
	currentFontSize int
	currentPath     string
	content         string
	title           string
	summary         string
	body            string
	hasFm           bool
	repo            *git.Repo
	watcher         *fsnotify.Watcher
	foldComments    bool // NEW
}

func NewService() *Service {
	return &Service{}
}

// defaultConfigDir returns the macOS-standard per-user config directory for
// the app. Using a stable, OS-managed location (rather than the current
// working directory) means the path is the same whether the app is launched
// from Finder, the dock, or a terminal.
func defaultConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, "Library", "Application Support", "YiwoDraftViewer")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func (s *Service) Startup(ctx context.Context) {
	s.ctx = ctx
	if dir, err := defaultConfigDir(); err == nil {
		s.SetConfigDir(dir)
	} else {
		s.SetConfigDir(".")
	}
	if cfg, err := config.Load(s.configDir); err == nil {
		s.cfg = cfg
		s.currentFontSize = cfg.FontSize
	}
}

func (s *Service) Ctx() context.Context {
	return s.ctx
}

// ConfigDir returns the directory where setting.local.yaml is read from and
// written to. Used by main.go's OnStartup to load saved window state.
func (s *Service) ConfigDir() string {
	return s.configDir
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
	s.updateWindowTitle()
	return model.Result{Ok: true, Data: map[string]interface{}{
		"path": path, "content": content,
		"title": fm.Title, "summary": fm.Summary,
		"hasFrontmatter": fm.HasFrontmatter,
		"charCount":      file.CountChars(content),
		"hasGit":         repo != nil,
		"fontSize":       s.cfg.FontSize,
	}}
}

func (s *Service) ListCommits() model.Result {
	if s.repo == nil {
		return model.Result{Ok: true, Data: map[string]interface{}{"items": []interface{}{}}}
	}
	rel, _ := filepath.Rel(s.repo.WorkTreePath(), s.currentPath)
	commits := s.repo.Commits(rel)

	items := []map[string]interface{}{}

	// Determine whether the working tree actually differs from HEAD.
	hasUnstaged := false
	if len(commits) > 0 {
		headContent := s.repo.Blob(commits[0].Hash, rel)
		hasUnstaged = (s.content != headContent)
	} else {
		// No commits yet: if file has any content, it's all "unstaged"
		hasUnstaged = (s.content != "")
	}

	if hasUnstaged {
		items = append(items, map[string]interface{}{
			"hash": "WORKING", "shortHash": "未提交",
			"message": "未提交的修改", "firstLine": "未提交",
			"hasMore": false, "timestamp": time.Now().Unix(),
			"isUnstaged": true,
		})
	}
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

	// Strip comments when toggle is on. This must happen BEFORE diff.Compute
	// so that no diff segment ever contains a comment fragment.
	if s.foldComments {
		leftContent = file.StripComments(leftContent)
		rightContent = file.StripComments(rightContent)
	}

	if rightContent == "" {
		return model.Result{Ok: true, Data: map[string]interface{}{
			"segments": []interface{}{}, "charCount": 0, "static": true,
		}}
	}
	segs := diff.Compute(leftContent, rightContent)

	// Split out comment fragments so the frontend can render them gray.
	// (Skip this when foldComments is on — there shouldn't be any comments
	// in leftContent/rightContent at that point.)
	if !s.foldComments {
		segs = splitOutComments(segs)
	}

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
	res := s.OpenFile(s.currentPath)
	return res
}

func (s *Service) CloseFile() model.Result {
	if s.watcher != nil {
		s.watcher.Close()
		s.watcher = nil
	}
	s.currentPath = ""
	s.content = ""
	s.title = ""
	s.summary = ""
	s.body = ""
	s.hasFm = false
	s.repo = nil
	s.updateWindowTitle()
	return model.Result{Ok: true}
}

func (s *Service) updateWindowTitle() {
	// No-op: macOS shows the app name in menu bar; the window title is also
	// fixed to "YiwoDraftViewer". File path is shown in the in-app status bar.
}

func (s *Service) SetFontSize(size int) model.Result {
	if size < 10 || size > 32 {
		return model.Result{Ok: false, Error: "font size out of range (10-32)"}
	}
	s.cfg.FontSize = size
	s.currentFontSize = size
	if err := config.Save(s.configDir, s.cfg); err != nil {
		return model.Result{Ok: false, Error: err.Error()}
	}
	return model.Result{Ok: true, Data: map[string]interface{}{"fontSize": size}}
}

func (s *Service) GetFontSize() int {
	return s.cfg.FontSize
}

func (s *Service) SetFoldComments(enabled bool) model.Result {
	s.foldComments = enabled
	return model.Result{Ok: true, Data: map[string]interface{}{"foldComments": enabled}}
}

func (s *Service) WindowChanged() model.Result {
	if s.ctx == nil {
		return model.Result{Ok: false, Error: "no context"}
	}
	w, h := runtime.WindowGetSize(s.ctx)
	x, y := runtime.WindowGetPosition(s.ctx)
	s.cfg.Window.Width = w
	s.cfg.Window.Height = h
	s.cfg.Window.X = x
	s.cfg.Window.Y = y
	if err := config.Save(s.configDir, s.cfg); err != nil {
		return model.Result{Ok: false, Error: err.Error()}
	}
	return model.Result{Ok: true, Data: map[string]interface{}{
		"width": w, "height": h, "x": x, "y": y,
	}}
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

var htmlCommentRE = regexp.MustCompile(`(?s)<!--.*?-->`)

// splitOutComments walks each segment and splits it into pieces, where any
// piece matching an HTML comment is emitted as its own DiffComment segment.
func splitOutComments(segs []model.DiffSegment) []model.DiffSegment {
	var out []model.DiffSegment
	for _, s := range segs {
		text := s.Text
		if text == "" {
			continue
		}
		last := 0
		for _, m := range htmlCommentRE.FindAllStringIndex(text, -1) {
			if m[0] > last {
				out = append(out, model.DiffSegment{Op: s.Op, Text: text[last:m[0]]})
			}
			out = append(out, model.DiffSegment{Op: model.DiffComment, Text: text[m[0]:m[1]]})
			last = m[1]
		}
		if last < len(text) {
			out = append(out, model.DiffSegment{Op: s.Op, Text: text[last:]})
		}
	}
	return out
}

// silence unused-import warning when fmt not referenced elsewhere.