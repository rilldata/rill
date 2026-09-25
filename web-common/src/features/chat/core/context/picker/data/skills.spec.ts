import { describe, expect, it } from "vitest";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import {
  getSkillAgent,
  getSkillPickerItems,
} from "@rilldata/web-common/features/chat/core/context/picker/data/skills.ts";
import { ToolName } from "@rilldata/web-common/features/chat/core/types.ts";

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

describe("getSkillAgent", () => {
  it("maps each chat agent to the agent named in skills", () => {
    expect(getSkillAgent(ToolName.ANALYST_AGENT)).toBe("analyst");
    expect(getSkillAgent(ToolName.DEVELOPER_AGENT)).toBe("developer");
  });

  it("has no skills for other agents", () => {
    expect(getSkillAgent(ToolName.ROUTER_AGENT)).toBeUndefined();
  });
});

describe("getSkillPickerItems", () => {
  it("lists the skills for the analyst with their descriptions", () => {
    const items = getSkillPickerItems(
      [
        skillResource("monthly-close", ["analyst"]),
        skillResource("glossary", ["analyst", "developer"]),
      ],
      "analyst",
    );
    expect(
      items.map((i) => [i.context.skill, i.context.label, i.description]),
    ).toEqual([
      ["monthly-close", "monthly-close", "Use for monthly-close"],
      ["glossary", "glossary", "Use for glossary"],
    ]);
  });

  it("lists only the skills for the given agent", () => {
    const resources = [
      skillResource("rill-model", ["developer"]),
      skillResource("monthly-close", ["analyst"]),
      skillResource("glossary", ["analyst", "developer"]),
    ];
    expect(
      getSkillPickerItems(resources, "analyst").map((i) => i.context.skill),
    ).toEqual(["monthly-close", "glossary"]);
    expect(
      getSkillPickerItems(resources, "developer").map((i) => i.context.skill),
    ).toEqual(["rill-model", "glossary"]);
  });

  it("does not list skills with a reconcile error", () => {
    const items = getSkillPickerItems(
      [
        skillResource(
          "monthly-close",
          ["analyst"],
          'metrics view "x" not found',
        ),
      ],
      "analyst",
    );
    expect(items).toEqual([]);
  });
});
