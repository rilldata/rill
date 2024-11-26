<script lang="ts">
  import Back from "@rilldata/web-common/components/icons/Back.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { V1ExploreWebView } from "@rilldata/web-common/runtime-client";

  const { dashboardStore, validSpecStore, webViewStore, basePresetStore } =
    getStateManagers();
  $: metricsSpec = $validSpecStore.data?.metricsView ?? {};
  $: exploreSpec = $validSpecStore.data?.explore ?? {};

  $: href = webViewStore.getUrlForView(
    V1ExploreWebView.EXPLORE_WEB_VIEW_OVERVIEW,
    $dashboardStore,
    metricsSpec,
    exploreSpec,
    $basePresetStore,
  );
</script>

<a class="flex flex-row items-center gap-x-1" {href}>
  <span class="ui-copy-icon">
    <Back color="var(--color-primary-600)" size="16px" />
  </span>
  <span class="text-primary-600 font-medium">Overview</span>
</a>
