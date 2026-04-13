import { promises as fs } from "node:fs";
import path from "node:path";

import docsConfig from "@/content.config.json";
import { slugify } from "@/lib/docs-markdown";

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

  const filePath = path.resolve(process.cwd(), "..", doc.source);
  const content = await fs.readFile(filePath, "utf8");

  return {
    ...doc,
    content,
    headings: extractHeadings(content),
    sourceUrl: `${REPO_URL}/blob/main/${doc.source}`,
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
