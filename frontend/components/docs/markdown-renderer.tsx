"use client";

import { Children, isValidElement, type ReactNode, useState } from "react";

import Link from "next/link";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

import { rewriteMarkdownHref, slugify } from "@/lib/docs-markdown";

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
          pre: ({ children }) => <CopyablePre>{children}</CopyablePre>,
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

function CopyablePre({ children }: { children: ReactNode }) {
  const [copied, setCopied] = useState(false);
  const child = Children.only(children);

  if (!isValidElement<{ className?: string; children?: ReactNode }>(child)) {
    return <pre>{children}</pre>;
  }

  const className = child.props.className ?? "";
  const language = className.replace("language-", "").trim() || "bash";
  const value = flattenText(child.props.children ?? "");

  async function onCopy() {
    await navigator.clipboard.writeText(value);
    setCopied(true);
    window.setTimeout(() => setCopied(false), 1400);
  }

  return (
    <div className="code-block">
      <div className="code-block-toolbar">
        <span className="code-language">{language}</span>
        <div className="copy-feedback">
          {copied ? "Copied" : ""}
          <button type="button" className="copy-button" onClick={onCopy}>
            Copy
          </button>
        </div>
      </div>
      <pre>{child}</pre>
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
