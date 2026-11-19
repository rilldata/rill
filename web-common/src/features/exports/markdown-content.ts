import DOMPurify from "dompurify";
import { marked } from "marked";

// A table in a markdown text: a grid of the text its cells show.
export interface MarkdownTable {
  // The table's caption, or else the heading it sits under; empty when it has neither.
  title: string;
  // The column header rows followed by the body rows, every row as wide as the table.
  rows: string[][];
  // How many of the rows are column headers.
  headerRows: number;
}

export type MarkdownPart =
  | { kind: "text"; text: string }
  | { kind: "table"; table: MarkdownTable };

// What a markdown text shows, in reading order: runs of plain text and tables.
// The markdown is rendered and sanitized the way the canvas markdown component shows it,
// so content written in markdown and in HTML both count, and what the sanitizer removes is not exported.
// Text keeps its lines with the formatting dropped; a text that shows nothing has no parts.
export function markdownParts(markdown: string): MarkdownPart[] {
  const html = DOMPurify.sanitize(marked.parse(markdown, { async: false }));
  // An inert document: nothing in it runs or loads.
  const document = new DOMParser().parseFromString(html, "text/html");

  const parts: MarkdownPart[] = [];
  const flow = new TextFlow();
  const endText = () => {
    const text = flow.take();
    if (text) parts.push({ kind: "text", text });
  };
  const addTable = (element: HTMLTableElement, inherited: string) => {
    const title = oneLine(element.caption) || inherited;
    const table = tableGrid(element, title);
    if (table) parts.push({ kind: "table", table });
    // A table inside a cell is also a table of its own, after the one that holds it.
    for (const nested of element.querySelectorAll("table")) {
      if (nested.parentElement?.closest("table") === element) {
        addTable(nested, title);
      }
    }
  };
  let heading = "";
  writeText(document.body, flow, (element) => {
    if (/^H[1-6]$/.test(element.tagName)) heading = oneLine(element);
    if (element.tagName !== "TABLE") return false;
    // A caption is shown with its table but is not one of its cells, so it reads as text above the table.
    flow.startLine(2);
    flow.add(oneLine((element as HTMLTableElement).caption), false);
    endText();
    addTable(element as HTMLTableElement, heading);
    return true;
  });
  endText();
  return parts;
}

// Elements whose content is not shown as text.
const UNSHOWN = new Set(["STYLE", "SCRIPT", "TEMPLATE", "TITLE"]);

// Elements that start on a line of their own.
const BLOCKS = new Set([
  "ADDRESS",
  "ARTICLE",
  "ASIDE",
  "DD",
  "DETAILS",
  "DIV",
  "DL",
  "DT",
  "FIELDSET",
  "FIGCAPTION",
  "FIGURE",
  "FOOTER",
  "FORM",
  "H1",
  "H2",
  "H3",
  "H4",
  "H5",
  "H6",
  "HEADER",
  "LI",
  "MAIN",
  "NAV",
  "SECTION",
  "SUMMARY",
  "TR",
]);

// Blocks the markdown component sets apart with a margin; they get a blank line around them.
const PARAGRAPHS = new Set(["BLOCKQUOTE", "HR", "OL", "P", "PRE", "UL"]);

// Columns one cell may span: far more than a table shows, and a bound on what a stray attribute can allocate.
const MAX_COL_SPAN = 100;

// Writes the text a node's content shows to the flow, in document order.
// `takeOver` is asked about every element before its content, and the content of one it takes over is skipped.
function writeText(
  node: Node,
  flow: TextFlow,
  takeOver: (element: Element) => boolean,
  preformatted = false,
): void {
  for (const child of node.childNodes) {
    if (child.nodeType === child.TEXT_NODE) {
      flow.add(child.textContent ?? "", preformatted);
      continue;
    }
    if (child.nodeType !== child.ELEMENT_NODE) continue;
    const element = child as Element;
    const tag = element.tagName;
    if (UNSHOWN.has(tag) || isHidden(element) || takeOver(element)) continue;

    if (tag === "BR") {
      flow.lineBreak();
      continue;
    }
    if (tag === "IMG") {
      flow.add(element.getAttribute("alt") ?? "", false);
      continue;
    }
    // Inside a list item, a paragraph or a nested list continues the item rather than standing apart.
    let lines = 0;
    if (PARAGRAPHS.has(tag) && !element.parentElement?.closest("li")) {
      lines = 2;
    } else if (PARAGRAPHS.has(tag) || BLOCKS.has(tag)) {
      lines = 1;
    }
    flow.startLine(lines);
    if (tag === "LI") flow.mark(listMarker(element));
    writeText(element, flow, takeOver, preformatted || tag === "PRE");
    if (tag === "LI") flow.mark("");
    // The cells of a row read as separate words.
    if (tag === "TD" || tag === "TH") flow.add(" ", false);
    flow.startLine(lines);
  }
}

// Whether an element is hidden by its own markup. Stylesheets are not applied to the document.
function isHidden(element: Element): boolean {
  return (
    element.hasAttribute("hidden") ||
    /(^|;)\s*display\s*:\s*none/i.test(element.getAttribute("style") ?? "")
  );
}

// The bullet or number of a list item, indented by the depth of its list.
function listMarker(item: Element): string {
  const list = item.parentElement;
  let indent = "";
  for (let a = list?.parentElement; a; a = a.parentElement) {
    if (a.tagName === "UL" || a.tagName === "OL") indent += "  ";
  }
  if (list?.tagName !== "OL") return `${indent}• `;
  const start = Number(list.getAttribute("start")) || 1;
  const index = Array.from(list.children)
    .filter((child) => child.tagName === "LI")
    .indexOf(item);
  return `${indent}${start + index}. `;
}

// The text an element shows, on one line.
function oneLine(element: Element | null): string {
  if (!element) return "";
  const flow = new TextFlow();
  writeText(element, flow, () => false);
  return flow.take().replace(/\s+/g, " ");
}

// Plain text built from inline text and line breaks, with white space collapsed the way a browser lays it out.
class TextFlow {
  private text = "";
  // Line breaks owed before the next text.
  private breaks = 0;
  // A list marker the next text starts with.
  private marker = "";

  add(text: string, preformatted: boolean): void {
    if (!preformatted) {
      text = text.replace(/\s+/g, " ");
      const atLineStart =
        this.text === "" || this.breaks > 0 || /\s$/.test(this.text);
      if (atLineStart) text = text.trimStart();
    }
    if (!text) return;
    if (this.breaks > 0) {
      this.text = this.text.trimEnd();
      if (this.text) this.text += "\n".repeat(this.breaks);
      this.breaks = 0;
    }
    this.text += this.marker + text;
    this.marker = "";
  }

  // Puts the next text on a new line, `lines - 1` blank lines below the text so far.
  startLine(lines: number): void {
    this.breaks = Math.max(this.breaks, lines);
  }

  // A line break written in the text, which adds to the ones owed.
  lineBreak(): void {
    this.breaks += 1;
  }

  mark(marker: string): void {
    this.marker = marker;
  }

  // Returns the text so far and starts over.
  take(): string {
    const text = this.text.trim();
    this.text = "";
    this.breaks = 0;
    this.marker = "";
    return text;
  }
}

// The cells of a table as a grid, or undefined for a table without cells.
// A cell that spans columns or rows fills its first position and leaves the others empty,
// except in the header rows, where every column repeats the label that spans it.
function tableGrid(
  table: HTMLTableElement,
  title: string,
): MarkdownTable | undefined {
  const trs = Array.from(table.rows);
  let headerRows = 0;
  while (
    headerRows < trs.length &&
    (trs[headerRows].parentElement?.tagName === "THEAD" ||
      (trs[headerRows].cells.length > 0 &&
        Array.from(trs[headerRows].cells).every((c) => c.tagName === "TH")))
  ) {
    headerRows++;
  }

  const grid: string[][] = trs.map(() => []);
  trs.forEach((tr, r) => {
    let c = 0;
    for (const cell of tr.cells) {
      // Positions a cell of an earlier row spans down into are taken.
      while (grid[r][c] !== undefined) c++;
      const text = oneLine(cell);
      const colSpan = Math.min(Math.max(cell.colSpan, 1), MAX_COL_SPAN);
      const rowSpan = Math.min(Math.max(cell.rowSpan, 1), trs.length - r);
      for (let dr = 0; dr < rowSpan; dr++) {
        for (let dc = 0; dc < colSpan; dc++) {
          const first = dr === 0 && dc === 0;
          grid[r + dr][c + dc] = first || r + dr < headerRows ? text : "";
        }
      }
      c += colSpan;
    }
  });

  const width = Math.max(0, ...grid.map((row) => row.length));
  if (width === 0) return undefined;
  return {
    title,
    // Positions no cell covers are empty.
    rows: grid.map((row) =>
      Array.from({ length: width }, (_, c) => row[c] ?? ""),
    ),
    headerRows,
  };
}
