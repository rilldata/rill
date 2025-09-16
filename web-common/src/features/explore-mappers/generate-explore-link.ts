import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { createLinkError } from "@rilldata/web-common/features/explore-mappers/explore-validation";
import { ExploreLinkErrorType } from "@rilldata/web-common/features/explore-mappers/types";
import { getExplorePageUrlSearchParams } from "@rilldata/web-common/features/explore-mappers/utils";

/**
 * Generates the explore page URL with proper search parameters
 */
export async function generateExploreLink(
  exploreState: Partial<ExploreState>,
  exploreName: string,
  organization?: string | undefined,
  project?: string | undefined,
  isEmbed?: boolean,
): Promise<string> {
  try {
    // Build base URL
    let url: URL;
    if (isEmbed) {
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

    // Generate search parameters from explore state
    const searchParams = await getExplorePageUrlSearchParams(
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
