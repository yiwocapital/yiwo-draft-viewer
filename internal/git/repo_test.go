package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func initRepo(t *testing.T, dir string) {
	t.Helper()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "test")
	run("config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(dir, "test.md"), []byte("first"), 0644); err != nil {
		t.Fatal(err)
	}
	run("add", "test.md")
	run("commit", "-m", "first commit")
	if err := os.WriteFile(filepath.Join(dir, "test.md"), []byte("second"), 0644); err != nil {
		t.Fatal(err)
	}
	run("add", "test.md")
	run("commit", "-m", "second commit")
}

func TestOpen_FindsGitInParent(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	sub := filepath.Join(dir, "sub", "deep")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	repo, err := Open(filepath.Join(sub, "test.md"))
	if err != nil {
		t.Fatalf("expected to find git, got: %v", err)
	}
	if repo == nil {
		t.Fatal("repo is nil")
	}
}

func TestOpen_NoGitReturnsNil(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "test.md"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	repo, err := Open(filepath.Join(dir, "test.md"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo != nil {
		t.Fatal("expected nil repo when no .git")
	}
}

func TestCommits_AreReverseChronological(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	repo, err := Open(filepath.Join(dir, "test.md"))
	if err != nil {
		t.Fatal(err)
	}
	commits := repo.Commits("test.md")
	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}
	if !strings.Contains(commits[0].Message, "second") {
		t.Errorf("expected first item to be 'second', got %q", commits[0].Message)
	}
}

func TestBlob_ReturnsContent(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	repo, err := Open(filepath.Join(dir, "test.md"))
	if err != nil {
		t.Fatal(err)
	}
	commits := repo.Commits("test.md")
	first := repo.Blob(commits[1].Hash, "test.md")
	if first != "first" {
		t.Errorf("expected 'first', got %q", first)
	}
	second := repo.Blob(commits[0].Hash, "test.md")
	if second != "second" {
		t.Errorf("expected 'second', got %q", second)
	}
}

func TestWorkingTree_ReturnsCurrentContent(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	if err := os.WriteFile(filepath.Join(dir, "test.md"), []byte("unstaged"), 0644); err != nil {
		t.Fatal(err)
	}
	repo, err := Open(filepath.Join(dir, "test.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got := repo.WorkingTree("test.md"); got != "unstaged" {
		t.Errorf("expected 'unstaged', got %q", got)
	}
}
