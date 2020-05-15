package utility

import (
	"testing"

	git "github.com/libgit2/git2go/v30"
)

var DjangoLikeQueryset = "testdata/django-like-queryset/.git"

func TestUpdate(t *testing.T) {
	type args struct {
		repoPath string
		cfg      map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"rename by head offset",
			args{
				DjangoLikeQueryset,
				map[string]string{
					"HEAD~1": "renamed HEAD~1",
					"HEAD~3": "renamed HEAD~3",
					"HEAD~7": "renamed HEAD~7",
				}},
			false,
		},
		{
			"inexistent commit",
			args{
				DjangoLikeQueryset,
				map[string]string{
					"kek":    "renamed kek",
					"HEAD~1": "renamed HEAD~1",
					"HEAD~3": "renamed HEAD~3",
				}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newHashes, err := Update(tt.args.repoPath, tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ERROR: %v", err)
				return
			}
			if tt.wantErr {
				return
			}

			repo, _ := git.OpenRepository(tt.args.repoPath)

			for commit, wantMsg := range tt.args.cfg {
				wantMsg += "\n"
				newCommit, err := getCommit(newHashes[commit], repo)
				if err != nil {
					t.Errorf("ERROR for commit %s: %v", commit, err)
				}
				if newCommit.Message() != wantMsg {
					t.Errorf("ERROR for commit %s: got '%s', want '%s'", commit, newCommit.Message(), wantMsg)
				}
			}
		})
	}
}
