import type {
  V1DirEntry,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
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
      getFileDeleteTarget(
        "/dashboards/orders.yaml",
        { kind, name: "orders" },
        [],
      ),
    ).toEqual({
      type: "file",
      path: "/dashboards/orders.yaml",
      resource: { kind: expected, name: "orders" },
      dependents: [],
    });
  });

  it("describes a file of any other resource kind as a plain file", () => {
    expect(
      getFileDeleteTarget(
        "/themes/brand.yaml",
        { kind: ResourceKind.Theme, name: "brand" },
        [],
      ),
    ).toEqual({ type: "file", path: "/themes/brand.yaml", dependents: [] });
  });

  it("describes a file without a parsed resource as a plain file", () => {
    expect(
      getFileDeleteTarget("/dashboards/broken.yaml", undefined, []),
    ).toEqual({
      type: "file",
      path: "/dashboards/broken.yaml",
      dependents: [],
    });
  });

  describe("dependents", () => {
    function resource(
      kind: ResourceKind,
      name: string,
      filePath: string,
      refs: { kind: ResourceKind; name: string }[] = [],
    ): V1Resource {
      return { meta: { name: { kind, name }, filePaths: [filePath], refs } };
    }

    const model = { kind: ResourceKind.Model, name: "orders" };
    const metricsView = {
      kind: ResourceKind.MetricsView,
      name: "orders_metrics",
    };
    const component = {
      kind: ResourceKind.Component,
      name: "orders_canvas--component-0-0",
    };
    const resources = [
      resource(model.kind, model.name, "/models/orders.sql"),
      // The ref is written in another case than the model's name, which the runtime accepts.
      resource(
        metricsView.kind,
        metricsView.name,
        "/metrics/orders_metrics.yaml",
        [{ kind: ResourceKind.Model, name: "Orders" }],
      ),
      // An explore defined in the metrics view's own file goes with it.
      resource(
        ResourceKind.Explore,
        "orders_metrics",
        "/metrics/orders_metrics.yaml",
        [metricsView],
      ),
      resource(
        ResourceKind.Explore,
        "orders_explore",
        "/dashboards/orders_explore.yaml",
        [metricsView],
      ),
      resource(
        component.kind,
        component.name,
        "/dashboards/orders_canvas.yaml",
        [metricsView],
      ),
      resource(
        ResourceKind.Canvas,
        "orders_canvas",
        "/dashboards/orders_canvas.yaml",
        [component],
      ),
      resource(ResourceKind.Model, "customers", "/models/customers.sql"),
      resource(
        ResourceKind.MetricsView,
        "customers_metrics",
        "/metrics/customers_metrics.yaml",
        [{ kind: ResourceKind.Model, name: "customers" }],
      ),
    ];

    it("lists the files that depend on the resource, directly or through others", () => {
      expect(
        getFileDeleteTarget("/models/orders.sql", model, resources).dependents,
      ).toEqual([
        "/dashboards/orders_canvas.yaml",
        "/dashboards/orders_explore.yaml",
        "/metrics/orders_metrics.yaml",
      ]);
    });

    it("leaves out the file being deleted", () => {
      expect(
        getFileDeleteTarget(
          "/metrics/orders_metrics.yaml",
          metricsView,
          resources,
        ).dependents,
      ).toEqual([
        "/dashboards/orders_canvas.yaml",
        "/dashboards/orders_explore.yaml",
      ]);
    });

    it("is empty for a resource nothing depends on", () => {
      expect(
        getFileDeleteTarget(
          "/dashboards/orders_explore.yaml",
          { kind: ResourceKind.Explore, name: "orders_explore" },
          resources,
        ).dependents,
      ).toEqual([]);
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
