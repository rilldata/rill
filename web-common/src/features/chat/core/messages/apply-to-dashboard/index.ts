import { type V1Message } from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
import { get } from "svelte/store";
import { fetchExploreSpec } from "@rilldata/web-common/features/explores/selectors.ts";
import { page } from "$app/stores";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import { goto } from "$app/navigation";
import { dedupe } from "@rilldata/web-common/lib/arrayUtils.ts";

type ApplyToExploreCallData = {
  name: string;
  dimensions?: string[];
  measures?: string[];
  sort_by?: string;
  sort_desc?: boolean;
};

export async function handleApplyToExploreToolCall(callMessage: V1Message) {
  if (!callMessage.contentData) return;
  try {
    const content = JSON.parse(
      callMessage.contentData,
    ) as ApplyToExploreCallData;
    const { explore } = await fetchExploreSpec(
      get(runtime).instanceId,
      content.name,
    );
    const exploreSpec = explore.explore?.state?.validSpec ?? {};
    const currentSearchParams = new URLSearchParams(get(page).url.searchParams);
    console.log(content);

    if (content.dimensions) {
      const currentDimensions = getOrDefaultVisibleValues(
        currentSearchParams.get(ExploreStateURLParams.VisibleDimensions),
        exploreSpec.dimensions ?? [],
      );
      const newDimensions = dedupe(
        [...content.dimensions, ...currentDimensions],
        (d) => d,
      );
      currentSearchParams.set(
        ExploreStateURLParams.VisibleDimensions,
        newDimensions.join(","),
      );
    }

    if (content.measures) {
      const currentMeasures = getOrDefaultVisibleValues(
        currentSearchParams.get(ExploreStateURLParams.VisibleMeasures),
        exploreSpec.measures ?? [],
      );
      const newMeasures = dedupe(
        [...content.measures, ...currentMeasures],
        (m) => m,
      );
      currentSearchParams.set(
        ExploreStateURLParams.VisibleMeasures,
        newMeasures.join(","),
      );
    }

    if (content.sort_by) {
      currentSearchParams.set(ExploreStateURLParams.SortBy, content.sort_by);

      const oldLeaderboardMeasues = (
        currentSearchParams.get(ExploreStateURLParams.LeaderboardMeasures) ??
        content.sort_by
      ).split(",");
      currentSearchParams.set(
        ExploreStateURLParams.LeaderboardMeasures,
        dedupe([content.sort_by, ...oldLeaderboardMeasues], (s) => s).join(","),
      );
    }
    if (content.sort_desc) {
      currentSearchParams.set(
        ExploreStateURLParams.SortDirection,
        content.sort_desc ? "DESC" : "ASC",
      );
    }

    const newUrl = new URL(get(page).url);
    newUrl.search = currentSearchParams.toString();
    void goto(newUrl);
  } catch (err) {
    console.error(err);
  }
}

function getOrDefaultVisibleValues(
  queryParam: string | null,
  values: string[],
) {
  if (queryParam === null || queryParam === "***") return values;
  return queryParam.split(",");
}
