<script lang="ts">
  import { page } from "$app/stores";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
  import { getStatesForExplore } from "@rilldata/web-common/features/dashboards/url-state/get-states-for-explore";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import {
    getExploreStates,
    useExploreValidSpec,
  } from "@rilldata/web-common/features/explores/selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import type { V1ExplorePreset } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import DashboardURLStateSyncV2 from "@rilldata/web-common/features/dashboards/url-state/DashboardURLStateSyncV2.svelte";
  import QueriesStatus from "@rilldata/web-common/runtime-client/QueriesStatus.svelte";
  import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";

  export let exploreName: string;

  $: ({ instanceId } = $runtime);
  $: exploreSpecQuery = useExploreValidSpec(instanceId, exploreName);
  $: exploreSpec = $exploreSpecQuery.data?.explore ?? {};
  $: metricsViewSpec = $exploreSpecQuery.data?.metricsView ?? {};
  $: metricsViewName = exploreSpec?.metricsView ?? "";

  $: fullTimeRangeQuery = useMetricsViewTimeRange(instanceId, metricsViewName);

  let defaultExploreState: Partial<MetricsExplorerEntity> = {};
  let explorePresetFromYAMLConfig: V1ExplorePreset = {};
  let exploreStateFromYAMLConfig: Partial<MetricsExplorerEntity> | undefined =
    undefined;
  let mostRecentPartialExploreState:
    | Partial<MetricsExplorerEntity>
    | undefined = undefined;
  let errors: Error[] = [];

  $: exploreStatesQuery = getStatesForExplore(
    instanceId,
    metricsViewName,
    exploreName,
    undefined,
  );
  $: if ($exploreStatesQuery.data) {
    ({
      defaultExploreState,
      explorePresetFromYAMLConfig,
      exploreStateFromYAMLConfig,
      mostRecentPartialExploreState,
      errors,
    } = $exploreStatesQuery.data);
  }

  $: ({ exploreStateFromSessionStorage, partialExploreStateFromUrl } =
    getExploreStates(
      exploreName,
      undefined,
      $page.url.searchParams,
      metricsViewSpec,
      exploreSpec,
      explorePresetFromYAMLConfig,
    ));

  $: initExploreState = {
    ...defaultExploreState,
    ...(exploreStateFromSessionStorage ??
      partialExploreStateFromUrl ??
      mostRecentPartialExploreState ??
      exploreStateFromYAMLConfig),
  };

  $: partialExploreState =
    exploreStateFromSessionStorage ?? partialExploreStateFromUrl;

  $: console.log(
    !!exploreStateFromSessionStorage,
    !!partialExploreStateFromUrl,
    !!mostRecentPartialExploreState,
    !!exploreStateFromYAMLConfig,
  );

  $: if (errors?.length) {
    const _errs = errors;
    setTimeout(() => {
      eventBus.emit("notification", {
        type: "error",
        message: _errs[0].message,
        options: { persisted: true },
      });
    }, 100);
  }

  $: queries = [
    {
      query: $exploreSpecQuery,
      label: "Explore",
    },
    {
      query: $fullTimeRangeQuery,
      label: "Time Range",
    },
    // We do not `exploreStatesQuery` since it depends on `exploreSpecQuery` and `fullTimeRangeQuery` already
  ];
  // $: console.log(
  //   queries.map(({ query, label }) => `${label}: ${query.isLoading}`),
  // );

  // beforeNavigate(({ from, to }) => {
  //   console.log("beforeNavigate", from?.url?.toString(), to?.url?.toString());
  // });
  //
  // afterNavigate(({ from, to, type }) => {
  //   console.log(
  //     "afterNavigate",
  //     type,
  //     from?.url?.toString(),
  //     to?.url?.toString(),
  //   );
  // });
</script>

<QueriesStatus {queries} longLoadThreshold={2000}>
  <svelte:fragment slot="loading" let:loadingForLong>
    <CtaLayoutContainer>
      <CtaContentContainer>
        <div class="h-36">
          <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
        </div>
        {#if loadingForLong}
          <CtaHeader variant="bold">
            Hang tight! We're building your dashboard...
          </CtaHeader>
        {/if}
      </CtaContentContainer>
    </CtaLayoutContainer>
  </svelte:fragment>
  <svelte:fragment slot="errors" let:errors>
    {errors[0].label}:{errors[0].error}
  </svelte:fragment>
  <DashboardURLStateSyncV2
    {exploreName}
    {explorePresetFromYAMLConfig}
    {initExploreState}
    {partialExploreState}
  >
    <slot />
  </DashboardURLStateSyncV2>
</QueriesStatus>
