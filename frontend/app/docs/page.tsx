import Link from "next/link";

import { getDocSections } from "@/lib/content";

export default function DocsIndexPage() {
  const sections = getDocSections();

  return (
    <main className="docs-index">
      <div className="section-heading">
        <p className="eyebrow">Documentation</p>
        <h1>Browse the lifi-cli knowledge base.</h1>
        <p>
          Start with installation and the first Earn flow, then move deeper into
          Composer, configuration, automation, and release operations.
        </p>
      </div>

      {sections.map((section) => (
        <section key={section.section} className="docs-index-section">
          <h2>{section.section}</h2>
          <div className="docs-grid">
            {section.docs.map((doc) => (
              <Link key={doc.slug} href={`/docs/${doc.slug}`} className="doc-card">
                <p className="card-kicker">{section.section}</p>
                <h3>{doc.title}</h3>
                <p>{doc.description}</p>
              </Link>
            ))}
          </div>
        </section>
      ))}
    </main>
  );
}
