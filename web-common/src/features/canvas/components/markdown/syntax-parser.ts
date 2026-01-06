/**
 * Parser to extract custom syntax (@measurename and [repeat ...]) from HTML content
 */

export interface MeasureMention {
  type: "measure";
  measureName: string;
  metricsViewName?: string;
  startIndex: number;
  endIndex: number;
}

export interface RepeatBlock {
  type: "repeat";
  measure: string;
  dimension: string;
  orderBy?: string;
  limit?: number;
  where?: string;
  startIndex: number;
  endIndex: number;
  content: string;
}

export type ParsedSyntax = MeasureMention | RepeatBlock;

/**
 * Parse HTML content to extract @measurename mentions and [repeat ...] blocks
 */
export function parseSyntax(html: string): ParsedSyntax[] {
  const results: ParsedSyntax[] = [];

  // Parse @measurename mentions from HTML
  // Look for <span class="measure-mention" data-measure="..." data-metrics-view="...">
  const measureMentionRegex =
    /<span[^>]*class="[^"]*measure-mention[^"]*"[^>]*data-measure="([^"]*)"[^>]*(?:data-metrics-view="([^"]*)")?[^>]*>@([^<]*)<\/span>/g;
  let match;
  while ((match = measureMentionRegex.exec(html)) !== null) {
    results.push({
      type: "measure",
      measureName: match[1] || match[3],
      metricsViewName: match[2],
      startIndex: match.index,
      endIndex: match.index + match[0].length,
    });
  }

  // Parse [repeat ...] blocks from HTML
  // Look for <div data-type="repeat-block" data-measure="..." data-dimension="..." ...>
  const repeatBlockRegex =
    /<div[^>]*data-type="repeat-block"[^>]*data-measure="([^"]*)"[^>]*data-dimension="([^"]*)"[^>]*(?:data-order-by="([^"]*)")?[^>]*(?:data-limit="([^"]*)")?[^>]*(?:data-where="([^"]*)")?[^>]*>([\s\S]*?)<\/div>/g;
  while ((match = repeatBlockRegex.exec(html)) !== null) {
    results.push({
      type: "repeat",
      measure: match[1],
      dimension: match[2],
      orderBy: match[3] || undefined,
      limit: match[4] ? parseInt(match[4], 10) : undefined,
      where: match[5] || undefined,
      content: match[6],
      startIndex: match.index,
      endIndex: match.index + match[0].length,
    });
  }

  // Also parse plain text @measurename patterns (for backward compatibility)
  const plainTextMentionRegex = /@([a-zA-Z_][a-zA-Z0-9_]*)/g;
  while ((match = plainTextMentionRegex.exec(html)) !== null) {
    // Skip if already captured as HTML mention
    const alreadyCaptured = results.some(
      (r) =>
        r.type === "measure" &&
        r.measureName === match[1] &&
        r.startIndex <= match.index &&
        r.endIndex >= match.index + match[0].length,
    );
    if (!alreadyCaptured) {
      results.push({
        type: "measure",
        measureName: match[1],
        startIndex: match.index,
        endIndex: match.index + match[0].length,
      });
    }
  }

  // Sort by start index
  results.sort((a, b) => a.startIndex - b.startIndex);

  return results;
}

/**
 * Extract text content from HTML (for plain text parsing)
 */
export function extractTextFromHTML(html: string): string {
  // Simple HTML to text conversion
  return html
    .replace(/<[^>]+>/g, "")
    .replace(/&nbsp;/g, " ")
    .replace(/&amp;/g, "&")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'");
}

