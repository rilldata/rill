import { render } from "@testing-library/svelte";
import { describe, expect, it, vi } from "vitest";

// The tooltip provider requires its own context that's set up at the app
// shell level. This test only cares about the watcher-context fallback, so
// stub tooltip components with a minimal slot-passthrough implementation.
vi.mock("@rilldata/web-common/components/tooltip-v2", async () => {
  const Stub = (await import("./__fixtures__/SlotPassthrough.svelte")).default;
  return {
    Root: Stub,
    Trigger: Stub,
    Content: Stub,
  };
});
vi.mock(
  "@rilldata/web-common/components/tooltip/TooltipContent.svelte",
  async () => ({
    default: (await import("./__fixtures__/SlotPassthrough.svelte")).default,
  }),
);

import RuntimeTrafficLights from "./RuntimeTrafficLights.svelte";

describe("RuntimeTrafficLights", () => {
  it("renders without throwing when no watcher context is provided", () => {
    // The component must not throw when mounted outside a
    // <FileAndResourceWatcher>. Its fallback is a static store set to
    // CLOSED, so the tooltip content still renders in the accessible tree.
    expect(() => render(RuntimeTrafficLights)).not.toThrow();
  });
});
