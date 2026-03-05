import { prettyPrintType } from "@rilldata/web-common/features/query/query-utils";
import { describe, expect, it } from "vitest";

describe("prettyPrintType", () => {
  it("returns UNKNOWN for undefined input", () => {
    expect(prettyPrintType(undefined)).toBe("UNKNOWN");
  });

  it("returns UNKNOWN for empty string", () => {
    expect(prettyPrintType("")).toBe("UNKNOWN");
  });

  it("strips the CODE_ prefix", () => {
    expect(prettyPrintType("CODE_INT32")).toBe("INT32");
  });

  it("strips the CODE_ prefix for various types", () => {
    const cases = [
      { input: "CODE_VARCHAR", expected: "VARCHAR" },
      { input: "CODE_FLOAT64", expected: "FLOAT64" },
      { input: "CODE_BOOLEAN", expected: "BOOLEAN" },
      { input: "CODE_TIMESTAMP", expected: "TIMESTAMP" },
      { input: "CODE_DATE", expected: "DATE" },
    ];
    for (const { input, expected } of cases) {
      expect(prettyPrintType(input)).toBe(expected);
    }
  });

  it("returns the code as-is when there is no CODE_ prefix", () => {
    expect(prettyPrintType("INT32")).toBe("INT32");
  });

  it("returns UNKNOWN for UNKNOWN(...) type codes after stripping prefix", () => {
    expect(prettyPrintType("CODE_UNKNOWN(42)")).toBe("UNKNOWN");
  });

  it("returns UNKNOWN for bare UNKNOWN(...) without prefix", () => {
    expect(prettyPrintType("UNKNOWN(999)")).toBe("UNKNOWN");
  });

  it("returns UNKNOWN for UNKNOWN(...) with nested content", () => {
    expect(prettyPrintType("UNKNOWN(some_type)")).toBe("UNKNOWN");
  });

  it("does not treat a bare UNKNOWN string (no parens) as unknown", () => {
    // "UNKNOWN" without parentheses does not start with "UNKNOWN("
    expect(prettyPrintType("UNKNOWN")).toBe("UNKNOWN");
  });

  it("only strips the first CODE_ occurrence", () => {
    // replace(/^CODE_/, "") only strips the leading prefix
    expect(prettyPrintType("CODE_CODE_INT32")).toBe("CODE_INT32");
  });
});
