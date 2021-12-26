package gui

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func isInstalled(pr string) bool {
	paths := strings.Split(os.Getenv("PATH"), ":")

	for _, p := range paths {
		path := filepath.Join(p, pr)
		_, err := os.Lstat(path)
		if err == nil {
			return true
		}
	}
	return false
}

type releaseFinder struct {
	api    *github.Client
	apiCtx context.Context
}

func newReleaseFinder(ctx context.Context, token string) *releaseFinder {
	hc := newHTTPClient(ctx, token)
	cli := github.NewClient(hc)

	return &releaseFinder{
		api:    cli,
		apiCtx: ctx,
	}
}

func binaryVersion(v string) string {
	return filepath.Join(stateDir(), "bin", fmt.Sprintf("edgevpn-%s", v))
}

func listDir(dir string) ([]string, error) {
	content := []string{}

	err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			content = append(content, path)

			return nil
		})

	return content, err
}

func availableVersions() (versions []string) {
	files, _ := listDir(filepath.Join(stateDir(), "bin"))
	for _, f := range files {
		v := strings.ReplaceAll(filepath.Base(f), "edgevpn-", "")
		if strings.HasPrefix(v, "v") {
			versions = append(versions, v)
		}
	}
	return
}

func newHTTPClient(ctx context.Context, token string) *http.Client {
	if token == "" {
		return http.DefaultClient
	}
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	return oauth2.NewClient(ctx, src)
}

func (f *releaseFinder) findAll(slug string) ([]string, error) {
	repo := strings.Split(slug, "/")
	if len(repo) != 2 || repo[0] == "" || repo[1] == "" {
		return nil, fmt.Errorf("Invalid slug format. It should be 'owner/name': %s", slug)
	}

	rels, res, err := f.api.Repositories.ListReleases(f.apiCtx, repo[0], repo[1], nil)
	if err != nil {
		log.Println("API returned an error response:", err)
		if res != nil && res.StatusCode == 404 {
			// 404 means repository not found or release not found. It's not an error here.
			err = nil
			log.Println("API returned 404. Repository or release not found")
		}
		return nil, err
	}

	versions := []string{}
	for _, rel := range rels {
		versions = append(versions, *rel.Name)

	}
	return versions, nil
}

func (f *releaseFinder) find(slug string, version string) (*github.RepositoryRelease, *github.ReleaseAsset, error) {
	repo := strings.Split(slug, "/")
	if len(repo) != 2 || repo[0] == "" || repo[1] == "" {
		return nil, nil, fmt.Errorf("Invalid slug format. It should be 'owner/name': %s", slug)
	}

	rels, res, err := f.api.Repositories.ListReleases(f.apiCtx, repo[0], repo[1], nil)
	if err != nil {
		log.Println("API returned an error response:", err)
		if res != nil && res.StatusCode == 404 {
			// 404 means repository not found or release not found. It's not an error here.
			err = nil
			log.Println("API returned 404. Repository or release not found")
		}
		return nil, nil, err
	}

	if version == "" {
		a := findAsset(rels[0].Assets)
		if a == nil {
			return nil, nil, fmt.Errorf("cannot find asset for '%s' '%s'", slug, version)
		}

		return rels[0], a, nil
	}
	for _, rel := range rels {
		fmt.Println("=====")
		fmt.Println(*rel.Name)
		if *rel.Name == version {
			a := findAsset(rel.Assets)
			if a == nil {
				return nil, nil, fmt.Errorf("cannot find asset for '%s' '%s'", slug, version)
			}

			return rel, a, nil
		}

	}
	//	url := asset.GetBrowserDownloadURL()
	//	log.Println("Successfully fetched the latest release. tag:", rel.GetTagName(), ", name:", rel.GetName(), ", URL:", rel.GetURL(), ", Asset:", url)
	return nil, nil, fmt.Errorf("No good release found for '%s' '%s'", slug, version)
}

func findAsset(ass []github.ReleaseAsset) *github.ReleaseAsset {
	for _, a := range ass {
		if strings.Contains(*a.Name, "Linux-x86_64") {
			return &a
		}
	}
	return nil
}
