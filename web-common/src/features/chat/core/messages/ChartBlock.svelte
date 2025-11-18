<script lang="ts">
  import { page } from "$app/stores";
  import {
    ChartContainer,
    type ChartType,
  } from "@rilldata/web-common/features/components/charts";
  import {
    ResourceKind,
    useFilteredResources,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { mapResolverExpressionToV1Expression } from "@rilldata/web-common/features/explore-mappers/map-metrics-resolver-query-to-dashboard";
  import { Theme } from "@rilldata/web-common/features/themes/theme";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { readable } from "svelte/store";

  export let chartType: ChartType;
  export let chartSpec: any;

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: instanceId = $runtime.instanceId;

  $: spec = readable(chartSpec);

  // Extract time range from the chart spec or use defaults
  $: timeRange = chartSpec.time_range
    ? {
        start: chartSpec.time_range.start,
        end: chartSpec.time_range.end,
        timeZone: chartSpec.time_range.time_zone || "UTC",
      }
    : {
        start: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(), // Default to last 7 days
        end: new Date().toISOString(),
        timeZone: "UTC",
      };

  $: timeAndFilterStore = readable({
    timeRange: timeRange,
    comparisonTimeRange: undefined,
    showTimeComparison: false,
    where: mapResolverExpressionToV1Expression(chartSpec.where) || {
      cond: {
        op: "OPERATION_AND",
        exprs: [],
      },
    },
    timeGrain: chartSpec.time_grain || "TIME_GRAIN_DAY",
    timeRangeState: undefined,
    comparisonTimeRangeState: undefined,
    hasTimeSeries: true,
  });

  $: themesQuery = useFilteredResources(instanceId, ResourceKind.Theme);

  $: themeNames = ($themesQuery?.data ?? [])
    .map((theme) => theme.meta?.name?.name ?? "")
    .filter((string) => !string.endsWith("--theme"));

  $: themeName =
    themeNames.find((name) => name.toLowerCase().startsWith("default")) ??
    themeNames?.[0];

  $: themeQuery = useResource(
    instanceId,
    themeName,
    ResourceKind.Theme,
    {
      enabled: !!themeName,
      select: (data) => {
        if (data.resource?.theme?.spec) {
          return new Theme(data.resource?.theme?.spec);
        } else {
          return undefined;
        }
      },
    },
    queryClient,
  );
</script>

<div
  class="border rounded-md border-gray-150 px-1 py-2 bg-surface w-full h-[400px]"
>
  <ChartContainer
    {chartType}
    {spec}
    {timeAndFilterStore}
    {project}
    theme={$themeQuery?.data}
    showExploreLink
    {organization}
    themeMode="light"
  />
</div>
