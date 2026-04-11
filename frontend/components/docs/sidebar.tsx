import Link from "next/link";

import { getDocSections } from "@/lib/content";

type DocsSidebarProps = {
  activeSlug?: string;
};

const quickExamples = [
  "lifi doctor --write-checks --chain opt",
  "lifi vaults --chain opt --asset USDC --transactional-only --limit 5",
  "lifi deposit --vault 0xVault --from-chain opt --from-token USDC --amount 10 --dry-run",
];

export function DocsSidebar({ activeSlug }: DocsSidebarProps) {
  const sections = getDocSections();

  return (
    <aside className="docs-sidebar">
      <div className="sidebar-card">
        <p className="sidebar-kicker">Documentation</p>
        <h2>LI.FI CLI</h2>
        <p>
          Installation, Earn, Composer, config, and operational docs for the{" "}
          <code>lifi</code> CLI.
        </p>
      </div>

      <div className="sidebar-examples">
        <p className="examples-title">Quick examples</p>
        {quickExamples.map((command) => (
          <code key={command} className="example-command">
            {command}
          </code>
        ))}
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
