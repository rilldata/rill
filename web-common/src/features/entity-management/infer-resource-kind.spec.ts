import { inferResourceKind } from "@rilldata/web-common/features/entity-management/infer-resource-kind";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { describe, expect, it } from "vitest";

describe("inferResourceName", () => {
  const testCases: Array<
    [title: string, path: string, contents: string, expected: ResourceKind]
  > = [
    [
      "implicit kind for yaml",
      "sources/AdBids.yaml",
      `connector: sql\nsql: select * from read_csv('data/AdBids.csv')`,
      ResourceKind.Source,
    ],
    [
      "implicit kind for sql",
      "models/AdBids_model.sql",
      `select * from AdBids`,
      ResourceKind.Model,
    ],
    [
      "explicit kind for yaml",
      "sources/AdBids_dashboard.yaml",
      `\n\ntype: metrics_view\nmodel: AdBids_model\nmeasures: []\ndimensions: []`,
      ResourceKind.MetricsView,
    ],
    [
      "explicit kind for sql",
      "sources/AdBids_model.sql",
      `\n\n-- @type: model\nselect * from AdBids`,
      ResourceKind.Model,
    ],
    [
      "explicit invalid kind for yaml",
      "sources/AdBids_dashboard.yaml",
      `\n\ntype : invalid\nmodel: AdBids_model\nmeasures: []\ndimensions: []`,
      ResourceKind.Source,
    ],
    [
      "explicit invalid kind for sql",
      "sources/AdBids_model.sql",
      `\n\n-- @type : invalid\nselect * from AdBids`,
      ResourceKind.Model,
    ],
    [
      "matches the correct type key",
      "dashboards/canvas.yaml",
      `rows:\n  type: invalid\ntype: canvas`,
      ResourceKind.Canvas,
    ],
    [
      "implicit kind for dotted yaml",
      "dashboards/dashboard.canvas.yaml",
      `type: canvas\nrows: []`,
      ResourceKind.Canvas,
    ],
    [
      "implicit kind for dotted sql",
      "models/orders.latest.sql",
      `select * from orders`,
      ResourceKind.Model,
    ],
  ];

  testCases.forEach(([title, path, contents, expected]) => {
    it(title, () => {
      expect(inferResourceKind(path, contents)).toEqual(expected);
    });
  });
});
