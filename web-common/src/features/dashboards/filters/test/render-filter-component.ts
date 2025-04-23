import Filters from "@rilldata/web-common/features/dashboards/filters/Filters.svelte";
import { DEFAULT_STORE_KEY } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { AD_BIDS_METRICS_NAME } from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { initStateManagers } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import { render } from "@testing-library/svelte";

export function renderFilterComponent(hasTimeSeries = false) {
  const { stateManagers, queryClient } = initStateManagers();

  const renderResults = render(Filters, {
    props: {
      timeRanges: [],
      metricsViewName: AD_BIDS_METRICS_NAME,
      hasTimeSeries,
    },
    context: new Map([
      [DEFAULT_STORE_KEY as unknown as string, stateManagers as unknown],
      ["$$_queryClient", queryClient as unknown],
    ]),
  });

  return { stateManagers, queryClient, renderResults };
}
