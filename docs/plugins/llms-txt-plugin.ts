import path from "path";
import fs from "fs";
import type { LoadContext, Plugin } from "@docusaurus/types";

export default function llmsTxtPlugin(context: LoadContext): Plugin<any> {
    return {
        name: "llms-txt-plugin",
        loadContent: async () => {
            const { siteDir } = context;
            const contentDir = path.join(siteDir, "docs");
            const allMdx: string[] = [];
            const docsRecords: {
                title: string;
                path: string;
                description: string;
            }[] = [];

            const getMdxFiles = async (
                dir: string,
                relativePath: string = "",
                depth: number = 0
            ) => {
                const entries = await fs.promises.readdir(dir, {
                    withFileTypes: true,
                });

                entries.sort((a, b) => {
                    if (a.isDirectory() && !b.isDirectory()) return -1;
                    if (!a.isDirectory() && b.isDirectory()) return 1;
                    return a.name.localeCompare(b.name);
                });

                for (const entry of entries) {
                    const fullPath = path.join(dir, entry.name);
                    const currentRelativePath = path.join(relativePath, entry.name);

                    if (entry.isDirectory()) {
                        const dirName = entry.name
                            .split("-")
                            .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
                            .join(" ");
                        const headingLevel = "#".repeat(depth + 2);
                        allMdx.push(`\n${headingLevel} ${dirName}\n`);
                        await getMdxFiles(fullPath, currentRelativePath, depth + 1);
                    } else if (entry.name.endsWith(".md")) {
                        const content = await fs.promises.readFile(fullPath, "utf8");
                        let title = entry.name.replace(".md", "");
                        let description = "";

                        const frontmatterMatch = content.match(/^---\n([\s\S]*?)\n---/);
                        if (frontmatterMatch) {
                            const titleMatch = frontmatterMatch[1].match(/title:\s*(.+)/);
                            const descriptionMatch =
                                frontmatterMatch[1].match(/description:\s*(.+)/);

                            if (titleMatch) {
                                title = titleMatch[1].trim();
                            }
                            if (descriptionMatch) {
                                description = descriptionMatch[1].trim();
                            }
                        }

                        const headingLevel = "#".repeat(depth + 3);
                        allMdx.push(`\n${headingLevel} ${title}\n\n${content}`);

                        // Add to docs records for llms.txt
                        docsRecords.push({
                            title,
                            path: currentRelativePath.replace(/\\/g, "/"),
                            description,
                        });
                    }
                }
            };

            await getMdxFiles(contentDir);
            return { allMdx, docsRecords };
        },
        postBuild: async ({ content, outDir }) => {
            const { allMdx, docsRecords } = content as {
                allMdx: string[];
                docsRecords: { title: string; path: string; description: string }[];
            };

            // Write concatenated MDX content
            const concatenatedPath = path.join(outDir, "llms-full.txt");
            await fs.promises.writeFile(concatenatedPath, allMdx.join("\n\n---\n\n"));

            // Create llms.txt with the requested format
            const llmsTxt = `# ${context.siteConfig.title
                }\n\n## Documentation\n\n${docsRecords
                    .map(
                        (doc) =>
                            `- [${doc.title}](${context.siteConfig.url}/${doc.path.replace(
                                ".md",
                                ""
                            )}): ${doc.description}`
                    )
                    .join("\n")}`;
            const llmsTxtPath = path.join(outDir, "llms.txt");
            await fs.promises.writeFile(llmsTxtPath, llmsTxt);
        },
    };
}