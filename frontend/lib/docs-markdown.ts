import docsConfig from "@/content.config.json";

const REPO_URL = "https://github.com/Kirillr-Sibirski/lifi-cli";

type DocConfig = {
  slug: string;
  title: string;
  description: string;
  section: string;
  source: string;
};

const docs = docsConfig.docs as DocConfig[];

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

  const resolved = normalizePosix(joinPosix(dirnamePosix(currentSource), target));

  const matched = docs.find(
    (doc) => normalizePosix(doc.source.replace(/^\.\.\//, "")) === resolved.replace(/^\.\.\//, ""),
  );

  const suffix = hash ? `#${slugify(hash)}` : "";
  if (matched) {
    return `/docs/${matched.slug}${suffix}`;
  }

  return `${REPO_URL}/blob/main/${resolved}${hash ? `#${hash}` : ""}`;
}

export function slugify(value: string): string {
  return value
    .toLowerCase()
    .trim()
    .replace(/[`*_~]/g, "")
    .replace(/[^\w\s-]/g, "")
    .replace(/\s+/g, "-");
}

function dirnamePosix(value: string): string {
  const normalized = normalizePosix(value);
  const parts = normalized.split("/");
  parts.pop();
  return parts.join("/") || ".";
}

function joinPosix(left: string, right: string): string {
  return `${left}/${right}`;
}

function normalizePosix(value: string): string {
  const input = value.replace(/\\/g, "/");
  const parts = input.split("/");
  const stack: string[] = [];

  for (const part of parts) {
    if (!part || part === ".") {
      continue;
    }
    if (part === "..") {
      stack.pop();
      continue;
    }
    stack.push(part);
  }

  return stack.join("/");
}
