import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";

const root = process.cwd();
const configPath = path.join(root, "content.config.json");
const outputDir = path.join(root, "content");

const config = JSON.parse(await readFile(configPath, "utf8"));

await mkdir(outputDir, { recursive: true });

for (const doc of config.docs) {
  const sourcePath = path.resolve(root, doc.source);
  const targetPath = path.join(outputDir, `${doc.slug}.md`);
  const markdown = await readFile(sourcePath, "utf8");
  await writeFile(targetPath, markdown, "utf8");
}

console.log(`synced ${config.docs.length} markdown files into frontend/content`);
