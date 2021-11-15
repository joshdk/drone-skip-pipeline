[![License][license-badge]][license-link]
[![Actions][github-actions-badge]][github-actions-link]
[![Releases][github-release-badge]][github-release-link]

# DroneCI Skip Pipeline

🤖 DroneCI plugin to skip pipelines based on files changes

## Motivations

This DroneCI plugin enables you skip (or short-circuit) a pipeline based on the files changed as part of the current pull request being built.
You can avoid running a given pipeline if none of the files involved in that pipeline have changed.
This plugin also uses the Github API in order to determine the list of files changes, and as such can be used **without** needing a clone step to be run first.

## Usage

This plugin can be added to your `.drone.yml` as a new step within an existing pipeline. 

```yaml
steps:
- name: debug
  image: ghcr.io/joshdk/drone-skip-pipeline:v0.1.0
  settings:
    rules:
    - package.json
    - app/
```

If your repository is private, a `GITHUB_TOKEN` environment variable must also be configured.

```yaml
steps:
- name: drone-skip-pipeline
  image: ghcr.io/joshdk/drone-skip-pipeline:v0.1.0
  ...
  environment:
    GITHUB_TOKEN:
      from_secret: GITHUB_TOKEN
```

You can then reconfigure any existing clone steps to depend on this new step.

```yaml
- name: clone
  ...
  depends_on:
  - drone-skip-pipeline
```

You must also disable automatic cloning at the pipeline level.

```yaml
clone:
  disable: true
```

## License

This code is distributed under the [MIT License][license-link], see [LICENSE.txt][license-file] for more information.

[github-actions-badge]:  https://github.com/joshdk/drone-skip-pipeline/workflows/Build/badge.svg
[github-actions-link]:   https://github.com/joshdk/drone-skip-pipeline/actions
[github-release-badge]:  https://img.shields.io/github/release/joshdk/drone-skip-pipeline/all.svg
[github-release-link]:   https://github.com/joshdk/drone-skip-pipeline/releases
[license-badge]:         https://img.shields.io/badge/license-MIT-green.svg
[license-file]:          https://github.com/joshdk/drone-skip-pipeline/blob/master/LICENSE.txt
[license-link]:          https://opensource.org/licenses/MIT
