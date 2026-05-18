# SimpleTrack Layered Docker Compose

## Compose Layers

- `docker-compose.yml`
  - Base infrastructure shared by the full local stack.
- `docker-compose.apps.yml`
  - Common application service definitions without local build details.
- `docker-compose.dev.yml`
  - Local development and acceptance overrides that build images from the checked-out source tree.
- `docker-compose.release.yml`
  - Release overrides that consume prebuilt images via `IMAGE_REGISTRY`, `IMAGE_NAMESPACE`, and `IMAGE_TAG`.

## Local Development

Copy `.env.dev.example` to `.env.dev` before the first startup when you need host-port or proxy overrides.

Preferred helper scripts:

- Windows PowerShell: `powershell -ExecutionPolicy Bypass -File src/simpletrack-saas/scripts/ensure_simpletrack_dev_stack.ps1`
- Linux/macOS shell: `bash src/simpletrack-saas/scripts/ensure_simpletrack_dev_stack.sh`

Both helpers start infra first, then application services, and print the public entry URLs when the stack is ready.

Manual compose equivalent:

```bash
docker compose \
  --env-file src/deploy/local/docker/.env.dev \
  -f src/deploy/local/docker/docker-compose.yml \
  -f src/deploy/local/docker/docker-compose.apps.yml \
  -f src/deploy/local/docker/docker-compose.dev.yml \
  up -d --build
```

## Public Entry URLs

- Marketing: `http://localhost:3004`
- Docs: `http://localhost:3002`
- Quickstart: `http://localhost:3002/quickstart`
- SaaS: `http://localhost:3005/login`
- analytics-service health: `http://localhost:8080/healthz`

## Stop

- Windows PowerShell: `powershell -ExecutionPolicy Bypass -File src/simpletrack-saas/scripts/stop_simpletrack_dev_stack.ps1`
- Linux/macOS shell: `bash src/simpletrack-saas/scripts/stop_simpletrack_dev_stack.sh`
