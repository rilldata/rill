<script lang="ts">
  import { afterNavigate, beforeNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceListBookmarks,
    type V1Project,
  } from "@rilldata/web-admin/client";
  import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
  import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
  import { convertPresetToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
  import { getExploreStateFromLocalStorage } from "@rilldata/web-common/features/dashboards/url-state/explore-persisted-store";
  import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    getExploreStates,
    useExploreValidSpec,
  } from "@rilldata/web-common/features/explores/selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { createQueryServiceMetricsViewSchema } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import DashboardURLStateSyncV2 from "@rilldata/web-common/features/dashboards/url-state/DashboardURLStateSyncV2.svelte";

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
  $: bookmarksQuery = createAdminServiceListBookmarks(
    {
      projectId: project.id,
      resourceKind: ResourceKind.Explore,
      resourceName: exploreName,
    },
    {
      query: {
        enabled: Boolean($userQuery.data?.user),
      },
    },
  );
  $: homeBookmark = $bookmarksQuery.data?.bookmarks?.find((b) => b.default);
  $: schemaQuery = createQueryServiceMetricsViewSchema(
    instanceId,
    metricsViewName,
    undefined,
    {
      query: {
        enabled: Boolean(homeBookmark),
      },
    },
  );
  $: exploreStateFromHomeBookmark = getDashboardStateFromUrl(
    homeBookmark?.data ?? "",
    metricsViewSpec,
    exploreSpec,
    $schemaQuery.data?.schema ?? {},
  );

  $: exploreStateFromLocalStorage = getExploreStateFromLocalStorage(
    exploreName,
    prefix,
    metricsViewSpec,
    exploreSpec,
  );

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
    ...(exploreStateFromSessionStorage ??
      partialExploreStateFromUrl ??
      exploreStateFromLocalStorage ??
      exploreStateFromHomeBookmark ??
      exploreStateFromYAMLConfig),
  };

  $: partialExploreState =
    exploreStateFromSessionStorage ?? partialExploreStateFromUrl;

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

  beforeNavigate(({ from, to }) => {
    console.log("beforeNavigate", from?.url?.toString(), to?.url?.toString());
  });

  afterNavigate(({ from, to, type }) => {
    console.log(
      "afterNavigate",
      type,
      from?.url?.toString(),
      to?.url?.toString(),
    );
  });
</script>

<DashboardURLStateSyncV2
  {exploreName}
  extraKeyPrefix={prefix}
  {defaultExplorePreset}
  {initExploreState}
  {partialExploreState}
>
  <slot />
</DashboardURLStateSyncV2>
