<script lang="ts">
  import { afterNavigate, goto, replaceState } from "$app/navigation";
  import { page } from "$app/stores";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
  import { convertExploreStateToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
  import { updateExploreSessionStore } from "@rilldata/web-common/features/dashboards/url-state/explore-web-view-store";
  import { getUpdatedUrlForExploreState } from "@rilldata/web-common/features/dashboards/url-state/getUpdatedUrlForExploreState";
  import {
    createQueryServiceMetricsViewSchema,
    type V1ExplorePreset,
  } from "@rilldata/web-common/runtime-client";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onMount } from "svelte";

  export let metricsViewName: string;
  export let exploreName: string;
  export let extraKeyPrefix: string | undefined = undefined;
  export let defaultExplorePreset: V1ExplorePreset;
  export let exploreStateFromYAMLConfig: Partial<MetricsExplorerEntity>;
  export let initExploreState: Partial<MetricsExplorerEntity> | undefined =
    undefined;
  export let partialExploreStateFromUrl: Partial<MetricsExplorerEntity>;
  export let exploreStateFromSessionStorage:
    | Partial<MetricsExplorerEntity>
    | undefined;

  const { dashboardStore, validSpecStore, timeRangeSummaryStore } =
    getStateManagers();
  $: exploreSpec = $validSpecStore.data?.explore;
  $: metricsSpec = $validSpecStore.data?.metricsView;

  const metricsViewSchema = createQueryServiceMetricsViewSchema(
    $runtime.instanceId,
    metricsViewName,
  );
  $: ({ error: schemaError } = $metricsViewSchema);
  $: ({ error } = $timeRangeSummaryStore);
  $: timeRangeSummaryError = error as HTTPError;

  afterNavigate(({ from, to, type }) => {
    if (
      // null checks
      !metricsSpec ||
      !exploreSpec ||
      !to ||
      // seems like a sveltekit bug where an additional afterNavigate is triggered with invalid fields
      (from !== null && !from.url)
    ) {
      return;
    }

    const isInit = !$dashboardStore;
    if (isInit) {
      // When a user changes url manually and clears the params the `type` will be "enter"
      // This signal is used in handleExploreInit to make sure we do not use sessionStorage
      const isManualUrlChange = type === "enter";
      handleExploreInit(isManualUrlChange);
      return;
    }

    let partialExplore = partialExploreStateFromUrl;
    let shouldUpdateUrl = false;
    if (exploreStateFromSessionStorage) {
      partialExplore = exploreStateFromSessionStorage;
      shouldUpdateUrl = true;
    }

    const redirectUrl = new URL(to.url);
    metricsExplorerStore.mergePartialExplorerEntity(
      exploreName,
      partialExplore,
      metricsSpec,
    );
    redirectUrl.search = getUpdatedUrlForExploreState(
      exploreSpec,
      defaultExplorePreset,
      partialExplore,
      $page.url.searchParams,
    );
    prevUrl = redirectUrl.toString();

    if (!shouldUpdateUrl || redirectUrl.search === to.url.toString()) {
      return;
    }

    replaceState(redirectUrl, $page.state);
  });

  let prevUrl = "";
  function handleExploreInit(isManualUrlChange: boolean) {
    let initState: Partial<MetricsExplorerEntity> | undefined;
    let shouldUpdateUrl = false;
    if (exploreStateFromSessionStorage && !isManualUrlChange) {
      // if there is state in session storage then merge state from config yaml with the state from session storage
      initState = {
        ...exploreStateFromYAMLConfig,
        ...exploreStateFromSessionStorage,
      };
      shouldUpdateUrl = true;
    } else if ($page.url.searchParams.size === 0) {
      // when there are no params set, state will be state from config yaml and any additional initial state like bookmark
      initState = {
        ...exploreStateFromYAMLConfig,
        ...(initExploreState ?? {}),
      };
      shouldUpdateUrl = !!initExploreState;
    } else {
      // else merge with explore from url
      initState = {
        ...exploreStateFromYAMLConfig,
        ...partialExploreStateFromUrl,
      };
    }

    metricsExplorerStore.init(exploreName, initState);
    const redirectUrl = new URL($page.url);
    redirectUrl.search = getUpdatedUrlForExploreState(
      exploreSpec!,
      defaultExplorePreset,
      initState,
      $page.url.searchParams,
    );
    prevUrl = redirectUrl.toString();

    if (!shouldUpdateUrl || redirectUrl.search === $page.url.search) {
      return;
    }

    replaceState(redirectUrl, $page.state);
    updateExploreSessionStore(
      exploreName,
      extraKeyPrefix,
      $dashboardStore,
      exploreSpec!,
    );
  }

  function gotoNewState() {
    if (!exploreSpec) return;

    const u = new URL(
      `${$page.url.protocol}//${$page.url.host}${$page.url.pathname}`,
    );
    u.search = convertExploreStateToURLSearchParams(
      $dashboardStore,
      exploreSpec,
      defaultExplorePreset,
    );
    const newUrl = u.toString();
    if (!prevUrl || prevUrl === newUrl) return;

    void goto(newUrl);
    updateExploreSessionStore(
      exploreName,
      extraKeyPrefix,
      $dashboardStore,
      exploreSpec,
    );
  }

  // reactive to only dashboardStore
  // but gotoNewState checks other fields
  $: if ($dashboardStore) {
    gotoNewState();
  }

  onMount(() => {
    setTimeout(() => {
      if (!$dashboardStore) {
        handleExploreInit(true);
      }
    });
  });
</script>

{#if schemaError}
  <ErrorPage
    statusCode={schemaError?.response?.status}
    header="Error loading dashboard"
    body="Unable to fetch the schema for this dashboard."
    detail={schemaError?.response?.data?.message}
  />
{:else if timeRangeSummaryError}
  <ErrorPage
    statusCode={timeRangeSummaryError?.response?.status}
    header="Error loading dashboard"
    body="Unable to fetch the time range for this dashboard."
    detail={timeRangeSummaryError?.response?.data?.message}
  />
{:else if $dashboardStore}
  <slot />
{/if}
