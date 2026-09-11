import { describe, expect, it } from "vitest";
import { isInitialBuild } from "./selectors";

const dashboard = { explore: {} };
const model = { model: {} };

describe("isInitialBuild", () => {
  it("is false once a dashboard exists, even while the runtime is still initializing", () => {
    expect(
      isInitialBuild({ resources: [dashboard, model], initializing: true }),
    ).toBe(false);
  });

  it("is true when no dashboards exist yet and the runtime is still initializing", () => {
    expect(isInitialBuild({ resources: [model], initializing: true })).toBe(
      true,
    );
  });

  it("is false when security policies denied every resource", () => {
    // ListResources returns 200 with an empty list. Reading that as "still building"
    // left embed users on a deny-by-default project staring at a permanent spinner.
    expect(isInitialBuild({ resources: [], initializing: false })).toBe(false);
  });

  it("is false when the project genuinely has no dashboards", () => {
    expect(isInitialBuild({ resources: [model], initializing: false })).toBe(
      false,
    );
  });
});
