import path from "path";
import fs from "fs";
import type { LoadContext, Plugin } from "@docusaurus/types";

type DocRecord = {
    title: string;
    permalink: string;
    description: string;
    source: string;
};

// Writes llms.txt (an index of every doc) and llms-full.txt (every doc's Markdown) to the build output.
// Docs come from the docs plugin's loaded content, so URLs, titles and descriptions match the built site.
export default function llmsTxtPlugin(context: LoadContext): Plugin<void> {
    let docs: DocRecord[] = [];

    return {
        name: "llms-txt-plugin",
        async allContentLoaded({ allContent }) {
            const content: any =
                allContent["docusaurus-plugin-content-docs"]?.["default"];
            docs = content.loadedVersions[0].docs
                .filter((d: any) => !d.draft && !d.unlisted)
                .map((d: any) => ({
                    title: d.title,
                    permalink: d.permalink,
                    description: d.description ?? "",
                    source: d.source,
                }))
                .sort((a: DocRecord, b: DocRecord) =>
                    a.permalink.localeCompare(b.permalink)
                );
        },
        async postBuild({ outDir }) {
            const site = context.siteConfig.url;

            const list = docs.map(
                (d) =>
                    `- [${d.title}](${site}${d.permalink})${d.description ? `: ${d.description}` : ""}`
            );
            await fs.promises.writeFile(
                path.join(outDir, "llms.txt"),
                `# ${context.siteConfig.title}\n\n## Documentation\n\n${list.join("\n")}\n`
            );

            const pages = await Promise.all(
                docs.map(async (d) => {
                    const raw = await fs.promises.readFile(
                        d.source.replace(/^@site\//, `${context.siteDir}/`),
                        "utf8"
                    );
                    // Drop front matter and MDX component imports; keep imports inside code samples.
                    const body = raw
                        .replace(/^---\n[\s\S]*?\n---\n/, "")
                        .replace(/^import .* from ['"]@(site|theme|docusaurus)\/.*$/gm, "")
                        .trim();
                    return `# ${d.title}\n\nURL: ${site}${d.permalink}\n\n${body}`;
                })
            );
            await fs.promises.writeFile(
                path.join(outDir, "llms-full.txt"),
                pages.join("\n\n---\n\n")
            );
        },
    };
}
