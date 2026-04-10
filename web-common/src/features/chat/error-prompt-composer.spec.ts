import { describe, it, expect } from "vitest";
import { composeErrorPrompt } from "./error-prompt-composer";

describe("composeErrorPrompt", () => {
  it("includes error message, file path, and resource type", () => {
    const result = composeErrorPrompt({
      errorMessage: "unexpected token at line 5",
      filePath: "/models/my_model.sql",
      fileContent: "SELECT *\nFROM table\nWHERE x = 1\nAND y = 2\nORDER BY z\nLIMIT 10",
    });

    expect(result).toContain("unexpected token at line 5");
    expect(result).toContain("/models/my_model.sql");
    expect(result).toContain("SQL model");
    expect(result).toContain("Please explain what's wrong and suggest how to fix it.");
  });

  it("includes line number and surrounding context when available", () => {
    const lines = Array.from({ length: 20 }, (_, i) => `line ${i + 1}`);
    const result = composeErrorPrompt({
      errorMessage: "syntax error",
      filePath: "/models/test.sql",
      fileContent: lines.join("\n"),
      lineNumber: 10,
    });

    expect(result).toContain("Line 10");
    // Should include 5 lines above (5-9) and 5 below (11-15)
    expect(result).toContain("line 5");
    expect(result).toContain("line 15");
    // Should NOT include distant lines
    expect(result).not.toContain("line 1\n");
    expect(result).not.toContain("line 20");
  });

  it("includes whole file if short and no line number", () => {
    const content = "SELECT *\nFROM table\nWHERE x = 1";
    const result = composeErrorPrompt({
      errorMessage: "error",
      filePath: "/models/short.sql",
      fileContent: content,
    });

    expect(result).toContain("SELECT *");
    expect(result).toContain("WHERE x = 1");
  });

  it("includes first 30 + last 10 lines for long files without line number", () => {
    const lines = Array.from({ length: 80 }, (_, i) => `line ${i + 1}`);
    const result = composeErrorPrompt({
      errorMessage: "error",
      filePath: "/models/long.sql",
      fileContent: lines.join("\n"),
    });

    expect(result).toContain("line 1");
    expect(result).toContain("line 30");
    expect(result).toContain("line 71");
    expect(result).toContain("line 80");
    expect(result).not.toContain("line 40");
  });

  it("detects resource types from file path", () => {
    expect(
      composeErrorPrompt({
        errorMessage: "e",
        filePath: "/metrics/mv.yaml",
        fileContent: "version: 1",
      }),
    ).toContain("metrics view");

    expect(
      composeErrorPrompt({
        errorMessage: "e",
        filePath: "/dashboards/canvas.yaml",
        fileContent: "type: canvas",
      }),
    ).toContain("canvas dashboard");

    expect(
      composeErrorPrompt({
        errorMessage: "e",
        filePath: "/connectors/pg.yaml",
        fileContent: "driver: postgres",
      }),
    ).toContain("connector");
  });

  it("appends additional error count", () => {
    const result = composeErrorPrompt({
      errorMessage: "first error",
      filePath: "/models/test.sql",
      fileContent: "SELECT 1",
      additionalErrorCount: 3,
    });

    expect(result).toContain("There are also 3 other errors in this file.");
  });

  it("strips credential-like fields from connector file content", () => {
    const content = [
      "driver: postgres",
      "host: db.example.com",
      "password: secret123",
      "token: abc-def-ghi",
      "secret: mysecret",
      "api_key: key123",
      "port: 5432",
    ].join("\n");

    const result = composeErrorPrompt({
      errorMessage: "connection failed",
      filePath: "/connectors/pg.yaml",
      fileContent: content,
    });

    expect(result).not.toContain("secret123");
    expect(result).not.toContain("abc-def-ghi");
    expect(result).not.toContain("mysecret");
    expect(result).not.toContain("key123");
    expect(result).toContain("host: db.example.com");
    expect(result).toContain("port: 5432");
  });

  it("keeps prompt under 2000 characters for typical errors", () => {
    const result = composeErrorPrompt({
      errorMessage: "unexpected token near 'SELECT'",
      filePath: "/models/orders.sql",
      fileContent: Array.from({ length: 30 }, (_, i) => `-- line ${i + 1}: some SQL code here`).join("\n"),
      lineNumber: 15,
    });

    expect(result.length).toBeLessThan(2000);
  });

  it("strips code block when prompt exceeds max length", () => {
    const longContent = Array.from(
      { length: 50 },
      (_, i) => `-- line ${i + 1}: ${"x".repeat(30)}`,
    ).join("\n");
    const result = composeErrorPrompt({
      errorMessage: "some error message",
      filePath: "/models/big.sql",
      fileContent: longContent,
    });

    // Should be under limit after stripping
    expect(result.length).toBeLessThan(2000);
    // Code block should be gone
    expect(result).not.toContain("```");
    expect(result).not.toContain("Relevant code");
    // But error message and closing should remain
    expect(result).toContain("some error message");
    expect(result).toContain("Please explain what's wrong");
  });
});
