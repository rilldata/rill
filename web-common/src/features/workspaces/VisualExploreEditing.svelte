<script lang="ts">
  import { createQueryServiceMetricsViewTimeRange } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { FileArtifact } from "../entity-management/file-artifact";
  import {
    ResourceKind,
    useFilteredResources,
  } from "../entity-management/resource-selectors";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import type { LineStatus } from "@rilldata/web-common/components/editor/line-status/state";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import { getNameFromFile } from "../entity-management/entity-mappers";
  import { initLocalUserPreferenceStore } from "../dashboards/user-preferences";
  import { YAMLSeq, Scalar, YAMLMap, parseDocument } from "yaml";
  import StateManagersProvider from "../dashboards/state-managers/StateManagersProvider.svelte";
  import DashboardStateProvider from "../dashboards/stores/DashboardStateProvider.svelte";
  import DashboardUrlStateProvider from "../dashboards/proto-state/DashboardURLStateProvider.svelte";
  import DashboardThemeProvider from "../dashboards/DashboardThemeProvider.svelte";
  import Dashboard from "../dashboards/workspace/Dashboard.svelte";
  import Spinner from "../entity-management/Spinner.svelte";
  import { asyncWait } from "@rilldata/web-common/lib/waitUtils";
  import SidebarWrapper from "../visual-editing/SidebarWrapper.svelte";
  import MeasureDimensionSelector from "../visual-editing/MeasureDimensionSelector.svelte";
  import TimeInput from "../visual-editing/TimeInput.svelte";
  import { DEFAULT_TIMEZONES } from "@rilldata/web-common/lib/time/config";
  import TimeRangeInput from "../visual-editing/TimeRangeInput.svelte";

  import ThemeInput from "../visual-editing/ThemeInput.svelte";

  const itemTypes = ["measures", "dimensions"] as const;

  export let fileArtifact: FileArtifact;
  export let errors: LineStatus[];
  export let switchView: () => void;

  $: ({ instanceId } = $runtime);
  $: ({
    hasUnsavedChanges,
    autoSave,
    path: filePath,
    resourceName,
    fileName,
    getResource,
    localContent,
    remoteContent,
    saveContent,
  } = fileArtifact);

  $: workspace = workspaces.get(filePath);

  $: exploreName = $resourceName?.name ?? getNameFromFile(filePath);

  $: resourceQuery = getResource(queryClient, instanceId);

  $: ({ data } = $resourceQuery);

  $: exploreResource = data?.explore;

  $: metricsViewName = data?.meta?.refs?.find(
    (ref) => ref.kind === ResourceKind.MetricsView,
  )?.name;

  $: initLocalUserPreferenceStore(exploreName);

  $: allErrorsQuery = fileArtifact.getAllErrors(queryClient, instanceId);
  $: allErrors = $allErrorsQuery;
  $: selectedView = workspace.view;

  $: exploreSpec = exploreResource?.state?.validSpec;

  $: parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");

  $: metricsViewsQuery = useFilteredResources(
    instanceId,
    ResourceKind.MetricsView,
  );

  $: metricsViews = $metricsViewsQuery?.data ?? [];

  $: metricsViewNames = metricsViews
    .map((view) => view.meta?.name?.name)
    .filter(isString);

  $: measures = metricsViewSpec?.measures ?? [];
  $: dimensions = metricsViewSpec?.dimensions ?? [];

  $: metricsViewResource = metricsViews.find(
    (view) => view.meta?.name?.name === metricsViewName,
  )?.metricsView;

  $: metricsViewSpec = metricsViewResource?.state?.validSpec;

  //   $: timeZones = new Set(exploreResource?.state?.validSpec?.timeZones ?? []);

  $: defaultTimeRange =
    exploreResource?.state?.validSpec?.defaultPreset?.timeRange;

  $: timeRangeQuery = createQueryServiceMetricsViewTimeRange(
    instanceId,
    metricsViewName ?? "",
    {},
    {
      query: { enabled: !!metricsViewSpec?.timeDimension },
    },
  );

  $: rawTitle = parsedDocument.get("title");
  $: rawMetricsView = parsedDocument.get("metrics_view");
  $: rawDimensions = parsedDocument.get("dimensions");
  $: rawMeasures = parsedDocument.get("measures");
  $: rawTimeZones = parsedDocument.get("time_zones");
  $: rawTheme = parsedDocument.get("theme");
  $: rawTimeRanges = parsedDocument.get("time_ranges");

  $: timeZones = new Set(
    rawTimeZones instanceof YAMLSeq ? rawTimeZones.toJSON() : [],
  );

  $: timeRanges = new Set(
    rawTimeRanges instanceof YAMLSeq ? rawTimeRanges.toJSON() : [],
  );

  $: title = stringGuard(rawTitle);
  $: metricsView = stringGuard(rawMetricsView);

  let selectedMeasureField: "all" | "subset" | "expression";
  let selectedDimensionField: "all" | "subset" | "expression";

  $: selectedMeasureField =
    rawMeasures === "*"
      ? "all"
      : rawMeasures instanceof YAMLSeq
        ? "subset"
        : "expression";

  $: selectedDimensionField =
    rawDimensions === "*"
      ? "all"
      : rawDimensions instanceof YAMLSeq
        ? "subset"
        : "expression";

  $: fields = {
    measures: selectedMeasureField,
    dimensions: selectedDimensionField,
  };

  $: subsets = {
    measures: subsetMeasures,
    dimensions: subsetDimensions,
  };

  $: expressions = {
    measures: measureExpression,
    dimensions: dimensionExpression,
  };

  $: subsetMeasures = new Set(
    rawMeasures instanceof YAMLSeq &&
    rawMeasures.items.every((item) => item instanceof Scalar)
      ? rawMeasures.items.map((item) => item.toString())
      : [],
  );

  $: subsetDimensions = new Set(
    rawDimensions instanceof YAMLSeq &&
    rawDimensions.items.every((item) => item instanceof Scalar)
      ? rawDimensions.items.map((item) => item.toString())
      : [],
  );

  $: measureExpression =
    rawMeasures instanceof YAMLMap ? rawMeasures?.get("expr") : "";
  $: dimensionExpression =
    rawDimensions instanceof YAMLMap ? rawDimensions?.get("expr") : "";

  $: themesQuery = useFilteredResources(instanceId, ResourceKind.Theme);

  $: themeNames = ($themesQuery?.data ?? [])
    .map((theme) => theme.meta?.name?.name ?? "")
    .filter((string) => !string.endsWith("--theme"));

  $: theme = parseTheme(rawTheme);

  async function parseTheme(theme: string | YAMLMap | undefined | unknown) {
    // if (typeof theme === "string") {
    //   const themeResource = await runtimeServiceGetResource(instanceId, {
    //     "name.kind": ResourceKind.Theme,
    //     "name.name": theme,
    //   });

    //   const spec = themeResource.resource?.theme?.spec;

    //   return {
    //     name: theme,
    //     ...spec,
    //   };
    // }

    if (theme instanceof YAMLMap) {
      return {
        name: "Custom",
        ...theme.toJSON().colors,
        custom: true,
      };
    }

    return {
      name: "Default",
      primary: undefined,
      secondary: undefined,
    };
  }

  export function isString(value: unknown): value is string {
    return typeof value === "string";
  }

  function stringGuard(value: unknown | undefined): string {
    return value && typeof value === "string" ? value : "";
  }

  async function updateProperties(
    newRecord: Record<string, unknown>,
    removeProperties?: string[],
  ) {
    Object.entries(newRecord).forEach(([property, value]) => {
      if (!value) {
        parsedDocument.delete(property);
      } else {
        parsedDocument.set(property, value);
      }
    });

    if (removeProperties) {
      removeProperties.forEach((prop) => {
        parsedDocument.delete(prop);
      });
    }

    await saveContent(parsedDocument.toString());
  }

  $: themeName = !rawTheme
    ? "Default"
    : typeof rawTheme === "string"
      ? rawTheme
      : rawTheme instanceof YAMLMap
        ? "Custom"
        : undefined;
</script>

<div class="flex gap-x-2 size-full">
  <div
    class="size-full border overflow-hidden rounded-[2px] bg-background flex flex-col items-center justify-center"
  >
    {#if metricsViewName}
      {#key metricsViewName + exploreName}
        <StateManagersProvider {metricsViewName} {exploreName}>
          <DashboardStateProvider {exploreName}>
            <DashboardUrlStateProvider {metricsViewName}>
              <DashboardThemeProvider>
                <Dashboard {metricsViewName} {exploreName} />
              </DashboardThemeProvider>
            </DashboardUrlStateProvider>
          </DashboardStateProvider>
        </StateManagersProvider>
      {/key}
    {:else}
      <Spinner size="48px" />
    {/if}
  </div>

  <SidebarWrapper title="Edit dashboard" width={320}>
    <Input
      textClass="text-sm"
      label="Title"
      bind:value={title}
      onBlur={async () => {
        await updateProperties({ title });
      }}
      onEnter={async () => {
        await updateProperties({ title });
      }}
    />

    <Input
      label="Metrics view referenced"
      bind:value={metricsView}
      sameWidth
      options={metricsViewNames.map((name) => ({
        label: name,
        value: name,
      }))}
      onChange={async () => {
        await updateProperties({ metrics_view: metricsView });
        await asyncWait(3000);
        if (!metricsViewSpec || !exploreSpec) return;
      }}
    />

    {#each itemTypes as type (type)}
      {@const items = type === "measures" ? measures : dimensions}
      <MeasureDimensionSelector
        {type}
        {items}
        expression={expressions[type]}
        selectedItems={subsets[type]}
        mode={fields[type]}
        setItems={(items) => {
          updateProperties({ [type]: items });
        }}
        onSelectAll={async () => {
          await updateProperties({ [type]: "*" });
        }}
        onSelectSubset={async () => {
          await updateProperties({ [type]: items.map(({ name }) => name) });
        }}
        onSelectExpression={async () => {
          await updateProperties({ [type]: { expr: "*" } });
        }}
        onExpressionBlur={async (value) => {
          await updateProperties({ [type]: { expr: value } });
        }}
        onSelectSubsetItem={async (item) => {
          const deleted = subsets[type].delete(item);
          if (!deleted) {
            subsets[type].add(item);
          }

          await updateProperties({ [type]: Array.from(subsets[type]) });
        }}
      />
    {/each}

    <TimeInput
      selectedItems={timeZones}
      onSelectDefault={async () => {
        await updateProperties({ time_zones: DEFAULT_TIMEZONES });
      }}
      onSelectCustomItem={async (item) => {
        const deleted = timeZones.delete(item);
        if (!deleted) {
          timeZones.add(item);
        }

        await updateProperties({ time_zones: Array.from(timeZones) });
      }}
      restoreDefaults={async () => {
        await updateProperties({ time_zones: DEFAULT_TIMEZONES });
      }}
    />

    <TimeRangeInput
      selectedItems={timeRanges}
      onSelectDefault={async (time_ranges) => {
        await updateProperties({ time_ranges });
      }}
      onSelectCustomItem={async (item) => {
        const deleted = timeRanges.delete(item);
        if (!deleted) {
          timeRanges.add(item);
        }

        await updateProperties({ time_ranges: Array.from(timeRanges) });
      }}
      setTimeRanges={async (time_ranges) => {
        await updateProperties({ time_ranges });
      }}
    />

    <ThemeInput
      {themeName}
      {themeNames}
      theme={exploreSpec?.embeddedTheme}
      onModeChange={async (value) => {
        if (value === "Custom") {
          await updateProperties({
            theme: {
              colors: {
                primary: "hsl(13, 98%, 54%)",
                secondary: "lightgreen",
              },
            },
          });
          return;
        } else if (value === "Default") {
          await updateProperties({}, ["theme"]);
        } else {
          await updateProperties({ theme: value });
        }
      }}
      onColorChange={async (primary, secondary) => {
        await updateProperties({
          theme: {
            colors: {
              primary,
              secondary,
            },
          },
        });
      }}
    />
  </SidebarWrapper>
</div>

<style lang="postcss">
  .wrapper {
    @apply size-full max-w-full max-h-full flex-none;
    @apply overflow-hidden;
    @apply flex gap-x-2;
  }

  h1 {
    @apply text-[16px] font-medium;
  }

  .main-area {
    @apply flex flex-col gap-y-4 size-full p-4 bg-background border;
    @apply flex-shrink overflow-hidden rounded-[2px] relative;
  }

  .section {
    @apply flex flex-none flex-col gap-y-2 justify-start w-full h-fit max-w-full;
  }
</style>
