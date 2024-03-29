import { parseKindAndNameFromFile } from "@rilldata/web-common/features/entity-management/file-content-utils";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { V1ResourceName } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

describe("parseKindAndNameFromFile", () => {
  const testCases: Array<
    [
      title: string,
      path: string,
      contents: string,
      expected: V1ResourceName | undefined,
    ]
  > = [
    [
      "implicit kind and name for yaml",
      "sources/AdBids.yaml",
      `type: sql\nsql: select * from read_csv('data/AdBids.csv')`,
      { kind: ResourceKind.Source, name: "AdBids" },
    ],
    [
      "implicit kind and name for sql",
      "models/AdBids_model.sql",
      `select * from AdBids`,
      { kind: ResourceKind.Model, name: "AdBids_model" },
    ],
    [
      "explicit kind and name for yaml",
      "sources/AdBids_dashboard.yaml",
      `\n\nkind : metrics_view\nname:AdBids_dashboard_name\nmodel: AdBids_model\nmeasures: []\ndimensions: []`,
      { kind: ResourceKind.MetricsView, name: "AdBids_dashboard_name" },
    ],
    [
      "explicit kind and name for sql",
      "sources/AdBids_model.sql",
      `\n\n-- @kind : model\n--@name:AdBids_model_name\nselect * from AdBids`,
      { kind: ResourceKind.Model, name: "AdBids_model_name" },
    ],
    [
      "explicit invalid kind for yaml",
      "sources/AdBids_dashboard.yaml",
      `\n\nkind : invalid\nmodel: AdBids_model\nmeasures: []\ndimensions: []`,
      undefined,
    ],
    [
      "explicit invalid kind for sql",
      "sources/AdBids_model.sql",
      `\n\n-- @kind : invalid\nselect * from AdBids`,
      undefined,
    ],
  ];

  testCases.forEach(([title, path, contents, expected]) => {
    it(title, () => {
      expect(parseKindAndNameFromFile(path, contents)).toEqual(expected);
    });
  });
});
