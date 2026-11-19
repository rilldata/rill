import { WorkbookColumnType } from "@rilldata/web-common/proto/gen/rill/runtime/v1/workbook_pb";
import { describe, expect, it } from "vitest";
import type { MarkdownTable } from "../markdown-content";
import { markdownExportTable } from "./markdown-table";

const table = (headerRows: number, ...rows: string[][]): MarkdownTable => ({
  title: "",
  rows,
  headerRows,
});

const { NUMBER, STRING } = WorkbookColumnType;

describe("markdownExportTable", () => {
  it("types the cells that show numbers and keeps how each column shows them", () => {
    const result = markdownExportTable(
      table(
        1,
        ["Rep", "Revenue", "Share", "Deals", "Balance", "Score", "Joined"],
        ["Ana", "$1,234.50", "83.3%", "1,200", "12.50€", "7", "2024-01-05"],
        ["Bo", "-$20", "5%", "15", "−3€", "8.25", "2024-02-01"],
        ["Cy", "n/a", "", "—", "0€", "+9", "2024-03-09"],
      ),
    );
    expect(
      result.columns.map((c) => [c.displayName, c.type, c.numberFormat]),
    ).toStrictEqual([
      ["Rep", STRING, undefined],
      ["Revenue", NUMBER, '"$"#,##0.00'],
      ["Share", NUMBER, "0.0%"],
      ["Deals", NUMBER, "#,##0"],
      ["Balance", NUMBER, '0.00"€"'],
      ["Score", NUMBER, "General"],
      // A date holds digits but is no number, so its column is text.
      ["Joined", STRING, undefined],
    ]);
    expect(result.columns.map((c) => c.name)).toStrictEqual([
      "col_1",
      "col_2",
      "col_3",
      "col_4",
      "col_5",
      "col_6",
      "col_7",
    ]);
    // Text without a digit may sit in a number column, and an empty cell is empty.
    expect(result.rows).toStrictEqual([
      ["Ana", 1234.5, 0.833, 1200, 12.5, 7, "2024-01-05"],
      ["Bo", -20, 0.05, 15, -3, 8.25, "2024-02-01"],
      ["Cy", "n/a", null, "—", 0, 9, "2024-03-09"],
    ]);
    expect(result.rowNumberFormats).toBeUndefined();
  });

  it("formats by row where a column's rows are different measures", () => {
    const result = markdownExportTable(
      table(
        1,
        ["Metric", "Target", "Actual", "Progress"],
        ["Spend", "$1.2M", "$1,000,000.50", "83.3%"],
        ["Deals", "15", "12", "80%"],
        ["Win rate", "25.5%", "n/a", "−5%"],
        ["Notes", "see above", "", ""],
      ),
    );
    expect(result.columns.map((c) => [c.type, c.numberFormat])).toStrictEqual([
      [STRING, undefined],
      [NUMBER, undefined],
      [NUMBER, undefined],
      [NUMBER, "0.0%"],
    ]);
    // A magnitude is written out, as in a table component's data sheet.
    expect(result.rows).toStrictEqual([
      ["Spend", 1200000, 1000000.5, 0.833],
      ["Deals", 15, 12, 0.8],
      ["Win rate", 0.255, "n/a", -0.05],
      ["Notes", "see above", null, null],
    ]);
    expect(result.rowNumberFormats).toStrictEqual([
      '"$"#,##0.00',
      "General",
      "0.0%",
      "",
    ]);
  });

  it("keeps identifiers and codes as text", () => {
    const result = markdownExportTable(
      table(
        1,
        ["Zip", "Order", "Version", "Mixed"],
        ["02134", "12345678901234567", "1.2.3", "12"],
        ["94107", "22345678901234567", "2.0.0", "3 of 5"],
      ),
    );
    expect(result.columns.map((c) => c.type)).toStrictEqual([
      STRING,
      STRING,
      STRING,
      STRING,
    ]);
    expect(result.rows[0]).toStrictEqual([
      "02134",
      "12345678901234567",
      "1.2.3",
      "12",
    ]);
  });

  it("takes the value behind a formatted measure, which the shown text rounds", () => {
    const result = markdownExportTable(
      table(
        1,
        ["Metric", "Value"],
        ["Revenue", "$1.2M"],
        ["Orders", "3.5k"],
        ["Latency", "2.3 s"],
        ["Rate", "12.3%"],
        ["Total", "1.234 €"],
        ["Average", "12,5 €"],
      ),
      table(
        1,
        ["Metric", "Value"],
        ["Revenue", "1234567.891"],
        ["Orders", "3456"],
        ["Latency", "2300"],
        ["Rate", "0.12345"],
        // A locale that groups thousands with "." and writes decimals after ",".
        ["Total", "1234.4"],
        ["Average", "12.49"],
      ),
    );
    expect(result.rows).toStrictEqual([
      ["Revenue", 1234567.891],
      ["Orders", 3456],
      // An interval has no cell format: the number behind it, shown plainly.
      ["Latency", 2300],
      ["Rate", 0.12345],
      // Not 1.234, which is how the text reads with "." before the decimals.
      ["Total", 1234.4],
      ["Average", 12.49],
    ]);
    expect(result.rowNumberFormats).toStrictEqual([
      '"$"#,##0.00',
      "#,##0",
      "General",
      "0.0%",
      '#,##0"€"',
      '0.0"€"',
    ]);
  });

  it("names the columns from the header rows, or by position without them", () => {
    expect(
      markdownExportTable(
        table(
          2,
          ["Team", "Q1", "Q1", ""],
          ["Team", "Target", "Actual", ""],
          ["North", "10", "12", "ok"],
        ),
      ).columns.map((c) => c.displayName),
    ).toStrictEqual(["Team", "Q1 Target", "Q1 Actual", "Column 4"]);

    const headless = markdownExportTable(table(0, ["Spend", "$5"]));
    expect(headless.columns.map((c) => c.displayName)).toStrictEqual([
      "Column 1",
      "Column 2",
    ]);
    expect(headless.rows).toStrictEqual([["Spend", 5]]);
  });
});
