import { describe, it, expect } from "vitest";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import {
  countByKind,
  groupErrorsByKind,
} from "@rilldata/web-common/features/resources/overview-utils";

function makeResource(kind: string): V1Resource {
  return { meta: { name: { kind, name: `test-${kind}` } } };
}

describe("overview-utils", () => {
  describe("countByKind", () => {
    it("returns empty array for no resources", () => {
      expect(countByKind([])).toEqual([]);
    });

    it("counts resources by kind", () => {
      const resources = [
        makeResource(ResourceKind.Source),
        makeResource(ResourceKind.Source),
        makeResource(ResourceKind.Model),
      ];
      const result = countByKind(resources);
      expect(result).toEqual([
        { kind: ResourceKind.Source, label: "Source", count: 2 },
        { kind: ResourceKind.Model, label: "Model", count: 1 },
      ]);
    });

    it("filters out non-display kinds", () => {
      const resources = [
        makeResource(ResourceKind.Source),
        makeResource(ResourceKind.ProjectParser),
      ];
      const result = countByKind(resources);
      expect(result).toEqual([
        { kind: ResourceKind.Source, label: "Source", count: 1 },
      ]);
    });

    it("preserves display order regardless of input order", () => {
      const resources = [
        makeResource(ResourceKind.Alert),
        makeResource(ResourceKind.Source),
        makeResource(ResourceKind.Model),
      ];
      const result = countByKind(resources);
      expect(result.map((r) => r.kind)).toEqual([
        ResourceKind.Source,
        ResourceKind.Model,
        ResourceKind.Alert,
      ]);
    });

    it("skips resources without meta.name.kind", () => {
      const resources: V1Resource[] = [
        { meta: { name: { kind: ResourceKind.Source, name: "ok" } } },
        { meta: {} },
        {},
      ];
      const result = countByKind(resources);
      expect(result).toEqual([
        { kind: ResourceKind.Source, label: "Source", count: 1 },
      ]);
    });

    it("omits kinds with zero count", () => {
      const resources = [makeResource(ResourceKind.Source)];
      const result = countByKind(resources);
      expect(result).toHaveLength(1);
      expect(result.find((r) => r.kind === ResourceKind.Model)).toBeUndefined();
    });
  });

  describe("groupErrorsByKind", () => {
    it("returns empty array for no resources", () => {
      expect(groupErrorsByKind([])).toEqual([]);
    });

    it("groups errored resources by kind", () => {
      const resources = [
        makeResource(ResourceKind.Source),
        makeResource(ResourceKind.Source),
        makeResource(ResourceKind.Model),
      ];
      const result = groupErrorsByKind(resources);
      expect(result).toEqual([
        { kind: ResourceKind.Source, label: "Source", count: 2 },
        { kind: ResourceKind.Model, label: "Model", count: 1 },
      ]);
    });

    it("sorts by count descending", () => {
      const resources = [
        makeResource(ResourceKind.Model),
        makeResource(ResourceKind.Source),
        makeResource(ResourceKind.Source),
        makeResource(ResourceKind.Source),
        makeResource(ResourceKind.Model),
      ];
      const result = groupErrorsByKind(resources);
      expect(result[0]).toEqual({
        kind: ResourceKind.Source,
        label: "Source",
        count: 3,
      });
      expect(result[1]).toEqual({
        kind: ResourceKind.Model,
        label: "Model",
        count: 2,
      });
    });

    it("includes any resource kind (not limited to displayKinds)", () => {
      const resources = [makeResource(ResourceKind.ProjectParser)];
      const result = groupErrorsByKind(resources);
      expect(result).toEqual([
        { kind: ResourceKind.ProjectParser, label: "ProjectParser", count: 1 },
      ]);
    });

    it("skips resources without meta.name.kind", () => {
      const resources: V1Resource[] = [
        { meta: { name: { kind: ResourceKind.Source, name: "ok" } } },
        { meta: {} },
        {},
      ];
      const result = groupErrorsByKind(resources);
      expect(result).toEqual([
        { kind: ResourceKind.Source, label: "Source", count: 1 },
      ]);
    });
  });
});
