import {
  MAX_SUGGESTED_PROMPTS,
  resolveSuggestedPrompts,
} from "@rilldata/web-common/features/chat/core/suggested-prompts/suggested-prompts";
import { describe, expect, it } from "vitest";

describe("resolveSuggestedPrompts", () => {
  const configured = [{ label: "Configured", prompt: "Configured prompt" }];
  const generated = [{ label: "Generated", prompt: "Generated prompt" }];
  const fallback = [{ label: "Fallback", prompt: "Fallback prompt" }];

  it("prefers earlier sources", () => {
    expect(resolveSuggestedPrompts([configured, generated, fallback])).toEqual(
      configured,
    );
    expect(resolveSuggestedPrompts([undefined, generated, fallback])).toEqual(
      generated,
    );
    expect(resolveSuggestedPrompts([[], undefined, fallback])).toEqual(
      fallback,
    );
  });

  it("skips sources whose prompts are all blank", () => {
    expect(
      resolveSuggestedPrompts([[{ label: "x", prompt: "  " }], fallback]),
    ).toEqual(fallback);
  });

  it("caps the result", () => {
    const many = Array.from({ length: 8 }, (_, i) => ({
      label: `L${i}`,
      prompt: `P${i}`,
    }));
    expect(resolveSuggestedPrompts([many])).toHaveLength(MAX_SUGGESTED_PROMPTS);
  });

  it("returns an empty list when nothing is usable", () => {
    expect(resolveSuggestedPrompts([undefined, []])).toEqual([]);
  });
});
