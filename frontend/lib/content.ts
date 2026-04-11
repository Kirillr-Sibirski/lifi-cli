import { promises as fs } from "node:fs";
import path from "node:path";

import docsConfig from "@/content.config.json";

export type DocConfig = {
  slug: string;
  title: string;
  description: string;
  section: string;
  source: string;
};

export type DocHeading = {
  id: string;
  text: string;
  level: number;
};

export type DocPage = DocConfig & {
  content: string;
  headings: DocHeading[];
  sourceUrl: string;
};

const REPO_URL = "https://github.com/Kirillr-Sibirski/lifi-cli";
const contentDirectory = path.join(process.cwd(), "content");
const docs = docsConfig.docs as DocConfig[];

export function getDocs(): DocConfig[] {
  return docs;
}

export function getDocSections(): Array<{ section: string; docs: DocConfig[] }> {
  const groups = new Map<string, DocConfig[]>();
  for (const doc of docs) {
    const existing = groups.get(doc.section) ?? [];
    existing.push(doc);
    groups.set(doc.section, existing);
  }

  return Array.from(groups.entries()).map(([section, sectionDocs]) => ({
    section,
    docs: sectionDocs,
  }));
}

export function getDocBySlugSync(slug: string): DocConfig | undefined {
  return docs.find((doc) => doc.slug === slug);
}

export async function getDocPage(slug: string): Promise<DocPage | null> {
  const doc = getDocBySlugSync(slug);
  if (!doc) {
    return null;
  }

  const filePath = path.join(contentDirectory, `${slug}.md`);
  const content = await fs.readFile(filePath, "utf8");

  return {
    ...doc,
    content,
    headings: extractHeadings(content),
    sourceUrl: `${REPO_URL}/blob/main/${doc.source.replace(/^\.\.\//, "")}`,
  };
}

export function getNeighborDocs(slug: string): {
  previous: DocConfig | null;
  next: DocConfig | null;
} {
  const index = docs.findIndex((doc) => doc.slug === slug);
  if (index === -1) {
    return { previous: null, next: null };
  }

  return {
    previous: docs[index - 1] ?? null,
    next: docs[index + 1] ?? null,
  };
}

export function rewriteMarkdownHref(currentSource: string, href?: string): string | undefined {
  if (!href) {
    return href;
  }

  if (
    href.startsWith("http://") ||
    href.startsWith("https://") ||
    href.startsWith("mailto:") ||
    href.startsWith("#")
  ) {
    return href;
  }

  const [target, hash] = href.split("#");
  if (!target.endsWith(".md")) {
    return href;
  }

  const resolved = path.posix.normalize(
    path.posix.join(path.posix.dirname(currentSource), target),
  );

  const matched = docs.find(
    (doc) => path.posix.normalize(doc.source.replace(/^\.\.\//, "")) === resolved.replace(/^\.\.\//, ""),
  );

  const suffix = hash ? `#${slugify(hash)}` : "";
  if (matched) {
    return `/docs/${matched.slug}${suffix}`;
  }

  return `${REPO_URL}/blob/main/${resolved}${hash ? `#${hash}` : ""}`;
}

function extractHeadings(markdown: string): DocHeading[] {
  const headings: DocHeading[] = [];
  const pattern = /^(#{1,3})\s+(.+)$/gm;
  let match: RegExpExecArray | null = pattern.exec(markdown);

  while (match) {
    const hashes = match[1] ?? "";
    const text = cleanHeading(match[2] ?? "");
    headings.push({
      id: slugify(text),
      text,
      level: hashes.length,
    });
    match = pattern.exec(markdown);
  }

  return headings;
}

function cleanHeading(value: string): string {
  return value
    .replace(/`/g, "")
    .replace(/\[(.*?)\]\((.*?)\)/g, "$1")
    .trim();
}

export function slugify(value: string): string {
  return value
    .toLowerCase()
    .trim()
    .replace(/[`*_~]/g, "")
    .replace(/[^\w\s-]/g, "")
    .replace(/\s+/g, "-");
}
