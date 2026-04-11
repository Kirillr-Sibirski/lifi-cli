import Link from "next/link";

import { getDocSections } from "@/lib/content";

type DocsSidebarProps = {
  activeSlug?: string;
};

export function DocsSidebar({ activeSlug }: DocsSidebarProps) {
  const sections = getDocSections();

  return (
    <aside className="docs-sidebar">
      <div className="sidebar-card">
        <p className="sidebar-kicker">Documentation</p>
        <h2>LI.FI CLI</h2>
        <p>
          A focused docs surface for the <code>lifi</code> command line tool,
          built around Earn and Composer flows.
        </p>
      </div>

      <nav className="sidebar-nav" aria-label="Documentation">
        {sections.map((section) => (
          <div key={section.section} className="sidebar-section">
            <p className="sidebar-section-title">{section.section}</p>
            <ul>
              {section.docs.map((doc) => {
                const active = doc.slug === activeSlug;
                return (
                  <li key={doc.slug}>
                    <Link
                      className={active ? "sidebar-link active" : "sidebar-link"}
                      href={`/docs/${doc.slug}`}
                    >
                      <span>{doc.title}</span>
                      <small>{doc.description}</small>
                    </Link>
                  </li>
                );
              })}
            </ul>
          </div>
        ))}
      </nav>
    </aside>
  );
}
