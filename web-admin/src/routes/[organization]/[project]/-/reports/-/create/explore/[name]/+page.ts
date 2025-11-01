import { ExploreUrlWebView } from "@rilldata/web-common/features/dashboards/url-state/mappers.ts";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import { redirect } from "@sveltejs/kit";

export async function load({ parent, url }) {
  await parent();

  const view = url.searchParams.get(ExploreStateURLParams.WebView);
  if (view !== ExploreUrlWebView.Pivot) {
    const newUrl = new URL(url);
    newUrl.searchParams.set(
      ExploreStateURLParams.WebView,
      ExploreUrlWebView.Pivot,
    );
    throw redirect(307, newUrl.toString());
  }
}
