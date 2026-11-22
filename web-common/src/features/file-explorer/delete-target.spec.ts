import type { V1DirEntry } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import { ResourceKind } from "../entity-management/resource-selectors";
import { getFileDeleteTarget, getFolderDeleteTarget } from "./delete-target";
import { transformFileList } from "./transform-file-list";

describe("getFileDeleteTarget", () => {
  it.each([
    [ResourceKind.Canvas, "canvas"],
    [ResourceKind.Explore, "explore"],
    [ResourceKind.MetricsView, "metrics_view"],
    [ResourceKind.Model, "model"],
    [ResourceKind.Source, "source"],
  ])("names the resource a %s file defines", (kind, expected) => {
    expect(
      getFileDeleteTarget("/dashboards/orders.yaml", { kind, name: "orders" }),
    ).toEqual({
      type: "file",
      path: "/dashboards/orders.yaml",
      resource: { kind: expected, name: "orders" },
    });
  });

  it("describes a file of any other resource kind as a plain file", () => {
    expect(
      getFileDeleteTarget("/themes/brand.yaml", {
        kind: ResourceKind.Theme,
        name: "brand",
      }),
    ).toEqual({ type: "file", path: "/themes/brand.yaml" });
  });

  it("describes a file without a parsed resource as a plain file", () => {
    expect(getFileDeleteTarget("/dashboards/broken.yaml", undefined)).toEqual({
      type: "file",
      path: "/dashboards/broken.yaml",
    });
  });
});

describe("getFolderDeleteTarget", () => {
  const tree = transformFileList([
    { path: "/dashboards", isDir: true },
    { path: "/dashboards/orders.yaml", isDir: false },
    { path: "/dashboards/archive", isDir: true },
    { path: "/dashboards/archive/orders_2024.yaml", isDir: false },
    { path: "/scratch", isDir: true },
    { path: "/drafts", isDir: true },
    { path: "/drafts/notes", isDir: true },
    { path: "/rill.yaml", isDir: false },
  ] as V1DirEntry[]);

  it("lists every file in the folder, including nested folders", () => {
    expect(getFolderDeleteTarget(tree, "/dashboards")).toEqual({
      type: "folder",
      path: "/dashboards",
      files: [
        "/dashboards/orders.yaml",
        "/dashboards/archive/orders_2024.yaml",
      ],
    });
  });

  it("targets a nested folder by its full path", () => {
    expect(getFolderDeleteTarget(tree, "/dashboards/archive")).toEqual({
      type: "folder",
      path: "/dashboards/archive",
      files: ["/dashboards/archive/orders_2024.yaml"],
    });
  });

  it("still confirms a folder that only holds empty folders", () => {
    expect(getFolderDeleteTarget(tree, "/drafts")).toEqual({
      type: "folder",
      path: "/drafts",
      files: [],
    });
  });

  it("needs no confirmation for an empty folder", () => {
    expect(getFolderDeleteTarget(tree, "/scratch")).toBeUndefined();
  });

  it("needs no confirmation for a folder that is not in the tree", () => {
    expect(getFolderDeleteTarget(tree, "/missing")).toBeUndefined();
  });
});
