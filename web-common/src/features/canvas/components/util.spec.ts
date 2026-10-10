import type {
  V1CanvasItem,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import { getComponentInstanceType } from "./util";

function resource(
  renderer: string,
  rendererProperties: Record<string, unknown> = {},
  params: { name: string; type: string }[] = [],
): V1Resource {
  return {
    component: {
      state: {
        validSpec: { renderer, rendererProperties, params },
      },
    },
  } as V1Resource;
}

const externalItem = {
  component: "standalone",
  definedInCanvas: false,
} as V1CanvasItem;

describe("getComponentInstanceType", () => {
  it("keeps non-chart standalone components on their existing renderer", () => {
    expect(getComponentInstanceType(resource("markdown"), externalItem)).toBe(
      "markdown",
    );
  });

  it("keeps legacy Vega custom charts on the multi-query renderer", () => {
    expect(
      getComponentInstanceType(
        resource("custom_chart", { vega_spec: "{}", metrics_sql: ["a", "b"] }),
        externalItem,
      ),
    ).toBe("custom_chart");
  });

  it("routes Flint custom charts through the component reference wrapper", () => {
    expect(
      getComponentInstanceType(
        resource("custom_chart", { spec: { chartType: "Bar Chart" } }),
        externalItem,
      ),
    ).toBe("component_ref");
  });

  it("routes parameterized custom charts through the component reference wrapper", () => {
    expect(
      getComponentInstanceType(
        resource("custom_chart", { vega_spec: "{}" }, [
          { name: "metrics_view", type: "metrics_view" },
        ]),
        externalItem,
      ),
    ).toBe("component_ref");
  });
});
