import { describe, expect, it } from "vitest";
import { parseDotEnv, serializeDotEnv } from "./dot-env";

describe("parseDotEnv", () => {
  it("parses basic KEY=VALUE pairs", () => {
    expect(parseDotEnv("FOO=bar\nBAZ=qux")).toEqual({ FOO: "bar", BAZ: "qux" });
  });

  it("ignores comment-only lines and blank lines", () => {
    const src = "# leading\n\nFOO=bar\n   # indented\nBAZ=qux\n";
    expect(parseDotEnv(src)).toEqual({ FOO: "bar", BAZ: "qux" });
  });

  it("strips trailing inline comments from unquoted values", () => {
    expect(parseDotEnv("FOO=bar # trailing comment")).toEqual({ FOO: "bar" });
  });

  it("treats unquoted `#` as the start of a comment", () => {
    // Standard dotenv behavior: unquoted `#` is a comment marker. Callers
    // that need a literal `#` in the value must quote it.
    expect(parseDotEnv("URL=https://example.com#section")).toEqual({
      URL: "https://example.com",
    });
  });

  it("preserves `#` inside quoted values", () => {
    expect(parseDotEnv('URL="https://example.com#section"')).toEqual({
      URL: "https://example.com#section",
    });
    expect(parseDotEnv("URL='https://example.com#section'")).toEqual({
      URL: "https://example.com#section",
    });
  });

  it("preserves `=` in unquoted values (splits on the first `=` only)", () => {
    expect(parseDotEnv("DSN=postgres://u:p@h/db?opt=1&other=2")).toEqual({
      DSN: "postgres://u:p@h/db?opt=1&other=2",
    });
  });

  it("strips matching surrounding single, double, and backtick quotes", () => {
    expect(parseDotEnv("A='single'\nB=\"double\"\nC=`back`")).toEqual({
      A: "single",
      B: "double",
      C: "back",
    });
  });

  it("expands \\n and \\r escapes inside double quotes", () => {
    expect(parseDotEnv('JSON="{\\n  \\"k\\": 1\\n}"').JSON).toBe(
      '{\n  \\"k\\": 1\n}',
    );
  });

  it("does not expand escapes inside single quotes", () => {
    expect(parseDotEnv("RAW='line1\\nline2'")).toEqual({
      RAW: "line1\\nline2",
    });
  });

  it("supports the `export` prefix", () => {
    expect(parseDotEnv("export FOO=bar")).toEqual({ FOO: "bar" });
  });

  it("returns empty values for KEY= (no value after equals)", () => {
    expect(parseDotEnv("EMPTY=\nFOO=bar")).toEqual({ EMPTY: "", FOO: "bar" });
  });

  it("normalizes CRLF line endings", () => {
    expect(parseDotEnv("FOO=bar\r\nBAZ=qux")).toEqual({
      FOO: "bar",
      BAZ: "qux",
    });
  });

  it("trims whitespace around bare values", () => {
    expect(parseDotEnv("FOO=   bar   ")).toEqual({ FOO: "bar" });
  });

  it("returns an empty object for empty input", () => {
    expect(parseDotEnv("")).toEqual({});
  });
});

describe("serializeDotEnv", () => {
  it("writes plain KEY=VALUE for simple values", () => {
    expect(serializeDotEnv({ FOO: "bar", BAZ: "qux" })).toBe(
      "FOO=bar\nBAZ=qux",
    );
  });

  it("writes KEY= for empty values", () => {
    expect(serializeDotEnv({ EMPTY: "", FOO: "bar" })).toBe("EMPTY=\nFOO=bar");
  });

  it("quotes values containing `#`", () => {
    expect(serializeDotEnv({ URL: "https://example.com#section" })).toBe(
      "URL='https://example.com#section'",
    );
  });

  it("quotes values containing whitespace", () => {
    expect(serializeDotEnv({ NOTE: "two words" })).toBe("NOTE='two words'");
  });

  it("uses double quotes when the value contains a single quote", () => {
    expect(serializeDotEnv({ MSG: "it's fine" })).toBe(`MSG="it's fine"`);
  });

  it("escapes newlines using double-quoted \\n", () => {
    expect(serializeDotEnv({ MULTI: "line1\nline2" })).toBe(
      `MULTI="line1\\nline2"`,
    );
  });
});

describe("round-trip", () => {
  it.each([
    { FOO: "bar" },
    { URL: "https://example.com#section" },
    { PASSWORD: "p@ss#w0rd!" },
    { DSN: "postgres://u:p@h/db?opt=1&other=2" },
    { NOTE: "two words" },
    { MSG: "it's fine" },
    { MULTI: "line1\nline2" },
    { EMPTY: "" },
    { FOO: "bar", URL: "https://example.com#section", MSG: "it's fine" },
  ] as Record<string, string>[])(
    "serialize → parse preserves %j",
    (entries) => {
      expect(parseDotEnv(serializeDotEnv(entries))).toEqual(entries);
    },
  );
});
