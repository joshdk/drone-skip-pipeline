// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/google/go-github/v40/github"
	"github.com/kelseyhightower/envconfig"
	ignore "github.com/sabhiram/go-gitignore"
	"golang.org/x/oauth2"
	"jdk.sh/meta"
)

func main() {
	if err := mainCmd(); err != nil {
		log.Println("joshdk/drone-skip-pipeline:", err)
		os.Exit(1)
	}
}

func mainCmd() error {
	log.Printf("joshdk/drone-skip-pipeline %s (%s)\n", meta.Version(), meta.ShortSHA())

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Create a new Github API client, with an optional access token.
	var httpClient *http.Client
	if cfg.GithubToken != "" {
		httpClient = oauth2.NewClient(ctx,
			oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: cfg.GithubToken},
			),
		)
	}
	client := github.NewClient(httpClient)

	// Get a list of all files (added, deleted, modified) that are a part of
	// the current pull request.
	commitFiles, _, err := client.PullRequests.ListFiles(ctx, cfg.RepoOwner, cfg.RepoName, cfg.PullRequest, nil)
	if err != nil {
		return err
	}

	matcher := ignore.CompileIgnoreLines(cfg.Rules...)

	// Examine every file in the current pull request, and try to match it
	// against the set of configured plugin rules.
	for _, commitFile := range commitFiles {
		filename := commitFile.GetFilename()
		if matched, how := matcher.MatchesPathHow(filename); matched {
			// File was matched by a rule.
			log.Printf("%s matched by rule %q\n", filename, how.Line)
		} else if how != nil {
			// File was matched by a rule, but then negated by another.
			log.Printf("%s not matched by negated rule %q\n", filename, how.Line)
		} else {
			// File was not matched by any rules.
			log.Printf("%s not matched by any rule\n", filename)
		}
	}

	return nil
}

type config struct {
	Event       string   `envconfig:"DRONE_BUILD_EVENT"`
	GithubToken string   `envconfig:"GITHUB_TOKEN"`
	PullRequest int      `envconfig:"DRONE_PULL_REQUEST"`
	RepoName    string   `envconfig:"DRONE_REPO_NAME"`
	RepoOwner   string   `envconfig:"DRONE_REPO_OWNER"`
	Rules       []string `envconfig:"PLUGIN_RULES"`
}

func loadConfig() (*config, error) {
	// Load plugin configuration from current working environment.
	var cfg config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}

	// Sanity check that plugin is sufficiently configured.
	switch {
	case cfg.Event == "":
		return nil, errors.New("missing DRONE_BUILD_EVENT")
	case cfg.PullRequest == 0:
		return nil, errors.New("missing DRONE_PULL_REQUEST")
	case cfg.RepoName == "":
		return nil, errors.New("missing DRONE_REPO_NAME")
	case cfg.RepoOwner == "":
		return nil, errors.New("missing DRONE_REPO_OWNER")
	}

	return &cfg, nil
}
