// Builds a download filename from a dashboard title plus a local timestamp,
// e.g. "sales-overview-20260619-130412.pdf". Shared by the canvas and explore
// PDF export orchestrators.
export function buildPdfFilename(title: string): string {
  const slug =
    title
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-+|-+$/g, "") || "dashboard";
  const now = new Date();
  const pad = (n: number) => String(n).padStart(2, "0");
  const stamp =
    `${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(now.getDate())}` +
    `-${pad(now.getHours())}${pad(now.getMinutes())}${pad(now.getSeconds())}`;
  return `${slug}-${stamp}.pdf`;
}
