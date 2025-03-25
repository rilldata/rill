<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetCurrentUser,
    type V1Project,
  } from "@rilldata/web-admin/client";
  import { getHomeBookmarkExploreState } from "@rilldata/web-admin/features/bookmarks/selectors";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
  import { convertPresetToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
  import { getMostRecentExploreState } from "@rilldata/web-common/features/dashboards/url-state/most-recent-explore-state";
  import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import {
    getExploreStates,
    useExploreValidSpec,
  } from "@rilldata/web-common/features/explores/selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import DashboardURLStateSyncV2 from "@rilldata/web-common/features/dashboards/url-state/DashboardURLStateSyncV2.svelte";
  import QueriesStatus from "@rilldata/web-common/runtime-client/QueriesStatus.svelte";

  export let organization: string;
  export let project: V1Project;
  export let exploreName: string;

  $: ({ instanceId } = $runtime);
  $: exploreSpecQuery = useExploreValidSpec(instanceId, exploreName);
  $: exploreSpec = $exploreSpecQuery.data?.explore ?? {};
  $: metricsViewSpec = $exploreSpecQuery.data?.metricsView ?? {};
  $: metricsViewName = exploreSpec?.metricsView ?? "";
  $: prefix = `${organization}__${project.name}__`;

  $: fullTimeRangeQuery = useMetricsViewTimeRange(instanceId, metricsViewName, {
    query: {
      enabled: Boolean(metricsViewSpec?.timeDimension),
    },
  });

  $: defaultExplorePreset = getDefaultExplorePreset(
    {
      ...exploreSpec,
      defaultPreset: {},
    },
    metricsViewSpec,
    $fullTimeRangeQuery.data,
  );
  $: ({ partialExploreState: defaultExploreState } =
    convertPresetToExploreState(
      metricsViewSpec,
      exploreSpec,
      defaultExplorePreset,
    ));

  $: explorePresetFromYAMLConfig = getDefaultExplorePreset(
    exploreSpec,
    metricsViewSpec,
    $fullTimeRangeQuery.data,
  );
  $: ({ partialExploreState: exploreStateFromYAMLConfig } =
    convertPresetToExploreState(
      metricsViewSpec,
      exploreSpec,
      explorePresetFromYAMLConfig,
    ));

  const userQuery = createAdminServiceGetCurrentUser();
  $: exploreStateFromHomeBookmarkQuery = getHomeBookmarkExploreState(
    project.id,
    instanceId,
    metricsViewName,
    exploreName,
    true,
    Boolean($userQuery.data?.user),
  );

  $: ({ partialExploreState: mostRecentPartialExploreState } =
    getMostRecentExploreState(
      exploreName,
      prefix,
      metricsViewSpec,
      exploreSpec,
    ));

  $: ({ exploreStateFromSessionStorage, partialExploreStateFromUrl } =
    getExploreStates(
      exploreName,
      prefix,
      $page.url.searchParams,
      metricsViewSpec,
      exploreSpec,
      explorePresetFromYAMLConfig,
    ));

  $: initExploreState = {
    ...defaultExploreState,
    // ...(exploreStateFromSessionStorage ??
    //   partialExploreStateFromUrl ??
    ...(partialExploreStateFromUrl ??
      mostRecentPartialExploreState ??
      $exploreStateFromHomeBookmarkQuery.data ??
      exploreStateFromYAMLConfig),
  };

  $: partialExploreState = partialExploreStateFromUrl;
  // exploreStateFromSessionStorage ?? partialExploreStateFromUrl;

  $: console.log(
    // !!exploreStateFromSessionStorage,
    !!partialExploreStateFromUrl,
    !!mostRecentPartialExploreState,
    !!$exploreStateFromHomeBookmarkQuery.data,
    !!exploreStateFromYAMLConfig,
  );

  $: errors = []; // TODO
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
    {
      query: $exploreStateFromHomeBookmarkQuery,
      label: "Bookmark",
    },
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
    extraKeyPrefix={prefix}
    defaultExplorePreset={explorePresetFromYAMLConfig}
    {initExploreState}
    {partialExploreState}
  >
    <slot />
  </DashboardURLStateSyncV2>
</QueriesStatus>
