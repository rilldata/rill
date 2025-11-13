<script lang="ts">
  import { replaceState } from "$app/navigation";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    DEFAULT_RANGES,
    isString,
    stringGuard,
  } from "@rilldata/web-common/features/workspaces/visual-util";
  import Inspector from "@rilldata/web-common/layout/workspace/Inspector.svelte";
  import {
    DEFAULT_TIMEZONES,
    DEFAULT_TIME_RANGES,
  } from "@rilldata/web-common/lib/time/config";
  import { allTimeZones } from "@rilldata/web-common/lib/time/timezone";
  import {
    TimeRangePreset,
    type DashboardTimeControls,
  } from "@rilldata/web-common/lib/time/types";
  import type { V1Explore } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { InfoIcon } from "lucide-svelte";
  import { Scalar, YAMLMap, YAMLSeq, parseDocument } from "yaml";
  import {
    metricsExplorerStore,
    useExploreState,
  } from "../dashboards/stores/dashboard-stores";
  import ZoneDisplay from "../dashboards/time-controls/super-pill/components/ZoneDisplay.svelte";
  import { FileArtifact } from "../entity-management/file-artifact";
  import {
    ResourceKind,
    useFilteredResources,
  } from "../entity-management/resource-selectors";
  import MeasureDimensionSelector from "../visual-editing/MeasureDimensionSelector.svelte";
  import MultiSelectInput from "../visual-editing/MultiSelectInput.svelte";
  import SidebarWrapper from "../visual-editing/SidebarWrapper.svelte";
  import ThemeInput from "../visual-editing/ThemeInput.svelte";

  const itemTypes = ["measures", "dimensions"] as const;

  export let fileArtifact: FileArtifact;
  export let exploreName: string;
  export let exploreResource: V1Explore | undefined;
  export let metricsViewName: string | undefined;
  export let viewingDashboard: boolean;
  export let autoSave: boolean;
  export let switchView: () => void;

  $: ({ instanceId } = $runtime);
  $: ({ editorContent, path, updateEditorContent } = fileArtifact);

  $: exploreSpec = exploreResource?.state?.validSpec;

  $: parsedDocument = parseDocument($editorContent ?? "");

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
  $: rawDisplayName = parsedDocument.get("display_name");
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

  $: rawMeasureSequence = getSequenceItems(rawMeasures);
  $: rawDimensionSequence = getSequenceItems(rawDimensions);

  $: title = stringGuard(rawTitle) || stringGuard(rawDisplayName);
  $: metricsView = stringGuard(rawMetricsView);

  $: excludeMode = {
    measures: rawMeasures instanceof YAMLMap && rawMeasures.has("exclude"),
    dimensions:
      rawDimensions instanceof YAMLMap && rawDimensions.has("exclude"),
  };

  $: subsetMeasures = new Set(
    rawMeasureSequence.items.every((item) => item instanceof Scalar)
      ? rawMeasureSequence.items.map((item) => item.toString())
      : [],
  );

  $: subsetDimensions = new Set(
    rawDimensionSequence.items.every((item) => item instanceof Scalar)
      ? rawDimensionSequence.items.map((item) => item.toString())
      : [],
  );

  $: fields = {
    measures: getMeasureOrDimensionState(rawMeasures),
    dimensions: getMeasureOrDimensionState(rawDimensions),
  };

  $: subsets = {
    measures: subsetMeasures,
    dimensions: subsetDimensions,
  };

  $: expressions = {
    measures: measureExpression,
    dimensions: dimensionExpression,
  };

  $: defaults = (
    rawDefaults instanceof YAMLMap ? rawDefaults.toJSON() : {}
  ) as Defaults;

  $: measureExpression =
    rawMeasures instanceof YAMLMap ? rawMeasures?.get("expr") : "";
  $: dimensionExpression =
    rawDimensions instanceof YAMLMap ? rawDimensions?.get("expr") : "";

  $: themesQuery = useFilteredResources(instanceId, ResourceKind.Theme);

  $: themeNames = ($themesQuery?.data ?? [])
    .map((theme) => theme.meta?.name?.name ?? "")
    .filter((string) => !string.endsWith("--theme"));

  $: theme = !rawTheme
    ? undefined
    : typeof rawTheme === "string"
      ? rawTheme
      : rawTheme instanceof YAMLMap
        ? exploreSpec?.embeddedTheme
        : undefined;

  $: exploreStateStore = useExploreState(exploreName);

  $: exploreState = $exploreStateStore;

  $: newDefaults = constructDefaultState(
    exploreState?.showTimeComparison,
    exploreState?.selectedComparisonDimension,
    exploreState?.visibleDimensions,
    exploreState?.visibleMeasures,
    exploreState?.selectedTimeRange,
  );

  $: hasDefaultsSet = rawDefaults instanceof YAMLMap;

  $: viewingDefaults =
    hasDefaultsSet &&
    Object.entries(newDefaults).every(([key, value]) => {
      if (Array.isArray(value) && Array.isArray(defaults[key])) {
        return (
          JSON.stringify(value.sort()) === JSON.stringify(defaults[key].sort())
        );
      }
      return JSON.stringify(value) === JSON.stringify(defaults[key]);
    });

  $: if (exploreSpec) metricsExplorerStore.sync(exploreName, exploreSpec);

  function getMeasureOrDimensionState(
    node: unknown,
  ): "all" | "subset" | "expression" | null {
    if (node === "*") {
      return "all";
    } else if (
      node instanceof YAMLSeq ||
      (node instanceof YAMLMap && node.has("exclude"))
    ) {
      return "subset";
    } else if (node instanceof YAMLMap && node.has("expr")) {
      return "expression";
    } else {
      return null;
    }
  }

  function updateProperties(
    newRecord: Record<string, unknown>,
    removeProperties?: Array<string | string[]>,
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
        try {
          if (Array.isArray(prop)) {
            parsedDocument.deleteIn(prop);
          } else {
            parsedDocument.delete(prop);
          }
        } catch {
          // ignore
        }
      });
    }

    killState();

    updateEditorContent(parsedDocument.toString(), false, autoSave);
  }

  function killState() {
    localStorage.removeItem(`${exploreName}-persistentDashboardStore`);

    replaceState(window.location.origin + window.location.pathname, {});
  }

  type Defaults = {
    measures?: string[] | undefined;
    dimensions?: string[] | undefined;
    comparison_mode?: "time" | "dimension" | "none" | undefined;
    comparison_dimension?: string | undefined;
    time_comparison?: boolean | undefined;
    time_range?: string | undefined;
  };

  function constructDefaultState(
    showTimeComparison?: boolean,
    selectedComparisonDimension?: string | undefined,
    visibleDimensions?: string[],
    visibleMeasures?: string[],
    selectedTimeRange?: DashboardTimeControls | undefined,
  ): Defaults {
    const newDefaults: Defaults = {
      measures: undefined,
      dimensions: undefined,
      comparison_mode: undefined,
      comparison_dimension: undefined,
      time_comparison: undefined,
      time_range: undefined,
    };

    if (showTimeComparison) {
      newDefaults.comparison_mode = "time";
    } else if (selectedComparisonDimension) {
      newDefaults.comparison_mode = "dimension";
      newDefaults.comparison_dimension = selectedComparisonDimension;
    }

    if (visibleDimensions?.length) {
      newDefaults.dimensions = [...visibleDimensions];
    }

    if (visibleMeasures?.length) {
      newDefaults.measures = [...visibleMeasures];
    }

    if (
      selectedTimeRange &&
      selectedTimeRange.name !== TimeRangePreset.CUSTOM &&
      selectedTimeRange.name !== TimeRangePreset.ALL_TIME
    ) {
      newDefaults.time_range = selectedTimeRange.name;
    }

    return newDefaults;
  }

  function onSelectTimeRangeItem(item: string) {
    const deleted = timeRanges.delete(item);
    if (!deleted) {
      timeRanges.add(item);
    }

    const time_ranges = Array.from(timeRanges);

    const properties: Record<string, unknown> = {
      time_ranges,
    };

    if (defaults?.time_range === item) {
      properties.defaults = { ...defaults, time_range: time_ranges[0] };
    }

    updateProperties(properties);
  }

  function getSequenceItems(node: unknown): YAMLSeq {
    if (node instanceof YAMLMap) {
      const exclude = node.get("exclude");

      if (exclude instanceof YAMLSeq) {
        return exclude;
      } else {
        return new YAMLSeq();
      }
    } else if (node instanceof YAMLSeq) {
      return node;
    } else {
      return new YAMLSeq();
    }
  }
</script>

<Inspector filePath={path}>
  <SidebarWrapper title="Edit dashboard">
    {#if autoSave}
      <p class="text-slate-500 text-sm">Changes below will be auto-saved.</p>
    {/if}

    <Input
      hint="Shown in global header and when deployed to Rill Cloud"
      capitalizeLabel={false}
      textClass="text-sm"
      label="Display name"
      bind:value={title}
      onBlur={() => {
        updateProperties({ display_name: title }, ["title"]);
      }}
      onEnter={() => {
        updateProperties({ display_name: title });
      }}
    />

    <Input
      hint="View documentation"
      link="https://docs.rilldata.com/reference/project-files/metrics-views"
      lockable
      lockTooltip="Unlock to change metrics view"
      label="Metrics view referenced"
      capitalizeLabel={false}
      bind:value={metricsView}
      sameWidth
      options={metricsViewNames.map((name) => ({
        label: name,
        value: name,
      }))}
      onChange={() => {
        killState();

        updateProperties(
          {
            metrics_view: metricsView,
            measures: "*",
            dimensions: "*",
          },
          ["defaults"],
        );
      }}
    />

    {#each itemTypes as type (type)}
      {@const items = type === "measures" ? measures : dimensions}
      <MeasureDimensionSelector
        {type}
        {items}
        expression={expressions[type]}
        selectedItems={subsets[type]}
        excludeMode={excludeMode[type]}
        mode={fields[type]}
        onSelectAll={() => {
          updateProperties({ [type]: "*" });
        }}
        onSelectExpression={() => {
          updateProperties({ [type]: { expr: "*" } });
        }}
        setItems={(items, exclude) => {
          const deleteKeys = [["defaults", type]];
          if (type === "dimensions") {
            deleteKeys.push(["defaults", "comparison_dimension"]);
            deleteKeys.push(["defaults", "comparison_mode"]);
          }

          if (exclude) {
            updateProperties({ [type]: { exclude: items } }, deleteKeys);
          } else {
            updateProperties({ [type]: items }, deleteKeys);
          }
        }}
        onExpressionBlur={(value) => {
          const deleteKeys = [["defaults", type]];
          if (type === "dimensions") {
            deleteKeys.push(["defaults", "comparison_dimension"]);
            deleteKeys.push(["defaults", "comparison_mode"]);
          }
          updateProperties({ [type]: { expr: value } }, deleteKeys);
        }}
        onSelectSubsetItem={(item) => {
          const deleted = subsets[type].delete(item);
          if (!deleted) {
            subsets[type].add(item);
          }

          const deleteKeys = [["defaults", type]];
          if (type === "dimensions") {
            deleteKeys.push(["defaults", "comparison_dimension"]);
            deleteKeys.push(["defaults", "comparison_mode"]);
          }

          if (excludeMode[type]) {
            updateProperties(
              { [type]: { exclude: Array.from(subsets[type]) } },
              deleteKeys,
            );
          } else {
            updateProperties({ [type]: Array.from(subsets[type]) }, deleteKeys);
          }
        }}
      />
    {/each}

    <MultiSelectInput
      label="Time ranges"
      id="visual-explore-range"
      hint="Time range shortcuts available via the dashboard filter bar"
      defaultItems={DEFAULT_RANGES}
      keyNotSet={!rawTimeRanges}
      selectedItems={timeRanges}
      onSelectCustomItem={onSelectTimeRangeItem}
      setItems={(time_ranges) => {
        if (time_ranges.length === 0) {
          updateProperties({ time_ranges }, [["defaults", "time_range"]]);
        } else {
          updateProperties({ time_ranges });
        }
      }}
      let:item
    >
      {DEFAULT_TIME_RANGES[item]?.label ?? item}
    </MultiSelectInput>

    <MultiSelectInput
      label="Time zones"
      id="visual-explore-zone"
      hint="Time zones selectable via the dashboard filter bar"
      searchableItems={allTimeZones}
      defaultItems={DEFAULT_TIMEZONES}
      keyNotSet={!rawTimeZones}
      selectedItems={timeZones}
      clearKey={() => {
        updateProperties({}, ["time_zones"]);
      }}
      onSelectCustomItem={(item) => {
        const deleted = timeZones.delete(item);
        if (!deleted) timeZones.add(item);

        updateProperties({ time_zones: Array.from(timeZones) });
      }}
      setItems={(time_zones) => {
        updateProperties({ time_zones });
      }}
      let:item
    >
      <ZoneDisplay iana={item} />
    </MultiSelectInput>

    <ThemeInput
      {theme}
      {themeNames}
      onThemeChange={(value) => {
        if (!value) {
          updateProperties({}, ["theme"]);
        } else {
          updateProperties({ theme: value });
        }
      }}
      onColorChange={(primary, secondary, isDarkMode) => {
        // TODO: Update to support dark mode - currently always sets light mode
        // Use new theme structure: theme.light or theme.dark
        const modeKey = isDarkMode ? "dark" : "light";
        updateProperties({
          theme: {
            [modeKey]: {
              primary,
              secondary,
            },
          },
        });
      }}
    />

    <svelte:fragment slot="footer">
      {#if viewingDashboard}
        <footer
          class="flex flex-col gap-y-4 mt-auto border-t px-5 py-5 pb-6 w-full text-sm text-gray-500"
        >
          <p>
            For more options,
            <button on:click={switchView} class="text-primary-600 font-medium">
              edit in YAML
            </button>
          </p>

          <Button
            class="group"
            type="subtle"
            gray={viewingDefaults}
            large
            onClick={() => {
              if (viewingDefaults) {
                updateProperties({}, ["defaults"]);
              } else {
                updateProperties({ defaults: newDefaults });
              }
            }}
          >
            {#if viewingDefaults}
              <span class="flex gap-x-1">
                <p class="group-hover:block hidden">Remove</p>
                <p class="group-hover:hidden">Viewing</p>
                <p>default state</p>
              </span>
            {:else}
              Save dashboard state as default
            {/if}

            <Tooltip distance={8} location="top">
              <InfoIcon
                size="14px"
                strokeWidth={2}
                class={viewingDefaults ? "group-hover:block hidden" : ""}
              />
              <TooltipContent slot="tooltip-content">
                {#if viewingDefaults}
                  Remove default settings for time range, comparison modes and
                  displayed measures/dimensions
                {:else}
                  Overwrite default settings for time range, comparison modes
                  and displayed measures/dimensions with the current dashboard
                  view
                {/if}
              </TooltipContent>
            </Tooltip>
          </Button>
        </footer>
      {/if}
    </svelte:fragment>
  </SidebarWrapper>
</Inspector>

<style lang="postcss">
</style>
