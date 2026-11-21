import ReadonlyFilters from "@rilldata/web-common/features/dashboards/filters/ReadonlyFilters.svelte";
import { mockResizeObserverForComponentTesting } from "@rilldata/web-common/features/dashboards/filters/test/filter-test-utils";
import {
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { useTestFilterManager } from "@rilldata/web-common/features/dashboards/url-state/test/url-state-test-utils";
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  RUNTIME_CONTEXT_KEY,
  RuntimeClient,
} from "@rilldata/web-common/runtime-client/v2";
import { render } from "@testing-library/svelte";
import { tick } from "svelte";
import { beforeEach, describe, expect, it } from "vitest";

describe("ReadonlyFilters", () => {
  mockAnimationsForComponentTesting();
  mockResizeObserverForComponentTesting();
  const getFilterManager = useTestFilterManager({
    [AD_BIDS_METRICS_NAME]: AD_BIDS_METRICS_INIT,
  });

  beforeEach(() => queryClient.clear());

  function renderFilters(
    expressionFilterManager: ReturnType<typeof getFilterManager>,
  ) {
    return render(ReadonlyFilters, {
      props: { expressionFilterManager },
      context: new Map<string | symbol, unknown>([
        ["$$_queryClient", queryClient],
        [
          RUNTIME_CONTEXT_KEY,
          new RuntimeClient({ host: "http://localhost", instanceId: "test" }),
        ],
      ]),
    });
  }

  it("renders a chip per condition when a dimension is filtered twice", () => {
    // An include and an exclude on one dimension is an advanced filter, which parses to a manager
    // per condition. The bar has no advanced mode of its own, so it shows both chips.
    const expressionFilterManager = getFilterManager();
    expressionFilterManager.setUrlParams(
      new URLSearchParams(
        `f=${AD_BIDS_PUBLISHER_DIMENSION} IN ('Facebook') AND ${AD_BIDS_PUBLISHER_DIMENSION} NIN ('Google')`,
      ),
    );

    const { container } = renderFilters(expressionFilterManager);

    const chips = container.querySelectorAll(".chip.dimension");
    expect(chips).toHaveLength(2);
    expect(chips[0]).toHaveTextContent("publisher Facebook");
    expect(chips[1]).toHaveTextContent("Exclude publisher Google");
  });

  it("keeps a chip mounted when the filters are parsed again", async () => {
    // Every URL change parses new managers, so a chip keyed by its manager would remount.
    const expressionFilterManager = getFilterManager();
    expressionFilterManager.setUrlParams(
      new URLSearchParams(`f=${AD_BIDS_PUBLISHER_DIMENSION} IN ('Facebook')`),
    );
    const { container } = renderFilters(expressionFilterManager);
    const [publisherChip] = container.querySelectorAll(".chip.dimension");

    expressionFilterManager.setUrlParams(
      new URLSearchParams(
        `f=${AD_BIDS_PUBLISHER_DIMENSION} IN ('Facebook') AND ${AD_BIDS_DOMAIN_DIMENSION} IN ('facebook.com')`,
      ),
    );
    await tick();

    const chips = container.querySelectorAll(".chip.dimension");
    expect(chips).toHaveLength(2);
    expect([...chips]).toContain(publisherChip);
  });
});
