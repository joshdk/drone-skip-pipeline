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

	"github.com/google/go-github/v51/github"
	"github.com/kelseyhightower/envconfig"
	ignore "github.com/sabhiram/go-gitignore"
	"golang.org/x/oauth2"
	"jdk.sh/meta"
)

// exitCodeDroneSkipPipeline is a special exit code that can be returned by a
// step to indicate that the rest of the pipeline should be skipped.
// See https://github.com/drone/drone-runtime/issues/51.
// See https://discourse.drone.io/t/how-to-exit-a-pipeline-early-without-failing/3951.
const exitCodeDroneSkipPipeline = 78

// errDroneSkipPipeline is a sentinel error indicating that the rest of the
// current pipeline should be skipped.
var errDroneSkipPipeline = errors.New("skipping pipeline")

func main() {
	switch err := mainCmd(); err { // nolint:errorlint
	case nil:
		log.Println("continuing pipeline")

		return
	case errDroneSkipPipeline:
		log.Println("skipping pipeline")
		os.Exit(exitCodeDroneSkipPipeline)
	default:
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
				&oauth2.Token{AccessToken: cfg.GithubToken}, // nolint:exhaustivestruct
			),
		)
	}

	client := github.NewClient(httpClient)

	opt := &github.ListOptions{PerPage: 100}
	var commitFiles []*github.CommitFile
	for {
		// Get a list of all files (added, deleted, modified) that are a part of
		// the current pull request.
		files, resp, err := client.PullRequests.ListFiles(ctx, cfg.RepoOwner, cfg.RepoName, cfg.PullRequest, opt)
		if err != nil {
			return err
		}

		commitFiles = append(commitFiles, files...)

		// Bail out if there are no more paginated results.
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	matcher := ignore.CompileIgnoreLines(cfg.Rules...)

	skip := true
	// Examine every file in the current pull request, and try to match it
	// against the set of configured plugin rules.
	for _, commitFile := range commitFiles {
		filename := commitFile.GetFilename()
		if matched, how := matcher.MatchesPathHow(filename); matched { // nolint:gocritic
			skip = false
			// File was matched by a rule.
			log.Printf("file %s matched by rule %q\n", filename, how.Line)
		} else if how != nil {
			// File was matched by a rule, but then negated by another.
			log.Printf("file %s not matched by negated rule %q\n", filename, how.Line)
		} else {
			// File was not matched by any rules.
			log.Printf("file %s not matched by any rule\n", filename)
		}
	}

	// No files were matched by any of the plugin rules. Skip the rest of the
	// pipeline.
	if skip {
		// Touch a sentinel file to signal to subsequent steps that the
		// pipeline should be skipped.
		if cfg.Touch != "" {
			log.Printf("touching file %s", cfg.Touch)
			if err := touchFile(cfg.Touch); err != nil {
				return err
			}
		}

		return errDroneSkipPipeline
	}

	// At least one file matched the plugin rules, and the pipeline should be
	// continued as normal.
	return nil
}

type config struct {
	Event       string   `envconfig:"DRONE_BUILD_EVENT"`
	GithubToken string   `envconfig:"GITHUB_TOKEN"`
	PullRequest int      `envconfig:"DRONE_PULL_REQUEST"`
	RepoName    string   `envconfig:"DRONE_REPO_NAME"`
	RepoOwner   string   `envconfig:"DRONE_REPO_OWNER"`
	Rules       []string `envconfig:"PLUGIN_RULES"`
	Touch       string   `envconfig:"PLUGIN_TOUCH"`
}

func loadConfig() (*config, error) {
	var cfg config
	// Load plugin configuration from current working environment.
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

func touchFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	return file.Close()
}
