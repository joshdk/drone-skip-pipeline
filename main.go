// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.
// SPDX-License-Identifier: MIT

package main

import (
	"errors"
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
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

	_, err := loadConfig()
	if err != nil {
		return err
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
