# Release Instructions

## Creating a New Release

1. **Update CHANGELOG.md**
   - Move items from `[Unreleased]` to a new version section
   - Follow semantic versioning (MAJOR.MINOR.PATCH)

2. **Create and Push Tag**
   ```bash
   # Create annotated tag
   git tag -a v1.0.0 -m "Release v1.0.0"

   # Push tag to trigger release workflow
   git push origin v1.0.0
   ```

3. **GitHub Actions Workflow**
   - Automatically builds multi-arch Docker images (amd64, arm64)
   - Pushes to GitHub Container Registry (ghcr.io)
   - Creates tags: `v1.0.0`, `v1.0`, `v1`, `latest`

4. **Verify Release**
   ```bash
   # Pull and test the image
   docker pull ghcr.io/joluc/cph-metro-exporter:v1.0.0

   # Run with demo mode
   docker run -p 9100:9100 -e DEMO_MODE=true ghcr.io/joluc/cph-metro-exporter:v1.0.0

   # Check metrics
   curl http://localhost:9100/metrics
   ```

## First Release Checklist

- [ ] All tests passing (`go test ./...`)
- [ ] Documentation updated (README.md)
- [ ] CHANGELOG.md updated with changes
- [ ] Dockerfile tested locally
- [ ] GitHub Actions workflows added
- [ ] Repository settings: Enable GitHub Packages

## Semantic Versioning Guide

- **MAJOR** (v2.0.0): Breaking changes (API changes, metric name changes)
- **MINOR** (v1.1.0): New features, backward compatible
- **PATCH** (v1.0.1): Bug fixes, backward compatible

## Rollback

If a release has issues:

```bash
# Delete tag locally
git tag -d v1.0.0

# Delete tag on remote
git push --delete origin v1.0.0

# Delete container image on GitHub
# Go to Packages → cph-metro-exporter → Package settings → Delete version
```
