<script lang="ts">
  import Back from "@rilldata/web-common/components/icons/Back.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { V1ExploreWebView } from "@rilldata/web-common/runtime-client";
  import { Button } from "../../../components/button";

  const { dashboardStore, validSpecStore, webViewStore, defaultExploreState } =
    getStateManagers();
  $: metricsSpec = $validSpecStore.data?.metricsView ?? {};
  $: exploreSpec = $validSpecStore.data?.explore ?? {};

  $: href = webViewStore.getUrlForView(
    V1ExploreWebView.EXPLORE_WEB_VIEW_OVERVIEW,
    $dashboardStore,
    metricsSpec,
    exploreSpec,
    $defaultExploreState,
  );
</script>

<a class="flex items-center" {href}>
  <Button type="link" forcedStyle="padding: 0; gap: 0px;">
    <Back size="16px" />
    <span>All measures</span>
  </Button>
</a>
