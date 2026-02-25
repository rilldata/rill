import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { createLinkError } from "@rilldata/web-common/features/explore-mappers/explore-validation";
import { ExploreLinkErrorType } from "@rilldata/web-common/features/explore-mappers/types";
import { getExplorePageUrlSearchParams } from "@rilldata/web-common/features/explore-mappers/utils";
import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";

/**
 * Generates the explore page URL with proper search parameters
 */
export async function generateExploreLink(
  instanceId: string,
  exploreState: Partial<ExploreState>,
  exploreName: string,
  organization?: string | undefined,
  project?: string | undefined,
): Promise<string> {
  try {
    // Build base URL
    const url = getUrlForExplore(exploreName, organization, project);

    // Generate search parameters from explore state
    const searchParams = await getExplorePageUrlSearchParams(
      instanceId,
      exploreName,
      exploreState,
    );

    searchParams.forEach((value, key) => {
      url.searchParams.set(key, value);
    });

    return url.toString();
  } catch (error) {
    throw createLinkError(
      ExploreLinkErrorType.TRANSFORMATION_ERROR,
      `Failed to generate explore link: ${error.message}`,
      error,
    );
  }
}

export function getUrlForExplore(
  exploreName: string,
  organization?: string | undefined,
  project?: string | undefined,
): URL {
  let url: URL;
  if (EmbedStore.isEmbedded()) {
    url = new URL(
      `/-/embed/explore/${encodeURIComponent(exploreName)}`,
      window.location.origin,
    );
  } else if (organization && project) {
    url = new URL(
      `/${organization}/${project}/explore/${encodeURIComponent(exploreName)}`,
      window.location.origin,
    );
  } else {
    url = new URL(
      `/explore/${encodeURIComponent(exploreName)}`,
      window.location.origin,
    );
  }
  return url;
}
