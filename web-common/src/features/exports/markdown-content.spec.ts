import { describe, expect, it } from "vitest";
import { markdownParts } from "./markdown-content";

const text = (markdown: string) =>
  markdownParts(markdown).map((part) =>
    part.kind === "text" ? part.text : part.table,
  );

describe("markdownParts", () => {
  it("reduces markdown to the text it shows, line by line", () => {
    expect(
      text(
        "# Title\n\nSome **bold** and _italic_ text with a [link](https://x.y) and `code`.\n\n- item\n  - nested\n- two\n\n1. one\n2. two\n\n> quote\n\n```\nselect 1\n  from t\n```\n\n![Logo](logo.png) first  \nsecond `a < b` &amp; snake_case_name costs 2 * 3 * 4, *really*",
      ),
    ).toStrictEqual([
      "Title\n\nSome bold and italic text with a link and code.\n\n• item\n  • nested\n• two\n\n1. one\n2. two\n\nquote\n\nselect 1\n  from t\n\nLogo first\nsecond a < b & snake_case_name costs 2 * 3 * 4, really",
    ]);
  });

  it("keeps the text of HTML and leaves out what the component does not show", () => {
    expect(
      text(`<style>#row-2 { display: none !important; }</style>
<div class="header" style="background: #123"><h1>Scorecard</h1><span>Region: <b>North</b></span></div>

<div style="display: flex">
  <div class="card"><div class="label">Revenue</div><div class="value">$1.2M</div></div>
  <div class="card"><div class="label">Orders</div><div class="value">3,456</div></div>
  <div hidden>draft</div><div style="color: red; display: none">off</div>
</div>
<script>alert(1)</script>
<p>Line one<br>line two<br><br>after a gap</p>`),
    ).toStrictEqual([
      "Scorecard\nRegion: North\nRevenue\n$1.2M\nOrders\n3,456\n\nLine one\nline two\n\nafter a gap",
    ]);
  });

  it("has no parts for a text that shows nothing", () => {
    expect(
      markdownParts(
        `<style>.card { color: red }</style>\n<div style="height: 10px"></div>`,
      ),
    ).toStrictEqual([]);
  });

  it("splits the text at its tables, written in markdown or HTML, and names each after its caption or heading", () => {
    expect(
      text(`## Q1 results

Intro.

| Rep | Revenue |
|---|---|
| **Ana** | $1,200 |

<table>
  <caption>Targets by <em>team</em></caption>
  <tr><th>Team</th><th>Target</th></tr>
  <tr><td>
    North &amp; East
  </td><td>10<br>units</td></tr>
</table>

Footnote.`),
    ).toStrictEqual([
      "Q1 results\n\nIntro.",
      {
        title: "Q1 results",
        rows: [
          ["Rep", "Revenue"],
          ["Ana", "$1,200"],
        ],
        headerRows: 1,
      },
      "Targets by team",
      {
        title: "Targets by team",
        rows: [
          ["Team", "Target"],
          ["North & East", "10 units"],
        ],
        headerRows: 1,
      },
      "Footnote.",
    ]);
  });

  it("lays spanning cells out on the table's grid, repeating a spanning label over its columns", () => {
    expect(
      text(`<table>
<thead>
  <tr><th rowspan="2">Team</th><th colspan="2">Q1</th></tr>
  <tr><th>Target</th><th>Actual</th></tr>
</thead>
<tbody>
  <tr><td rowspan="2">North</td><td>10</td><td>12</td></tr>
  <tr><td colspan="2">n/a</td></tr>
  <tr><td>South</td><td>5</td></tr>
</tbody>
</table>`),
    ).toStrictEqual([
      {
        title: "",
        rows: [
          ["Team", "Q1", "Q1"],
          ["Team", "Target", "Actual"],
          ["North", "10", "12"],
          ["", "n/a", ""],
          ["South", "5", ""],
        ],
        headerRows: 2,
      },
    ]);
  });

  it("reads a table without header cells as rows only, and a table in a cell as a table of its own", () => {
    expect(
      text(
        `<table><tr><td>Spend <table><tr><th>Week</th></tr><tr><td>1</td></tr></table></td><td>$5</td></tr></table>`,
      ),
    ).toStrictEqual([
      { title: "", rows: [["Spend Week 1", "$5"]], headerRows: 0 },
      { title: "", rows: [["Week"], ["1"]], headerRows: 1 },
    ]);
  });
});
