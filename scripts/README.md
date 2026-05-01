# Scripts

## sync-to-ce.sh

Synchronizes a tagged release from the private repo to KubeCoDriver-ce (community edition) using rsync.

### Usage

```bash
# Sync specific tag (CE repo auto-detected as sibling directory)
./scripts/sync-to-ce.sh v1.1.0

# Sync to custom CE repo path
./scripts/sync-to-ce.sh v1.1.0 /path/to/KubeCoDriver-ce
```

### What it does

1. Extracts the specified git tag to a temporary directory
2. Rsyncs to the CE repo with `--delete` (removes files not in the tag)
3. Excludes `.git/`, `.kiro*`, and `.agents/`
4. Commits and pushes changes to the CE repo

### Workflow

```bash
git tag v1.1.0
git push origin v1.1.0
./scripts/sync-to-ce.sh v1.1.0
```

## update-deps.sh

Updates Go dependencies via Docker (no local Go installation required).

### Usage

```bash
./scripts/update-deps.sh
# or via Makefile:
make mod-tidy
```

### What it does

1. Authenticates to ECR Public (for the Go base image)
2. Runs `go mod tidy` inside a Docker container matching the project's Go version
3. Outputs updated `go.mod` and `go.sum` for review and commit
