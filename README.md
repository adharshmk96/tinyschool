# Tiny School

Nuxt UI dashboard backed by a Go API with realistic in-memory placeholder data.

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
