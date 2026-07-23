import { describe, expect, it } from "vitest";
import { getStatementRange, trimRange } from "./query-utils";

function statementAt(sql: string, cursor: number) {
  const range = getStatementRange(sql, cursor);
  return range ? sql.slice(range.from, range.to) : null;
}

describe("getStatementRange", () => {
  it("returns a single statement without requiring a semicolon", () => {
    const sql = "\n  SELECT * FROM events  \n";

    expect(statementAt(sql, sql.indexOf("events"))).toBe(
      "SELECT * FROM events",
    );
  });

  it("returns the statement containing the cursor", () => {
    const sql = "SELECT 1;\n\nSELECT 2;\nSELECT 3";

    expect(statementAt(sql, sql.indexOf("2"))).toBe("SELECT 2;");
  });

  it("ignores semicolons inside strings and quoted identifiers", () => {
    const sql = `SELECT ';' AS value, "semi;colon" AS "name;";\nSELECT 2;`;

    expect(statementAt(sql, sql.indexOf("value"))).toBe(
      `SELECT ';' AS value, "semi;colon" AS "name;";`,
    );
  });

  it("ignores semicolons inside comments", () => {
    const sql = `-- explain; this query\nSELECT 1 /* ; */;\nSELECT 2;`;

    expect(statementAt(sql, sql.indexOf("1"))).toBe(
      "-- explain; this query\nSELECT 1 /* ; */;",
    );
  });

  it("ignores semicolons inside dollar-quoted strings", () => {
    const sql = "SELECT $$one;two$$ AS value;\nSELECT 2;";

    expect(statementAt(sql, sql.indexOf("value"))).toBe(
      "SELECT $$one;two$$ AS value;",
    );
  });

  it("falls back to the previous statement from trailing whitespace", () => {
    const sql = "SELECT 1;\n\n";

    expect(statementAt(sql, sql.length)).toBe("SELECT 1;");
  });

  it("returns null for an empty worksheet", () => {
    expect(getStatementRange("  \n", 1)).toBeNull();
  });
});

describe("trimRange", () => {
  it("trims selected whitespace without changing the selected SQL", () => {
    const sql = "before  SELECT 1;  after";

    expect(trimRange(sql, { from: 6, to: 19 })).toEqual({
      from: 8,
      to: 17,
    });
  });
});
