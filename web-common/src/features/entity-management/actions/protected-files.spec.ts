import { beforeAll, afterAll, describe, expect, it } from "vitest";
import { setRuntimeEditEnvironment } from "../edit-environment.ts";
import { isPinned, isProtectedDirectory, isReadonly } from "./protected-files";

describe("isPinned", () => {
  it("matches /rill.yaml exactly", () => {
    expect(isPinned("/rill.yaml")).toBe(true);
  });

  it("does not match rill.yaml in a subdirectory", () => {
    expect(isPinned("/foo/rill.yaml")).toBe(false);
  });

  it("does not match a path that merely starts with /rill.yaml", () => {
    expect(isPinned("/rill.yaml.bak")).toBe(false);
  });

  it("does not match unrelated yaml files", () => {
    expect(isPinned("/dashboards/orders.yaml")).toBe(false);
  });
});

describe("isReadonly on local", () => {
  it("does not lock .env files", () => {
    expect(isReadonly("/.env")).toBe(false);
    expect(isReadonly("/foo/.env")).toBe(false);
    expect(isReadonly("/.dev.env")).toBe(false);
  });

  it("does not lock arbitrary files", () => {
    expect(isReadonly("/models/orders.sql")).toBe(false);
  });
});

describe("isReadonly on cloud", () => {
  beforeAll(() => {
    setRuntimeEditEnvironment("cloud");
  });

  afterAll(() => {
    setRuntimeEditEnvironment("local");
  });

  it("locks /.env at the project root", () => {
    expect(isReadonly("/.env")).toBe(true);
  });

  it("locks nested .env files", () => {
    expect(isReadonly("/foo/.env")).toBe(true);
  });

  it("locks prefixed env files like .dev.env", () => {
    expect(isReadonly("/.dev.env")).toBe(true);
    expect(isReadonly("/.prod.env")).toBe(true);
    expect(isReadonly("/foo/.dev.env")).toBe(true);
  });

  it("does not lock files that merely end in 'env'", () => {
    expect(isReadonly("/aenv")).toBe(false);
    expect(isReadonly("/env")).toBe(false);
    expect(isReadonly("/foo.env")).toBe(false);
  });

  it("does not lock .envrc", () => {
    expect(isReadonly("/.envrc")).toBe(false);
  });

  it("does not lock unrelated files", () => {
    expect(isReadonly("/models/orders.sql")).toBe(false);
  });
});

describe("isProtectedDirectory", () => {
  it("matches /tmp exactly", () => {
    expect(isProtectedDirectory("/tmp")).toBe(true);
  });

  it("matches paths inside /tmp", () => {
    expect(isProtectedDirectory("/tmp/scratch.sql")).toBe(true);
  });

  it("matches /.git exactly", () => {
    expect(isProtectedDirectory("/.git")).toBe(true);
  });

  it("matches paths inside /.git", () => {
    expect(isProtectedDirectory("/.git/HEAD")).toBe(true);
  });

  it("does not match /tmpfile", () => {
    expect(isProtectedDirectory("/tmpfile")).toBe(false);
  });

  it("does not match /.gitignore", () => {
    expect(isProtectedDirectory("/.gitignore")).toBe(false);
  });

  it("does not match unrelated paths", () => {
    expect(isProtectedDirectory("/models/orders.sql")).toBe(false);
  });

  it("does not match /tmp as a non-leading segment", () => {
    expect(isProtectedDirectory("/foo/tmp")).toBe(false);
  });
});
