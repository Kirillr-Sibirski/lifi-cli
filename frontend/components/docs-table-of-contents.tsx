import { DocHeading } from "@/lib/content";

type DocsTableOfContentsProps = {
  headings: DocHeading[];
};

export function DocsTableOfContents({ headings }: DocsTableOfContentsProps) {
  const visible = headings.filter((heading) => heading.level <= 3);
  if (visible.length === 0) {
    return null;
  }

  return (
    <aside className="toc">
      <p className="toc-title">On this page</p>
      <ul>
        {visible.map((heading) => (
          <li key={heading.id} className={`level-${heading.level}`}>
            <a href={`#${heading.id}`}>{heading.text}</a>
          </li>
        ))}
      </ul>
    </aside>
  );
}
