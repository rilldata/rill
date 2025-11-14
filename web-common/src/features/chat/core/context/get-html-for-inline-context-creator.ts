import {
  ChatContextEntryType,
  ChatContextRegex,
} from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
import {
  getDimensionDisplayName,
  getExploreDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import { getExploreNameStore } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
import { getValidDashboardsQueryOptions } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { getExploreValidSpecQueryOptions } from "@rilldata/web-common/features/explores/selectors.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { createQuery } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

export function getHtmlForInlineContextCreator() {
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
      const metricsViewSpec = validSpecResp.data?.metricsViewSpec ?? {};
      return (text: string) => {
        const lines = text.split("\n");
        const htmlLines = lines.map((line) => {
          return line.replace(ChatContextRegex, (_, type, value) => {
            let label = value;
            switch (type) {
              case ChatContextEntryType.Measures:
                label =
                  getMeasureDisplayName(
                    metricsViewSpec.measures?.find((m) => m.name === value),
                  ) ?? value;
                break;

              case ChatContextEntryType.Dimensions:
                label =
                  getDimensionDisplayName(
                    metricsViewSpec.dimensions?.find((d) => d.name === value),
                  ) ?? value;
                break;

              case ChatContextEntryType.Explore:
                label =
                  getExploreDisplayName(
                    exploresSpecResp.data?.find(
                      (e) => e.meta?.name?.name === value,
                    ),
                  ) ?? value;
                break;
            }
            return `<span data-value="${value}" class="underline">${label}</span>`;
          });
        });
        return htmlLines.join("<br>");
      };
    },
  );
}
