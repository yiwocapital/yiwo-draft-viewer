package app

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
		item := map[string]interface{}{"op": int(sg.Op), "text": sg.Text}
		if sg.IsComment {
			item["isComment"] = true
		}
		flat = append(flat, item)
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
	return model.Result{Ok: true, Data: map[string]interface{}{"text": cleanForCopy(text)}}
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

var (
	htmlCommentRE = regexp.MustCompile(`(?s)<!--.*?-->`)
	blankLineRE   = regexp.MustCompile(`\n{3,}`)
)

// cleanForCopy removes HTML comment blocks and collapses runs of 3+ newlines
// (i.e. 2+ consecutive blank lines) into a single blank line, then trims any
// leading/trailing whitespace. Used for all clipboard output so the user pastes
// clean text without author notes or excessive spacing.
//
// The blank-line collapse runs in a loop because stripping a comment that
// sits between blank lines can leave a longer run than `\n{3,}` matches in a
// single pass (e.g. `a\n\n\n\n<!-- x -->\nb` becomes `a\n\n\n\n\nb` after
// the comment is removed).
func cleanForCopy(text string) string {
	text = htmlCommentRE.ReplaceAllString(text, "")
	for blankLineRE.MatchString(text) {
		text = blankLineRE.ReplaceAllString(text, "\n\n")
	}
	text = strings.TrimSpace(text)
	return text
}

// splitOutComments emits each character's source segment (its op) plus a
// running "is the cursor inside an HTML comment?" flag. It re-emits characters
// as segments, joining adjacent characters that share (op, isComment).
//
// This correctly handles comments that diff-match-patch splits across segments
// (e.g. when a comment is modified between two versions): the `<!--...-->`
// regex is applied to the reconstructed left-side and right-side text, not to
// each segment in isolation, so cross-segment boundaries don't break the match.
func splitOutComments(segs []model.DiffSegment) []model.DiffSegment {
	// We walk each character and remember which side (left/right) of the diff
	// it belongs to. Equal characters appear on both sides; delete characters
	// only on the left; insert characters only on the right. We build the
	// left-side text (equal+delete) and right-side text (equal+insert) so we
	// can run the comment regex on each, and then mark a character as being
	// inside a comment if it falls inside any match on its own side.
	type charRec struct {
		op     model.DiffOp
		raw    string // original UTF-8 bytes of the rune
		leftIx int    // byte offset in left text, or -1 if not on left
		rightIx int   // byte offset in right text, or -1 if not on right
	}
	chars := make([]charRec, 0)
	leftBuf := make([]byte, 0)
	rightBuf := make([]byte, 0)
	for _, s := range segs {
		for _, r := range s.Text {
			raw := string(r)
			rec := charRec{op: s.Op, raw: raw}
			if s.Op != model.DiffInsert {
				rec.leftIx = len(leftBuf)
				leftBuf = append(leftBuf, raw...)
			} else {
				rec.leftIx = -1
			}
			if s.Op != model.DiffDelete {
				rec.rightIx = len(rightBuf)
				rightBuf = append(rightBuf, raw...)
			} else {
				rec.rightIx = -1
			}
			chars = append(chars, rec)
		}
	}

	// Run the regex on each side.
	leftMatches := htmlCommentRE.FindAllIndex(leftBuf, -1)
	rightMatches := htmlCommentRE.FindAllIndex(rightBuf, -1)

	// Helper: returns true if byte offset x falls inside any of the given matches.
	contains := func(matches [][]int, x int) bool {
		for _, m := range matches {
			if x >= m[0] && x < m[1] {
				return true
			}
		}
		return false
	}

	// Mark each character as being inside a comment if its left offset is
	// inside a left-side match, or its right offset is inside a right-side
	// match.
	inComment := make([]bool, len(chars))
	for ci, c := range chars {
		leftHit := c.leftIx >= 0 && contains(leftMatches, c.leftIx)
		rightHit := c.rightIx >= 0 && contains(rightMatches, c.rightIx)
		inComment[ci] = leftHit || rightHit
	}

	if len(chars) == 0 {
		return nil
	}

	// Walk the chars, emitting segments. Group consecutive chars that share
	// (op, isComment). The output preserves the original segment ordering.
	var out []model.DiffSegment
	var buf []byte
	var curOp model.DiffOp
	var curComment bool
	flush := func() {
		if len(buf) == 0 {
			return
		}
		out = append(out, model.DiffSegment{Op: curOp, Text: string(buf), IsComment: curComment})
		buf = buf[:0]
	}
	for ci, c := range chars {
		if len(buf) == 0 {
			curOp = c.op
			curComment = inComment[ci]
		} else if c.op != curOp || inComment[ci] != curComment {
			flush()
			curOp = c.op
			curComment = inComment[ci]
		}
		buf = append(buf, c.raw...)
	}
	flush()
	return out
}

// silence unused-import warning when fmt not referenced elsewhere.