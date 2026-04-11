# lifi-cli docs frontend

This folder contains the deployable documentation frontend for `lifi-cli`.

It is a small Next.js app that:

- syncs the markdown files from the repository root
- renders them in a dark GitBook-style layout
- provides a landing page for the CLI and its Earn + Composer flows

## Local development

```bash
cd frontend
bun install
bun run dev
```

The content is synced automatically from:

- `../README.md`
- `../docs/*.md`
- `../CHANGELOG.md`

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
