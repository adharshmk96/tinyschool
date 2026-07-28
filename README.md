# Tiny School

Nuxt UI dashboard backed by a persistent Go, GORM, and SQLite API.

## Run locally

```bash
./run-local.sh
```

Open `http://127.0.0.1:3000`. Logs and PID files are written to
`.runs/local`.

Override occupied ports when needed:

```bash
TINYSCHOOL_API_PORT=8180 TINYSCHOOL_UI_PORT=3100 ./run-local.sh
```

Stop both servers with:

```bash
./stop-local.sh
```

Local data is stored in `.runs/local/tinyschool.db`. A new database is migrated
and seeded automatically. The seeded login is:

```text
alex@tinyschool.local
password
```

## API architecture

```text
Cobra/Viper -> server -> delivery/http -> service -> storage interface
                                                     |
                                                     v
                                            storage/gormsqlite
```

- `internal/model`: database-independent domain models.
- `internal/dto`: request and response contracts.
- `internal/service`: validation and business rules.
- `internal/storage`: replaceable persistence interface.
- `internal/storage/gormsqlite`: all GORM, SQLite, migration, seed, and query
  code.
- `internal/delivery/http`: handlers and HTTP middleware.
- `internal/server`: routes and graceful shutdown.

The detailed endpoint contract and implementation plan is in
`requirements/api-plan.md`.

Run the API directly:

```bash
cd tinyschool-api
go run . --database ./tinyschool.db --address :8080
```

## Deploy (SSH)

The same script deploys from a laptop or GitHub Actions:

```text
upload source → Docker Compose builds → Caddy serves HTTPS → readiness check
```

The server needs Docker with the Compose plugin and SSH key access. Caddy was
chosen over Traefik because this single-app setup only needs one small routing
file and automatic HTTPS.

```bash
ssh-copy-id root@147.93.97.228
./deploy.sh
```

Default URL: `https://tinyschool.147.93.97.228.nip.io`

Override configuration with `DEPLOY_HOST`, `DEPLOY_USER`, `DEPLOY_PATH`, or
`DEPLOY_DOMAIN`. SQLite and Caddy certificates remain in named Docker volumes
between releases.

### GitHub Actions

Add the private key as the `DEPLOY_SSH_PRIVATE_KEY` repository secret.
Optional repository variables use the same `DEPLOY_*` names above.

The workflow deploys on pushes to `main` and can also be run manually.
