<script lang="ts">
  import { replaceState } from "$app/navigation";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
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
    type V1Explore,
    createRuntimeServiceGetInstance,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { Scalar, YAMLMap, YAMLSeq, parseDocument } from "yaml";
  import { metricsExplorerStore } from "../dashboards/stores/dashboard-stores";
  import ZoneDisplay from "../dashboards/time-controls/super-pill/components/ZoneDisplay.svelte";
  import { FileArtifact } from "../entity-management/file-artifact";
  import {
    ResourceKind,
    useFilteredResources,
  } from "../entity-management/resource-selectors";
  import Tab from "@rilldata/web-common/features/dashboards/tab-bar/Tab.svelte";
  import ExploreDefaultFilterDisplay from "@rilldata/web-common/features/dashboards/workspace/ExploreDefaultFilterDisplay.svelte";
  import MeasureDimensionSelector from "../visual-editing/MeasureDimensionSelector.svelte";
  import MultiSelectInput from "../visual-editing/MultiSelectInput.svelte";
  import SidebarWrapper from "../visual-editing/SidebarWrapper.svelte";
  import ThemeInput from "../visual-editing/ThemeInput.svelte";

  const itemTypes = ["measures", "dimensions"] as const;

  let activeTab: "options" | "filters" = "options";

  export let fileArtifact: FileArtifact;
  export let exploreName: string;
  export let exploreResource: V1Explore | undefined;
  export let metricsViewName: string | undefined;
  export let viewingDashboard: boolean;
  export let autoSave: boolean;
  export let switchView: () => void;

  $: if (exploreSpec) metricsExplorerStore.sync(exploreName, exploreSpec);

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
  ) as Record<string, unknown>;

  $: measureExpression =
    rawMeasures instanceof YAMLMap ? rawMeasures?.get("expr") : "";
  $: dimensionExpression =
    rawDimensions instanceof YAMLMap ? rawDimensions?.get("expr") : "";

  $: themesQuery = useFilteredResources(instanceId, ResourceKind.Theme);

  $: themeNames = ($themesQuery?.data ?? [])
    .map((theme) => theme.meta?.name?.name ?? "")
    .filter((string) => !string.endsWith("--theme"));

  $: defaultThemeQuery = createRuntimeServiceGetInstance(
    instanceId,
    undefined,
    {
      query: {
        select: (data) => data?.instance?.theme,
      },
    },
  );

  $: projectDefaultTheme = $defaultThemeQuery?.data;

  $: theme = !rawTheme
    ? undefined
    : typeof rawTheme === "string"
      ? rawTheme
      : rawTheme instanceof YAMLMap
        ? exploreSpec?.embeddedTheme
        : undefined;

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
  <SidebarWrapper
    type="secondary"
    disableHorizontalPadding
    title="Edit dashboard"
  >
    <div class="mr-4 bg-surface-background" slot="header">
      <div class="flex gap-x-2">
        <Tab
          selected={activeTab === "options"}
          on:click={() => (activeTab = "options")}
        >
          Options
        </Tab>
        <Tab
          selected={activeTab === "filters"}
          on:click={() => (activeTab = "filters")}
        >
          Filters
        </Tab>
      </div>
    </div>
    {#if activeTab === "options"}
      <div class="px-5 flex flex-col gap-y-3 border-t">
        {#if autoSave}
          <p class="text-fg-secondary text-sm mt-2">
            Changes below will be auto-saved.
          </p>
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
                updateProperties(
                  { [type]: Array.from(subsets[type]) },
                  deleteKeys,
                );
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
          {projectDefaultTheme}
          onThemeChange={(value) => {
            if (!value) {
              updateProperties({}, ["theme"]);
            } else {
              updateProperties({ theme: value });
            }
          }}
          onColorChange={(primary, secondary, isDarkMode) => {
            const modeKey = isDarkMode ? "dark" : "light";
            const altMode = isDarkMode ? "light" : "dark";

            // check if theme exists for alt mode
            const setAltMode = !parsedDocument.hasIn(["theme", altMode]);

            parsedDocument.setIn(["theme", modeKey, "primary"], primary);
            parsedDocument.setIn(["theme", modeKey, "secondary"], secondary);

            if (setAltMode) {
              parsedDocument.setIn(["theme", altMode, "primary"], primary);
              parsedDocument.setIn(["theme", altMode, "secondary"], secondary);
            }

            killState();

            updateEditorContent(parsedDocument.toString(), false, autoSave);
          }}
        />

        {#if viewingDashboard}
          <footer
            class="flex flex-col gap-y-4 mt-auto border-t py-5 pb-6 w-full text-sm text-fg-muted"
          >
            <p>
              For more options,
              <button
                on:click={switchView}
                class="text-primary-600 font-medium"
              >
                edit in YAML
              </button>
            </p>
          </footer>
        {/if}
      </div>
    {:else if activeTab === "filters"}
      <ExploreDefaultFilterDisplay {fileArtifact} {autoSave} />
    {/if}
  </SidebarWrapper>
</Inspector>

<style lang="postcss">
</style>
