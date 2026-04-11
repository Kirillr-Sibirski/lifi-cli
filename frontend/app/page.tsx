import Link from "next/link";

import { getDocs } from "@/lib/content";

const featuredCommands = [
  "lifi doctor --write-checks --chain opt",
  "lifi vaults --chain opt --asset USDC --transactional-only --sort apy --limit 5",
  "lifi quote --vault 0xVaultAddress --from-chain opt --from-token USDC --amount 10",
  "lifi deposit --vault 0xVaultAddress --from-chain opt --from-token USDC --amount 10 --dry-run",
];

export default function HomePage() {
  const docs = getDocs();

  return (
    <main className="landing">
      <section className="hero">
        <div className="hero-copy">
          <p className="eyebrow">LI.FI CLI documentation</p>
          <h1>
            A dark, minimal docs home for the
            <span> Earn + Composer terminal workflow.</span>
          </h1>
          <p className="hero-text">
            <code>lifi</code> helps builders discover vaults, generate Composer
            quotes, run safer deposits, and verify positions without building a
            frontend first.
          </p>

          <div className="hero-actions">
            <Link href="/docs/getting-started" className="button-primary">
              Open docs
            </Link>
            <a
              className="button-secondary"
              href="https://github.com/Kirillr-Sibirski/lifi-cli"
              target="_blank"
              rel="noreferrer"
            >
              View on GitHub
            </a>
          </div>
        </div>

        <div className="hero-panel">
          <div className="terminal-card">
            <div className="terminal-dots">
              <span />
              <span />
              <span />
            </div>
            <pre>
              <code>{`brew tap Kirillr-Sibirski/lifi-cli
brew install lifi

lifi doctor --write-checks --chain opt
lifi vaults --chain opt --asset USDC --limit 5
lifi quote --vault 0xVault --from-chain opt --from-token USDC --amount 10`}</code>
            </pre>
          </div>
        </div>
      </section>

      <section className="feature-grid">
        <article className="feature-card">
          <p className="card-kicker">Earn</p>
          <h2>Discover and inspect vaults</h2>
          <p>
            Browse transactional vaults, rank them, inspect their metadata, and
            verify what is depositable before you touch a wallet.
          </p>
        </article>

        <article className="feature-card">
          <p className="card-kicker">Composer</p>
          <h2>Build executable routes</h2>
          <p>
            Generate route quotes, inspect approvals, export unsigned payloads,
            and track execution status in a single CLI surface.
          </p>
        </article>

        <article className="feature-card">
          <p className="card-kicker">Safety</p>
          <h2>Run preflights before broadcast</h2>
          <p>
            Doctor checks, dry-runs, simulation hints, and gas validation keep
            the write path readable and safer for real funds.
          </p>
        </article>

        <article className="feature-card">
          <p className="card-kicker">Automation</p>
          <h2>Scriptable JSON for builders</h2>
          <p>
            Feed quotes, portfolio reads, and deposit preflights into agents,
            CI, and shell workflows with a consistent output contract.
          </p>
        </article>
      </section>

      <section className="showcase">
        <div className="showcase-copy">
          <p className="eyebrow">What you can do</p>
          <h2>From vault discovery to verified positions.</h2>
          <p>
            The docs are organized around the real product flow so you can start
            at install, move through Earn discovery, and finish on Composer
            execution and ops.
          </p>
        </div>

        <div className="command-list">
          {featuredCommands.map((command) => (
            <code key={command}>{command}</code>
          ))}
        </div>
      </section>

      <section className="docs-preview">
        <div className="section-heading">
          <p className="eyebrow">Useful sections</p>
          <h2>Everything important in one place.</h2>
        </div>

        <div className="docs-grid">
          {docs.map((doc) => (
            <Link key={doc.slug} href={`/docs/${doc.slug}`} className="doc-card">
              <p className="card-kicker">{doc.section}</p>
              <h3>{doc.title}</h3>
              <p>{doc.description}</p>
            </Link>
          ))}
        </div>
      </section>
    </main>
  );
}
