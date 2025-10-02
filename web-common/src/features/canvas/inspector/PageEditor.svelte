<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
  import ZoneDisplay from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/ZoneDisplay.svelte";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import {
    ResourceKind,
    useFilteredResources,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import MultiSelectInput from "@rilldata/web-common/features/visual-editing/MultiSelectInput.svelte";
  import SidebarWrapper from "@rilldata/web-common/features/visual-editing/SidebarWrapper.svelte";
  import ThemeInput from "@rilldata/web-common/features/visual-editing/ThemeInput.svelte";
  import {
    DEFAULT_RANGES,
    isString,
    numberGuard,
    stringGuard,
  } from "@rilldata/web-common/features/workspaces/visual-util";
  import {
    DEFAULT_TIME_RANGES,
    DEFAULT_TIMEZONES,
  } from "@rilldata/web-common/lib/time/config";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { parseDocument, YAMLMap, YAMLSeq } from "yaml";
  import { DEFAULT_DASHBOARD_WIDTH } from "../layout-util";
  import { allTimeZones } from "@rilldata/web-common/lib/time/timezone";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import { InfoIcon } from "lucide-svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let updateProperties: (
    newRecord: Record<string, unknown>,
    removeProperties?: Array<string | string[]>,
  ) => Promise<void>;
  export let fileArtifact: FileArtifact;
  export let canvasName: string;

  $: ({
    canvasEntity: {
      convertStateToDefault,
      spec: { canvasSpec },
      filters: { setFilters },
    },
  } = getCanvasStore(canvasName, instanceId));

  $: ({ instanceId } = $runtime);

  $: ({ editorContent } = fileArtifact);

  $: parsedDocument = parseDocument($editorContent ?? "");

  $: rawTitle = parsedDocument.get("title");
  $: rawDisplayName = parsedDocument.get("display_name");
  $: rawTheme = parsedDocument.get("theme");
  $: rawTimeRanges = parsedDocument.get("time_ranges");
  $: rawTimeZones = parsedDocument.get("time_zones");
  $: rawMaxWidth = parsedDocument.get("max_width");

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

  async function onSelectTimeRangeItem(item: string) {
    const deleted = timeRanges.delete(item);
    if (!deleted) {
      timeRanges.add(item);
    }

    const time_ranges = Array.from(timeRanges);

    const properties: Record<string, unknown> = {
      time_ranges,
    };

    await updateProperties(properties);
  }

  $: maxWidth = numberGuard(rawMaxWidth) ?? DEFAULT_DASHBOARD_WIDTH;

  $: title = stringGuard(rawTitle) || stringGuard(rawDisplayName);

  $: themesQuery = useFilteredResources(instanceId, ResourceKind.Theme);

  $: themeNames = ($themesQuery?.data ?? [])
    .map((theme) => theme.meta?.name?.name ?? "")
    .filter((string) => !string.endsWith("--theme"));

  $: showFilterBar = $canvasSpec?.filtersEnabled ?? true;
  $: theme = !rawTheme
    ? undefined
    : typeof rawTheme === "string"
      ? rawTheme
      : rawTheme instanceof YAMLMap
        ? $canvasSpec?.embeddedTheme
        : undefined;

  async function toggleFilterBar() {
    const updatedShowFilterBar = !showFilterBar;

    if (!updatedShowFilterBar) {
      setFilters(createAndExpression([]));
    }

    await updateProperties({
      filters: { enable: updatedShowFilterBar },
    });
  }
</script>

<SidebarWrapper type="secondary" disableHorizontalPadding title="Canvas">
  <div class="page-param">
    <Input
      hint="Shown in global header and when deployed to Rill Cloud"
      capitalizeLabel={false}
      size="sm"
      labelGap={2}
      label="Display name"
      bind:value={title}
      onBlur={async () => {
        await updateProperties({ display_name: title }, ["title"]);
      }}
      onEnter={async () => {
        await updateProperties({ display_name: title });
      }}
    />
  </div>
  <div class="page-param">
    <Input
      capitalizeLabel={false}
      size="sm"
      labelGap={2}
      label="Max width"
      inputType="number"
      bind:value={maxWidth}
      onBlur={async () => {
        await updateProperties({ max_width: maxWidth });
      }}
      onEnter={async () => {
        await updateProperties({ max_width: maxWidth });
      }}
    />
  </div>
  <div class="page-param flex flex-col gap-y-2">
    <div
      class="flex items-center justify-between {showFilterBar ? 'pb-1' : ''}"
    >
      <InputLabel
        capitalize={false}
        id="canvas-filter"
        faint={!showFilterBar}
        small
        label="Filter bar"
      />
      <Switch checked={showFilterBar} on:click={toggleFilterBar} small />
    </div>

    {#if showFilterBar}
      <div class="flex flex-col gap-y-2">
        <MultiSelectInput
          small
          label="Time ranges"
          id="canvas-time-range"
          defaultLabel="Default time ranges"
          showLabel={false}
          defaultItems={DEFAULT_RANGES}
          keyNotSet={!rawTimeRanges}
          selectedItems={timeRanges}
          onSelectCustomItem={onSelectTimeRangeItem}
          setItems={async (time_ranges) => {
            if (time_ranges.length === 0) {
              await updateProperties({ time_ranges }, [["time_range"]]);
            } else {
              await updateProperties({ time_ranges });
            }
          }}
          let:item
        >
          {DEFAULT_TIME_RANGES[item]?.label ?? item}
        </MultiSelectInput>

        <MultiSelectInput
          small
          label="Time zones"
          id="visual-explore-zone"
          showLabel={false}
          defaultLabel="Default time zones"
          searchableItems={allTimeZones}
          defaultItems={DEFAULT_TIMEZONES}
          keyNotSet={!rawTimeZones}
          selectedItems={timeZones}
          clearKey={async () => {
            await updateProperties({}, ["time_zones"]);
          }}
          onSelectCustomItem={async (item) => {
            const deleted = timeZones.delete(item);
            if (!deleted) timeZones.add(item);

            await updateProperties({ time_zones: Array.from(timeZones) });
          }}
          setItems={async (time_zones) => {
            await updateProperties({ time_zones });
          }}
          let:item
        >
          <ZoneDisplay iana={item} />
        </MultiSelectInput>

        <Button
          class="group"
          type="subtle"
          large
          onClick={convertStateToDefault}
        >
          Save filter state as default
          <Tooltip distance={16} location="top">
            <InfoIcon size="14px" strokeWidth={2} />
            <TooltipContent slot="tooltip-content">
              By default, the canvas will load with the currently applied time,
              measure and dimension filters
            </TooltipContent>
          </Tooltip>
        </Button>
      </div>
    {/if}
  </div>
  <div class="page-param">
    <ThemeInput
      small
      {theme}
      {themeNames}
      onThemeChange={async (value) => {
        if (!value) {
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
  </div>
</SidebarWrapper>

<style lang="postcss">
  .page-param {
    @apply py-3 px-5;
    @apply border-t;
  }
</style>
