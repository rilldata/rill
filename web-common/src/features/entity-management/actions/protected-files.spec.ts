import { beforeAll, afterAll, describe, expect, it } from "vitest";
import { setRuntimeEditEnvironment } from "../edit-environment.ts";
import { isPinned, isProtectedDirectory, isManaged } from "./protected-files";

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
    expect(isManaged("/.env")).toBe(false);
    expect(isManaged("/foo/.env")).toBe(false);
    expect(isManaged("/.dev.env")).toBe(false);
  });

  it("does not lock arbitrary files", () => {
    expect(isManaged("/models/orders.sql")).toBe(false);
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
    expect(isManaged("/.env")).toBe(true);
  });

  it("locks nested .env files", () => {
    expect(isManaged("/foo/.env")).toBe(true);
  });

  it("locks prefixed env files like .dev.env", () => {
    expect(isManaged("/.dev.env")).toBe(true);
    expect(isManaged("/.prod.env")).toBe(true);
    expect(isManaged("/foo/.dev.env")).toBe(true);
  });

  it("does not lock files that merely end in 'env'", () => {
    expect(isManaged("/aenv")).toBe(false);
    expect(isManaged("/env")).toBe(false);
    expect(isManaged("/foo.env")).toBe(false);
  });

  it("does not lock .envrc", () => {
    expect(isManaged("/.envrc")).toBe(false);
  });

  it("does not lock unrelated files", () => {
    expect(isManaged("/models/orders.sql")).toBe(false);
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
