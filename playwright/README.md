# Playwright screenshots

Captures every public, dashboard, detail-tab, settings and admin page at a
1440×1000 viewport. API responses are fulfilled by Playwright; no API server or
database is needed.

```bash
./scripts/take-screenshot.sh
```

On the first run, install Chromium once:

```bash
cd playwright
bunx playwright install chromium
```

Screenshots are written to `playwright/output/light/` and
`playwright/output/dark/`, which are git ignored.
