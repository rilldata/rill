<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import ChevronRight from "@rilldata/web-common/components/icons/ChevronRight.svelte";
  import { BaseChart } from "@rilldata/web-common/features/canvas/components/charts/BaseChart";
  import VegaSpecInput from "@rilldata/web-common/features/canvas/inspector/chart/VegaSpecInput.svelte";
  import type { BaseCanvasComponent } from "../components/BaseCanvasComponent";
  import { DEFAULT_MAP_COLOR_RANGE, MapComponent } from "../components/map";
  import MapColorSelector from "../components/map/MapColorSelector.svelte";
  import MapRegionsSelector from "../components/map/MapRegionsSelector.svelte";
  import { PivotCanvasComponent } from "../components/pivot";
  import type { ComponentSpec } from "../components/types";
  import AIGenerateButton from "./AIGenerateButton.svelte";
  import AlignmentInput from "./AlignmentInput.svelte";
  import CanvasFieldSwitcher from "./CanvasFieldSwitcher.svelte";
  import ChartTypeSelector from "./chart/ChartTypeSelector.svelte";
  import MarkSelector from "./chart/MarkSelector.svelte";
  import MetricsSQLInput from "./chart/MetricsSQLInput.svelte";
  import PositionalFieldConfig from "./chart/PositionalFieldConfig.svelte";
  import type { FieldConfig } from "@rilldata/web-common/features/components/charts/types";
  import { get } from "svelte/store";
  import ComparisonInput from "./ComparisonInput.svelte";
  import {
    ephemeralSpecsToDefs,
    type EphemeralMeasureSpec,
  } from "@rilldata/web-common/features/dashboards/ephemeral-measures/canvas";
  import MultiFieldFormatInput from "./fields/MultiFieldFormatInput.svelte";
  import MultiFieldInput from "./fields/MultiFieldInput.svelte";
  import SingleFieldInput from "./fields/SingleFieldInput.svelte";
  import LabelsInput from "./LabelsInput.svelte";
  import MetricSelectorDropdown from "./MetricSelectorDropdown.svelte";
  import SparklineInput from "./SparklineInput.svelte";
  import TableTypeSelector from "./TableTypeSelector.svelte";
  import type { AllKeys, ComponentInputParam } from "./types";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { ComponentRefComponent } from "../components/component-ref";
  import { slotFields } from "../controls/apply-fields";
  import { listFieldTargets, readControls } from "../controls/control-edit";
  import { controlRequest } from "../controls/editor/control-request";

  export let component: BaseCanvasComponent;

  // Track collapsed state for collapsible sections (metrics_sql, vega_spec)
  let collapsedSections: Record<string, boolean> = {
    metrics_sql: false,
    vega_spec: false,
  };

  function toggleSection(key: string) {
    collapsedSections[key] = !collapsedSections[key];
  }

  function updatePositionalField(
    key: AllKeys<ComponentSpec>,
    updatedConfig: FieldConfig | undefined,
    metricsView: string,
  ) {
    localParamValues[key] = updatedConfig;
    // A map's geo dimension needs a color measure, which a regions-only map lacks;
    // set both together so the spec stays valid,
    // defaulting to the regions' measure or the first measure of the metrics view.
    if (
      component instanceof MapComponent &&
      key === "geo_dimension" &&
      updatedConfig?.field
    ) {
      const mapSpec = get(component.specStore);
      if (!mapSpec.color?.measure) {
        const measure =
          mapSpec.regions?.color?.measure ||
          get(component.parent.metricsView.getMetricsViewFromName(metricsView))
            .metricsView?.measures?.[0]?.name ||
          "";
        void component.updateProperties({
          geo_dimension: { ...updatedConfig, type: "nominal" },
          color: { measure, colorRange: DEFAULT_MAP_COLOR_RANGE },
        });
        return;
      }
    }
    component.updateProperty(key, updatedConfig);
  }

  $: ({
    specStore,
    parent: { name: canvasName },
  } = component);

  $: localParamValues = $specStore;

  $: inputParams = component.inputParams(
    component instanceof PivotCanvasComponent
      ? "columns" in $specStore
        ? "table"
        : "pivot"
      : undefined,
  ).options;

  $: metricsView = (
    "metrics_view" in localParamValues ? localParamValues.metrics_view : null
  ) as string | null;

  $: rawComponentEphemeralMeasures =
    "adhoc_measures" in localParamValues &&
    Array.isArray(localParamValues.adhoc_measures)
      ? (localParamValues.adhoc_measures as EphemeralMeasureSpec[])
      : undefined;
  $: componentEphemeralMeasures = ephemeralSpecsToDefs(
    rawComponentEphemeralMeasures,
  );

  // Components that support ephemeral measures declare the param (hidden from
  // the UI); their measure selectors then offer creating and editing them.
  $: supportsEphemeralMeasures = "adhoc_measures" in inputParams;

  $: entries = Object.entries(inputParams) as [
    AllKeys<ComponentSpec>,
    ComponentInputParam,
  ][];

  // "Let viewers choose" under a field input: opens the dialog for a control that lets
  // viewers swap that field, or pick the fields of a list such as a KPI grid's measures
  // (see ViewerControlsEditor).
  const { canvasControls } = featureFlags;
  $: ({ _metricsViews, parsedContent } = component.parent);
  $: offersControls =
    $canvasControls && !(component instanceof ComponentRefComponent);
  $: metricsViewSpec = metricsView
    ? $_metricsViews[metricsView]?.state?.validSpec
    : undefined;
  $: itemPath = component.pathInYAML.slice(0, -1);
  // Fields the item's own controls already swap, and properties whose fields they let viewers pick.
  $: itemControls = readControls($parsedContent, itemPath);
  $: controlledFields = new Set(
    itemControls.flatMap((c) =>
      [c.measure, c.dimension].filter((f) => typeof f === "string"),
    ),
  );
  $: controlledSlots = new Set(
    itemControls.flatMap((c) => (typeof c.slot === "string" ? [c.slot] : [])),
  );
  $: slotTargets = new Set(
    listFieldTargets(inputParams).map((target) => target.target),
  );

  // The measure or dimension a field input holds, if a viewer control could swap it.
  function swappableField(
    config: ComponentInputParam,
    value: unknown,
  ): { kind: "measure" | "dimension"; field: string } | null {
    let field: unknown;
    if (config.type === "measure" || config.type === "dimension") {
      field = value;
    } else if (config.type === "positional") {
      field = (value as { field?: unknown } | undefined)?.field;
    }
    if (typeof field !== "string" || !field || controlledFields.has(field)) {
      return null;
    }
    if (metricsViewSpec?.measures?.some((f) => f.name === field)) {
      return { kind: "measure", field };
    }
    if (metricsViewSpec?.dimensions?.some((f) => f.name === field)) {
      return { kind: "dimension", field };
    }
    return null;
  }

  // A list of fields: viewers pick among the ones it shows, and the designer can offer more.
  function letViewersPick(key: string) {
    controlRequest.set({
      itemPath: itemPath.join("/"),
      draft: {
        kind: "slot",
        field: key,
        options: slotFields(localParamValues, key),
        default: [],
        multiple: true,
        display: "",
        label: "",
        name: "",
      },
    });
  }

  function letViewersChoose(kind: "measure" | "dimension", field: string) {
    controlRequest.set({
      itemPath: itemPath.join("/"),
      draft: {
        kind,
        field,
        options: [],
        default: [],
        multiple: false,
        display: "",
        label: "",
        name: "",
      },
    });
  }
</script>

{#if component instanceof BaseChart}
  <ChartTypeSelector {component} />
{/if}

{#if metricsView && component instanceof PivotCanvasComponent}
  <TableTypeSelector {component} />
{/if}

<div>
  {#each entries as [key, config] (`${component.id}-${key}`)}
    {#if config.showInUI !== false}
      <div
        class="component-param"
        class:grouped={config.meta?.layout === "grouped"}
      >
        <!-- AI GENERATE (opens dev agent sidebar) -->
        {#if config.type === "ai_generate"}
          <AIGenerateButton {component} />

          <!-- TEXT, NUMBER, RILL_TIME -->
        {:else if config.type === "text" || config.type === "number" || config.type === "rill_time"}
          <Input
            inputType={config.type === "number" ? "number" : "text"}
            capitalizeLabel={false}
            textClass="text-sm"
            size="sm"
            placeholder={config?.meta?.placeholder ?? ""}
            labelGap={2}
            label={config.label ?? key}
            bind:value={$specStore[key]}
            onBlur={() => {
              component.updateProperty(key, localParamValues[key]);
            }}
            onEnter={() => {
              component.updateProperty(key, localParamValues[key]);
            }}
          />

          <!-- METRICS SELECTOR -->
        {:else if config.type === "metrics"}
          <MetricSelectorDropdown {component} {key} inputParam={config} />

          <!-- MEASURE / DIMENSION -->
        {:else if metricsView && (config.type === "measure" || config.type === "dimension")}
          <SingleFieldInput
            {canvasName}
            label={config.label ?? key}
            metricName={metricsView}
            id={key}
            type={config.type}
            includeTime={config.meta?.includeTime ?? false}
            timeFieldsOnly={config.meta?.timeFieldsOnly ?? false}
            searchableItems={config.meta?.searchableItems}
            selectedItem={localParamValues[key]}
            ephemeralMeasures={componentEphemeralMeasures}
            component={supportsEphemeralMeasures ? component : undefined}
            isRemovable={config.meta?.isRemovable ?? false}
            onSelect={(field) => {
              component.updateProperty(key, field);
            }}
            onRemove={() => {
              component.updateProperty(key, undefined);
            }}
          />

          <!-- MULTIPLE MEASURE / MULTIPLE DIMENSION / MULTIPLE FIELDS -->
        {:else if metricsView && config.type === "multi_fields"}
          <MultiFieldInput
            {canvasName}
            label={config.label ?? key}
            metricName={metricsView}
            id={key}
            types={config.meta?.allowedTypes ?? ["measure", "dimension"]}
            selectedItems={localParamValues[key]}
            ephemeralMeasures={componentEphemeralMeasures}
            component={supportsEphemeralMeasures ? component : undefined}
            onMultiSelect={(field) => {
              component.updateProperty(key, field);
            }}
          />

          <!-- MULTIPLE FIELDS WITH PER-MEASURE CONDITIONAL FORMATTING -->
        {:else if metricsView && config.type === "multi_fields_format" && component instanceof PivotCanvasComponent}
          <MultiFieldFormatInput
            {component}
            {canvasName}
            label={config.label ?? key}
            metricName={metricsView}
            id={key}
            types={config.meta?.allowedTypes ?? ["measure", "dimension"]}
            selectedItems={localParamValues[key]}
            ephemeralMeasures={componentEphemeralMeasures}
          />

          <!-- BOOLEAN SWITCH -->
        {:else if config.type === "boolean"}
          <div class="flex items-center justify-between py-1">
            <InputLabel
              small
              label={config.label ?? key}
              hint={config.description}
              id={key}
              faint={config.meta?.invertBoolean
                ? ($specStore[key] ?? config.meta?.defaultValue ?? false)
                : !($specStore[key] ?? config.meta?.defaultValue ?? false)}
            />
            <Switch
              checked={config.meta?.invertBoolean
                ? !($specStore[key] ?? config.meta?.defaultValue ?? false)
                : ($specStore[key] ?? config.meta?.defaultValue ?? false)}
              onCheckedChange={(next) => {
                component.updateProperty(
                  key,
                  config.meta?.invertBoolean ? !next : next,
                );
              }}
              small
            />
          </div>

          <!-- TEXT AREA -->
        {:else if config.type === "textarea"}
          <div class="flex flex-col gap-y-2">
            <InputLabel
              hint={config?.description}
              small
              label={config.label ?? key}
              id={key}
            />
            <textarea
              class="w-full p-2 border border-gray-300 rounded-sm"
              rows="8"
              bind:value={$specStore[key]}
              onblur={() => {
                component.updateProperty(key, localParamValues[key]);
              }}
              placeholder={config.label ?? key}
            ></textarea>
          </div>

          <!-- METRICS SQL -->
        {:else if config.type === "metrics_sql"}
          <button class="collapsible-header" onclick={() => toggleSection(key)}>
            <span class="chevron" class:expanded={!collapsedSections[key]}>
              <ChevronRight size="12px" />
            </span>
            <span class="collapsible-label">{config.label ?? key}</span>
            {#if Array.isArray(localParamValues[key])}
              <span class="collapsible-badge"
                >{localParamValues[key].length}</span
              >
            {/if}
          </button>
          {#if !collapsedSections[key]}
            <MetricsSQLInput
              {key}
              label={undefined}
              description={config?.description}
              value={localParamValues[key]}
              onChange={(updatedSQL) => {
                localParamValues[key] = updatedSQL;
                component.updateProperty(key, updatedSQL);
              }}
            />
          {/if}

          <!-- VEGA SPEC -->
        {:else if config.type === "vega_spec"}
          <button class="collapsible-header" onclick={() => toggleSection(key)}>
            <span class="chevron" class:expanded={!collapsedSections[key]}>
              <ChevronRight size="12px" />
            </span>
            <span class="collapsible-label">{config.label ?? key}</span>
          </button>
          {#if !collapsedSections[key]}
            <VegaSpecInput
              value={localParamValues[key]}
              onChange={(updatedSpec) => {
                localParamValues[key] = updatedSpec;
                component.updateProperty(key, updatedSpec);
              }}
            />
          {/if}
          <!-- SELECT DROPDOWN -->
        {:else if config.type === "select"}
          <Select
            id={key}
            label={config.label ?? key}
            options={config.meta?.options ?? []}
            value={config.meta?.numeric || config.meta?.paramType === "number"
              ? String($specStore[key] ?? config.meta?.default ?? "")
              : ($specStore[key] ?? config.meta?.default)}
            placeholder={config.meta?.placeholder ?? ""}
            full={true}
            size="sm"
            sameWidth
            fontSize={12}
            onChange={(newValue) => {
              // Numeric selects store a number in the YAML, not the option string.
              component.updateProperty(
                key,
                config.meta?.numeric || config.meta?.paramType === "number"
                  ? Number(newValue)
                  : newValue,
              );
            }}
          />

          <!-- SWITCHER TABS -->
        {:else if config.type === "switcher_tab"}
          <CanvasFieldSwitcher
            {key}
            label={config.label ?? key}
            options={config.meta?.options ?? []}
            value={config.meta?.value ??
              localParamValues[key] ??
              config.meta?.default}
            onChange={(newValue) => {
              component.updateProperty(
                key,
                config.meta?.omitDefault && newValue === config.meta?.default
                  ? undefined
                  : newValue,
              );
            }}
          />

          <!-- KPI SPARKLINE INPUT -->
        {:else if config.type === "sparkline"}
          <SparklineInput
            {key}
            label={config.label ?? key}
            value={localParamValues[key]}
            onChange={(updatedSparkline) => {
              localParamValues[key] = updatedSparkline;
              component.updateProperty(key, updatedSparkline);
            }}
          />

          <!-- DATA LABELS INPUT -->
        {:else if config.type === "labels"}
          <LabelsInput
            {key}
            label={config.label ?? key}
            value={localParamValues[key]}
            onChange={(next) => {
              localParamValues[key] = next;
              component.updateProperty(key, next);
            }}
          />

          <!-- COMPARISON OPTIONS INPUT -->
        {:else if config.type === "comparison_options"}
          <ComparisonInput
            {key}
            label={config.label ?? key}
            options={localParamValues[key]}
            onChange={(options) => {
              localParamValues[key] = options;
              component.updateProperty(key, options);
            }}
          />

          <!-- COMPONENT CONTENTS ALIGNMENT -->
        {:else if config.type === "alignment"}
          <AlignmentInput
            {key}
            label={config.label ?? key}
            position={localParamValues[key]}
            defaultAlignment={config.meta?.defaultAlignment}
            onChange={(updatedPosition) => {
              localParamValues[key] = updatedPosition;
              component.updateProperty(key, updatedPosition);
            }}
          />
          <!-- POSITIONAL CONFIG -->
        {:else if metricsView && config.type === "positional"}
          <PositionalFieldConfig
            {canvasName}
            {key}
            {config}
            {metricsView}
            fieldConfig={localParamValues[key] || {}}
            ephemeralMeasures={componentEphemeralMeasures}
            component={supportsEphemeralMeasures ? component : undefined}
            onChange={(updatedConfig) =>
              updatePositionalField(key, updatedConfig, metricsView)}
          />
          <!-- COLOR CONFIG -->
        {:else if metricsView && config.type === "mark"}
          <MarkSelector
            {canvasName}
            {key}
            {config}
            {metricsView}
            markConfig={localParamValues[key] || {}}
            onChange={(updatedConfig) => {
              localParamValues[key] = updatedConfig;
              component.updateProperty(key, updatedConfig);
            }}
          />
          <!-- MAP COLOR CONFIG -->
        {:else if metricsView && config.type === "map_color"}
          <MapColorSelector
            {canvasName}
            {metricsView}
            colorConfig={localParamValues[key]}
            onChange={(updatedConfig) => {
              localParamValues[key] = updatedConfig;
              component.updateProperty(key, updatedConfig);
            }}
          />
          <!-- MAP REGIONS CONFIG -->
        {:else if metricsView && config.type === "map_regions" && component instanceof MapComponent}
          <MapRegionsSelector
            {component}
            {canvasName}
            {metricsView}
            regions={localParamValues[key]}
            onChange={(updatedConfig) => {
              localParamValues[key] = updatedConfig;
              component.updateProperty("regions", updatedConfig);
            }}
          />
        {/if}
        {#if offersControls}
          {@const swappable = swappableField(config, localParamValues[key])}
          {#if swappable}
            <button
              type="button"
              class="viewers-choose"
              onclick={() => letViewersChoose(swappable.kind, swappable.field)}
            >
              {m.canvas_control_let_viewers_choose()}
            </button>
          {:else if slotTargets.has(key) && !controlledSlots.has(key)}
            <button
              type="button"
              class="viewers-choose"
              onclick={() => letViewersPick(key)}
            >
              {m.canvas_control_let_viewers_choose()}
            </button>
          {/if}
        {/if}
      </div>
    {/if}
  {/each}
</div>

<style lang="postcss">
  .component-param {
    @apply py-3 px-5;
    @apply border-t;
  }
  .component-param.grouped {
    @apply py-0;
    @apply border-none;
  }

  .viewers-choose {
    @apply mt-1.5 text-[11px] font-medium text-primary-600;
  }

  .viewers-choose:hover {
    @apply underline;
  }

  .collapsible-header {
    @apply flex items-center gap-1.5 w-full py-1 cursor-pointer;
    @apply text-xs font-semibold text-gray-600;
    @apply bg-transparent border-none text-left;
  }

  .collapsible-header:hover {
    @apply text-gray-800;
  }

  .chevron {
    @apply transition-transform duration-150;
    @apply flex items-center;
  }

  .chevron.expanded {
    transform: rotate(90deg);
  }

  .collapsible-label {
    @apply flex-1;
  }

  .collapsible-badge {
    @apply text-[10px] text-gray-400 bg-gray-100 rounded-full px-1.5 py-0.5;
  }
</style>
