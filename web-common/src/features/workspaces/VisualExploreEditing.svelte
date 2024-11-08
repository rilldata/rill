<script lang="ts">
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { FileArtifact } from "../entity-management/file-artifact";
  import {
    ResourceKind,
    useFilteredResources,
  } from "../entity-management/resource-selectors";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
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
  import type { V1Explore } from "@rilldata/web-common/runtime-client";
  import { useExploreStore } from "../dashboards/stores/dashboard-stores";
  import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";

  const itemTypes = ["measures", "dimensions"] as const;

  export let fileArtifact: FileArtifact;
  export let exploreName: string;
  export let exploreResource: V1Explore;
  export let metricsViewName: string;
  export let switchView: () => void;

  $: ({ instanceId } = $runtime);
  $: ({ localContent, remoteContent, saveContent } = fileArtifact);

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

  $: rawTitle = parsedDocument.get("title");
  $: rawMetricsView = parsedDocument.get("metrics_view");
  $: rawDimensions = parsedDocument.get("dimensions");
  $: rawMeasures = parsedDocument.get("measures");
  $: rawTimeZones = parsedDocument.get("time_zones");
  $: rawTheme = parsedDocument.get("theme");
  $: rawTimeRanges = parsedDocument.get("time_ranges");
  $: rawDefaults = parsedDocument.get("defaults");

  $: timeZones = new Set(
    rawTimeZones instanceof YAMLSeq
      ? rawTimeZones.toJSON().filter(isString)
      : [],
  );

  $: timeRanges = new Set(
    rawTimeRanges instanceof YAMLSeq
      ? rawTimeRanges.toJSON().filter(isString)
      : [],
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

  $: defaults =
    rawDefaults instanceof YAMLMap ? (rawDefaults.toJSON() as Defaults) : {};

  $: measureExpression =
    rawMeasures instanceof YAMLMap ? rawMeasures?.get("expr") : "";
  $: dimensionExpression =
    rawDimensions instanceof YAMLMap ? rawDimensions?.get("expr") : "";

  $: themesQuery = useFilteredResources(instanceId, ResourceKind.Theme);

  $: themeNames = ($themesQuery?.data ?? [])
    .map((theme) => theme.meta?.name?.name ?? "")
    .filter((string) => !string.endsWith("--theme"));

  $: themeName = !rawTheme
    ? "Default"
    : typeof rawTheme === "string"
      ? rawTheme
      : rawTheme instanceof YAMLMap
        ? "Custom"
        : undefined;

  $: exploreStateStore = useExploreStore(exploreName);

  $: exploreStore = $exploreStateStore;

  $: newDefaults = constructDefaultState(
    exploreStore?.showTimeComparison,
    exploreStore?.selectedComparisonDimension,
    exploreStore?.visibleDimensionKeys,
    exploreStore?.visibleMeasureKeys,
    exploreStore?.selectedTimeRange,
  );

  $: viewingDefaults = Object.entries(defaults).every(([key, value]) => {
    if (Array.isArray(value) && Array.isArray(newDefaults[key])) {
      return (
        JSON.stringify(value.sort()) === JSON.stringify(newDefaults[key].sort())
      );
    }
    return JSON.stringify(value) === JSON.stringify(newDefaults[key]);
  });

  function isString(value: unknown): value is string {
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

  type Defaults = {
    measures?: string[];
    dimensions?: string[];
    comparison_mode?: "time" | "dimension" | "none";
    comparison_dimension?: string;
    time_comparison?: boolean;
    time_range?: string;
  };

  function constructDefaultState(
    showTimeComparison?: boolean,
    selectedComparisonDimension?: string | undefined,
    visibleDimensionKeys?: Set<string>,
    visibleMeasureKeys?: Set<string>,
    selectedTimeRange?: DashboardTimeControls | undefined,
  ): Defaults {
    const newDefaults: Defaults = {};

    if (showTimeComparison) {
      newDefaults.comparison_mode = "time";
    } else if (selectedComparisonDimension) {
      newDefaults.comparison_mode = "dimension";
      newDefaults.comparison_dimension = selectedComparisonDimension;
    }

    if (visibleDimensionKeys?.size) {
      newDefaults.dimensions = Array.from(visibleDimensionKeys);
    }

    if (visibleMeasureKeys?.size) {
      newDefaults.measures = Array.from(visibleMeasureKeys);
    }

    if (selectedTimeRange) {
      newDefaults.time_range = selectedTimeRange.name;
    }

    return newDefaults;
  }
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
      lockable
      label="Metrics view referenced"
      capitalizeLabel={false}
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
      keyNotSet={!rawTimeZones}
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
      keyNotSet={!rawTimeRanges}
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

    <footer
      slot="footer"
      class="flex flex-col gap-y-2 mt-auto border-t px-5 py-3 w-full text-gray-500"
    >
      {#if viewingDefaults}
        <p>Viewing current default state</p>
      {:else}
        <button
          class="text-primary-600 font-medium w-fit"
          on:click={async () => {
            await updateProperties({ defaults: newDefaults });
          }}
        >
          Use left settings as defaults
        </button>
      {/if}
      <p>
        For more options,
        <button on:click={switchView} class="text-primary-600 font-medium">
          edit in YAML
        </button>
      </p>
    </footer>
  </SidebarWrapper>
</div>

<style lang="postcss">
</style>
