package git

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yiwocapital/yiwo-draft-viewer/internal/model"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Repo struct {
	repo     *gogit.Repository
	workTree string
}

func Open(filePath string) (*Repo, error) {
	dir := filepath.Dir(filePath)
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			repo, err := gogit.PlainOpen(dir)
			if err != nil {
				return nil, fmt.Errorf("open repo: %w", err)
			}
			wt, err := repo.Worktree()
			if err != nil {
				return nil, err
			}
			return &Repo{repo: repo, workTree: wt.Filesystem.Root()}, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, nil
		}
		dir = parent
	}
}

func (r *Repo) Commits(filePath string) []model.Commit {
	if r == nil {
		return nil
	}
	iter, err := r.repo.Log(&gogit.LogOptions{
		FileName: &filePath,
	})
	if err != nil {
		return nil
	}
	defer iter.Close()

	out := []model.Commit{}
	err = iter.ForEach(func(c *object.Commit) error {
		firstLine := c.Message
		hasMore := false
		for i, r := range c.Message {
			if r == '\n' {
				firstLine = c.Message[:i]
				hasMore = true
				break
			}
		}
		out = append(out, model.Commit{
			Hash:      c.Hash.String(),
			ShortHash: c.Hash.String()[:7],
			Message:   c.Message,
			FirstLine: firstLine,
			HasMore:   hasMore,
			Timestamp: c.Author.When.Unix(),
		}) // gogit Log already yields reverse-chronological (newest first)
		return nil
	})
	if err != nil {
		return nil
	}
	return out
}

func (r *Repo) Blob(hash, filePath string) string {
	if r == nil {
		return ""
	}
	c, err := r.repo.CommitObject(plumbing.NewHash(hash))
	if err != nil {
		return ""
	}
	f, err := c.File(filePath)
	if err != nil {
		return ""
	}
	reader, err := f.Reader()
	if err != nil {
		return ""
	}
	defer reader.Close()
	buf := make([]byte, f.Size)
	_, err = reader.Read(buf)
	if err != nil {
		return ""
	}
	return string(buf)
}

func (r *Repo) WorkingTree(filePath string) string {
	data, err := os.ReadFile(filepath.Join(r.workTree, filePath))
	if err != nil {
		return ""
	}
	return string(data)
}
