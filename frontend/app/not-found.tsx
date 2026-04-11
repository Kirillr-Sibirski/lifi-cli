import Link from "next/link";

export default function NotFoundPage() {
  return (
    <main className="docs-layout">
      <section className="docs-article-shell">
        <p className="eyebrow">Not found</p>
        <h1>This page does not exist.</h1>
        <p className="docs-description">
          The docs content may have moved, or the link might be stale. Start
          from the docs entry point or jump back to the repository.
        </p>
        <div style={{ display: "flex", gap: "0.75rem", marginTop: "1.25rem" }}>
          <Link href="/docs/getting-started" className="source-link">
            Open docs
          </Link>
          <Link href="/" className="source-link">
            Go home
          </Link>
        </div>
      </section>
    </main>
  );
}
