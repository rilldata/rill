import { describe, expect, it } from "vitest";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { getSkillPickerItems } from "@rilldata/web-common/features/chat/core/context/picker/data/skills.ts";

function skillResource(
  name: string,
  agents: string[],
  reconcileError?: string,
): V1Resource {
  return {
    meta: { name: { kind: ResourceKind.Skill, name }, reconcileError },
    skill: { spec: { description: `Use for ${name}`, agents } },
  };
}

describe("getSkillPickerItems", () => {
  it("lists the skills for the analyst with their descriptions", () => {
    const items = getSkillPickerItems([
      skillResource("monthly-close", ["analyst"]),
      skillResource("glossary", ["analyst", "developer"]),
    ]);
    expect(
      items.map((i) => [i.context.skill, i.context.label, i.description]),
    ).toEqual([
      ["monthly-close", "monthly-close", "Use for monthly-close"],
      ["glossary", "glossary", "Use for glossary"],
    ]);
  });

  it("does not list skills only for the developer", () => {
    const items = getSkillPickerItems([
      skillResource("rill-model", ["developer"]),
      skillResource("monthly-close", ["analyst"]),
    ]);
    expect(items.map((i) => i.context.skill)).toEqual(["monthly-close"]);
  });

  it("does not list skills with a reconcile error", () => {
    const items = getSkillPickerItems([
      skillResource("monthly-close", ["analyst"], 'metrics view "x" not found'),
    ]);
    expect(items).toEqual([]);
  });
});
