# lifi-cli docs frontend

This folder contains the deployable documentation frontend for `lifi-cli`.

It is a small Next.js app that:

- renders them in a minimal docs-first layout
- sends `/` and `/docs` straight to the Getting Started page
- keeps the authored docs next to the app

## Local development

```bash
cd frontend
bun install
bun run dev
```

Authored docs live in `frontend/content/`.

## Production build

```bash
cd frontend
bun run build
bun run start
```

## Vercel

Set the Vercel project root directory to:

```text
frontend
```

No extra runtime services are needed. The app is static and generated from the
repo markdown at build time.
