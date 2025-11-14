import type { ContextMetadata } from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
import { getExploreNameStore } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
import { getValidDashboardsQueryOptions } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { getExploreValidSpecQueryOptions } from "@rilldata/web-common/features/explores/selectors.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { createQuery } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

export function getContextMetadataStore(): Readable<ContextMetadata> {
  const exploreNameStore = getExploreNameStore();

  const exploresSpecQuery = createQuery(
    getValidDashboardsQueryOptions(),
    queryClient,
  );
  const validSpecQuery = createQuery(
    getExploreValidSpecQueryOptions(exploreNameStore),
    queryClient,
  );

  return derived(
    [exploresSpecQuery, validSpecQuery],
    ([exploresSpecResp, validSpecResp]) => {
      const validExplores = Object.fromEntries(
        (exploresSpecResp.data ?? []).map(
          (e) => [e.meta?.name?.name ?? "", e] as const,
        ),
      );

      const metricsViewSpec = validSpecResp.data?.metricsViewSpec ?? {};
      const measures = Object.fromEntries(
        (metricsViewSpec.measures ?? []).map((m) => [m.name!, m]),
      );
      const dimensions = Object.fromEntries(
        (metricsViewSpec.dimensions ?? []).map((d) => [d.name!, d]),
      );

      return {
        validExplores,
        measures,
        dimensions,
      };
    },
  );
}
