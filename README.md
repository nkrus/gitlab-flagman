![logo](logo.png)
# gitlab-flagman - simple GitOps feature flag syncer for Gitlab

[![Build Status](https://github.com/nkrus/gitlab-flagman/actions/workflows/ci.yml/badge.svg)](https://github.com/nkrus/gitlab-flagman/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/nkrus/gitlab-flagman)](https://goreportcard.com/report/github.com/nkrus/gitlab-flagman)
[![GitHub downloads](https://img.shields.io/github/downloads/nkrus/gitlab-flagman/total?label=github%20downloads&style=flat-square)](https://github.com/nkrus/gitlab-flagman/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

gitlab-flagman is a command-line tool for syncing feature flags between managed YAML configuration and [GitLab Feature flags](https://docs.gitlab.com/ee/operations/feature_flags.html). This tool implements the GitOps approach of feature flags management.

## Key Features

- Manage feature flags via the `yaml` file.
- Integrates with GitLab API for automatic updates to feature flags.
- Easy setup and usage.

## Requirements

- [Go](https://golang.org/) version 1.23 or later.
- Access to the GitLab API with a personal access token.
