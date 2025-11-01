CI / CD for Cookpedia-backend

This repo includes two GitHub Actions workflows:

- `.github/workflows/ci.yml` — runs on push and PR to `main` and `develop`. It performs `go vet`, `go test`, and `golangci-lint`.
- `.github/workflows/cd.yml` — runs on push to `main`. It builds a Docker image and pushes it to GitHub Container Registry (GHCR).

How to enable CD to GHCR
1. By default the workflow uses the repository `GITHUB_TOKEN` to authenticate with GHCR. No further secrets are required for pushing to GHCR from the same repository (the default token is fine), but you must ensure the repository owner allows packages to be written by Actions.
2. If you prefer Docker Hub, edit `.github/workflows/cd.yml` and replace the `docker/login-action` parameters and add `DOCKER_USERNAME` and `DOCKER_PASSWORD` secrets in the repository settings.

Notes and troubleshooting
- If the CI or CD workflows fail because of missing tools (e.g. network or permission), review the action logs from the GitHub Actions tab.
- The Dockerfile is a multi-stage build; the CD workflow builds and pushes amd64 images. Adjust `platforms` in `.github/workflows/cd.yml` if you need other architectures.

Local testing tips
- To run the same checks locally:
  - go vet ./...
  - go test ./...
  - install golangci-lint and run `golangci-lint run`
