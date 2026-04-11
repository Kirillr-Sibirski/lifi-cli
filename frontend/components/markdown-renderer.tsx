import type { ReactNode } from "react";

import Link from "next/link";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

import { rewriteMarkdownHref, slugify } from "@/lib/content";

type MarkdownRendererProps = {
  content: string;
  source: string;
};

export function MarkdownRenderer({ content, source }: MarkdownRendererProps) {
  return (
    <div className="docs-markdown">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          h1: ({ children }) => {
            const text = flattenText(children);
            const id = slugify(text);
            return (
              <h1 id={id}>
                <a href={`#${id}`}>{children}</a>
              </h1>
            );
          },
          h2: ({ children }) => {
            const text = flattenText(children);
            const id = slugify(text);
            return (
              <h2 id={id}>
                <a href={`#${id}`}>{children}</a>
              </h2>
            );
          },
          h3: ({ children }) => {
            const text = flattenText(children);
            const id = slugify(text);
            return (
              <h3 id={id}>
                <a href={`#${id}`}>{children}</a>
              </h3>
            );
          },
          a: ({ href, children }) => {
            const resolved = rewriteMarkdownHref(source, href);
            if (!resolved) {
              return <span>{children}</span>;
            }

            if (resolved.startsWith("http")) {
              return (
                <a href={resolved} target="_blank" rel="noreferrer">
                  {children}
                </a>
              );
            }

            if (resolved.startsWith("#")) {
              return <a href={resolved}>{children}</a>;
            }

            return <Link href={resolved}>{children}</Link>;
          },
          code: ({ className, children }) => {
            const value = String(children).replace(/\n$/, "");
            const isBlock = Boolean(className);
            if (isBlock) {
              return <code className={className}>{value}</code>;
            }

            return <code>{value}</code>;
          },
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}

function flattenText(value: ReactNode): string {
  if (typeof value === "string") {
    return value;
  }

  if (Array.isArray(value)) {
    return value.map(flattenText).join("");
  }

  if (value && typeof value === "object" && "props" in value) {
    const props = value.props as { children?: React.ReactNode };
    return flattenText(props.children ?? "");
  }

  return "";
}
