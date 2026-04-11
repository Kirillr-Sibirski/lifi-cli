import type { Metadata } from "next";
import Link from "next/link";

import "./globals.css";

export const metadata: Metadata = {
  title: "lifi-cli docs",
  description: "Minimal documentation for the lifi CLI and its Earn + Composer flows.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <div className="site-shell">
          <header className="site-header">
            <Link href="/" className="brand">
              <span className="brand-mark">li.fi</span>
              <span className="brand-copy">
                <strong>lifi-cli</strong>
                <small>earn + composer docs</small>
              </span>
            </Link>

            <nav className="site-nav">
              <Link href="/docs/getting-started">Docs</Link>
              <a
                href="https://github.com/Kirillr-Sibirski/lifi-cli"
                target="_blank"
                rel="noreferrer"
              >
                GitHub
              </a>
            </nav>
          </header>

          {children}
        </div>
      </body>
    </html>
  );
}
