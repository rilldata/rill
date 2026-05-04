import { describe, it, expect } from "vitest";
import {
  matchReadonlyDir,
  matchReadonlyFile,
  ReadonlyEnvFilesRegex,
  type ReadonlyMatcher,
} from "./readonly-files";

describe("matchReadonlyFile", () => {
  it("matches /rill.yaml as a protected file with allowFileEdit", () => {
    const result = matchReadonlyFile("/rill.yaml");
    expect(result).toBeDefined();
    expect(result?.allowFileEdit).toBe(true);
  });

  it("does not match a regular yaml file", () => {
    expect(matchReadonlyFile("/dashboards/orders.yaml")).toBeUndefined();
  });

  it("matches rill.yaml regardless of folder", () => {
    expect(matchReadonlyFile("/foo/rill.yaml")?.allowFileEdit).toBe(true);
  });

  it("returns undefined when no additional matchers are passed and the path is unprotected", () => {
    expect(matchReadonlyFile("/models/orders.sql")).toBeUndefined();
  });

  it("returns the additional matcher whose regex matches the file name", () => {
    const matcher: ReadonlyMatcher = { matcher: /^locked\.yaml$/ };
    expect(matchReadonlyFile("/foo/locked.yaml", [matcher])).toBe(matcher);
  });

  it("returns undefined when no additional matcher matches the file name", () => {
    const matcher: ReadonlyMatcher = { matcher: /^locked\.yaml$/ };
    expect(matchReadonlyFile("/foo/unlocked.yaml", [matcher])).toBeUndefined();
  });

  it("returns the first additional matcher that matches", () => {
    const first: ReadonlyMatcher = { matcher: /^orders\.sql$/ };
    const second: ReadonlyMatcher = {
      matcher: /^orders\.sql$/,
      allowFileEdit: true,
    };
    expect(matchReadonlyFile("/orders.sql", [first, second])).toBe(first);
  });

  it("prefers PROTECTED_FILES over additional matchers", () => {
    const matcher: ReadonlyMatcher = { matcher: /.*/ };
    const result = matchReadonlyFile("/rill.yaml", [matcher]);
    expect(result?.allowFileEdit).toBe(true);
  });

  it("matches /.env via ReadonlyEnvFilesRegex passed as an additional matcher", () => {
    const matcher: ReadonlyMatcher = { matcher: ReadonlyEnvFilesRegex };
    expect(matchReadonlyFile("/.env", [matcher])).toBe(matcher);
  });

  it("matches a nested .env via ReadonlyEnvFilesRegex", () => {
    const matcher: ReadonlyMatcher = { matcher: ReadonlyEnvFilesRegex };
    expect(matchReadonlyFile("/foo/.env", [matcher])).toBe(matcher);
  });

  it("matches prefixed env files like .dev.env via ReadonlyEnvFilesRegex", () => {
    const matcher: ReadonlyMatcher = { matcher: ReadonlyEnvFilesRegex };
    expect(matchReadonlyFile("/.dev.env", [matcher])).toBe(matcher);
    expect(matchReadonlyFile("/.prod.env", [matcher])).toBe(matcher);
    expect(matchReadonlyFile("/foo/.dev.env", [matcher])).toBe(matcher);
    expect(matchReadonlyFile("/foo/.prod.env", [matcher])).toBe(matcher);
  });

  it("does not match files that merely end in 'env' via ReadonlyEnvFilesRegex", () => {
    const matcher: ReadonlyMatcher = { matcher: ReadonlyEnvFilesRegex };
    expect(matchReadonlyFile("/aenv", [matcher])).toBeUndefined();
    expect(matchReadonlyFile("/env", [matcher])).toBeUndefined();
    expect(matchReadonlyFile("/.envrc", [matcher])).toBeUndefined();
    expect(matchReadonlyFile("/foo.env", [matcher])).toBeUndefined();
  });

  it("does not match a non-env file when ReadonlyEnvFilesRegex is the only additional matcher", () => {
    const matcher: ReadonlyMatcher = { matcher: ReadonlyEnvFilesRegex };
    expect(matchReadonlyFile("/models/orders.sql", [matcher])).toBeUndefined();
  });
});

describe("matchReadonlyDir", () => {
  it("matches /tmp exactly", () => {
    expect(matchReadonlyDir("/tmp")).toBe(true);
  });

  it("matches paths inside /tmp", () => {
    expect(matchReadonlyDir("/tmp/scratch.sql")).toBe(true);
  });

  it("matches /.git exactly", () => {
    expect(matchReadonlyDir("/.git")).toBe(true);
  });

  it("matches paths inside /.git", () => {
    expect(matchReadonlyDir("/.git/HEAD")).toBe(true);
  });

  it("does not match /tmpfile (must be exact dir or have a trailing slash)", () => {
    expect(matchReadonlyDir("/tmpfile")).toBe(false);
  });

  it("does not match /.gitignore", () => {
    expect(matchReadonlyDir("/.gitignore")).toBe(false);
  });

  it("does not match unrelated paths", () => {
    expect(matchReadonlyDir("/models/orders.sql")).toBe(false);
  });

  it("does not match /tmp as a non-leading segment", () => {
    expect(matchReadonlyDir("/foo/tmp")).toBe(false);
  });
});
