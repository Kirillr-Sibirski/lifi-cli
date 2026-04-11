import Link from "next/link";

export default function NotFoundPage() {
  return (
    <main className="docs-index">
      <div className="section-heading">
        <p className="eyebrow">Not found</p>
        <h1>This page does not exist.</h1>
        <p>
          The docs content may have moved, or the link might be stale. Start
          from the docs index or jump back to the landing page.
        </p>
      </div>

      <div className="hero-actions">
        <Link href="/docs" className="button-primary">
          Open docs
        </Link>
        <Link href="/" className="button-secondary">
          Go home
        </Link>
      </div>
    </main>
  );
}
