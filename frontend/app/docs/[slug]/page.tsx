import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

import { DocsSidebar } from "@/components/docs-sidebar";
import { DocsTableOfContents } from "@/components/docs-table-of-contents";
import { MarkdownRenderer } from "@/components/markdown-renderer";
import { getDocBySlugSync, getDocPage, getDocs, getNeighborDocs } from "@/lib/content";

type DocPageProps = {
  params: Promise<{ slug: string }>;
};

export async function generateStaticParams() {
  return getDocs().map((doc) => ({ slug: doc.slug }));
}

export async function generateMetadata({ params }: DocPageProps): Promise<Metadata> {
  const { slug } = await params;
  const doc = getDocBySlugSync(slug);
  if (!doc) {
    return {};
  }

  return {
    title: `${doc.title} | lifi-cli docs`,
    description: doc.description,
  };
}

export default async function DocPage({ params }: DocPageProps) {
  const { slug } = await params;
  const doc = await getDocPage(slug);
  if (!doc) {
    notFound();
  }

  const neighbors = getNeighborDocs(slug);

  return (
    <main className="docs-layout">
      <DocsSidebar activeSlug={slug} />

      <article className="docs-article-shell">
        <div className="docs-meta">
          <div>
            <p className="eyebrow">{doc.section}</p>
            <h1>{doc.title}</h1>
            <p className="docs-description">{doc.description}</p>
          </div>
          <a href={doc.sourceUrl} target="_blank" rel="noreferrer" className="source-link">
            View source on GitHub
          </a>
        </div>

        <MarkdownRenderer content={doc.content} source={doc.source} />

        <div className="docs-pagination">
          {neighbors.previous ? (
            <Link href={`/docs/${neighbors.previous.slug}`} className="pager-card">
              <span>Previous</span>
              <strong>{neighbors.previous.title}</strong>
            </Link>
          ) : (
            <div />
          )}

          {neighbors.next ? (
            <Link href={`/docs/${neighbors.next.slug}`} className="pager-card align-right">
              <span>Next</span>
              <strong>{neighbors.next.title}</strong>
            </Link>
          ) : null}
        </div>
      </article>

      <DocsTableOfContents headings={doc.headings} />
    </main>
  );
}
