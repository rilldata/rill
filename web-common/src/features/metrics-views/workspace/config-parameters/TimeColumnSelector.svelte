<script lang="ts">
  import { goto } from "$app/navigation";
  import { IconButton } from "@rilldata/web-common/components/button";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import { SelectMenu } from "@rilldata/web-common/components/menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { selectTimestampColumnFromSchema } from "@rilldata/web-common/features/metrics-views/column-selectors";
  import type { V1Model } from "@rilldata/web-common/runtime-client";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import {
    CONFIG_SELECTOR,
    CONFIG_TOP_LEVEL_LABEL_CLASSES,
    INPUT_ELEMENT_CONTAINER,
    SELECTOR_CONTAINER,
  } from "../styles";
  import FormattedSelectorText from "./FormattedSelectorText.svelte";

  export let metricsInternalRep;
  export let selectedModel: V1Model;

  let metricsConfigErrorStore = getContext(
    "rill:metrics-config:errors"
  ) as Writable<any>;

  $: currentTimestampColumn = $metricsInternalRep.getMetricKey("timeseries");
  $: timeColumnEmpty = currentTimestampColumn === "";
  $: timeColumnIsInModel =
    currentTimestampColumn !== "" &&
    selectedModel?.schema?.fields?.some(
      (field) => field.name === currentTimestampColumn
    );

  $: timeColumnSelectedValue = currentTimestampColumn || "__DEFAULT_VALUE__";

  let timestampColumns: Array<string>;
  $: if (selectedModel) {
    timestampColumns = selectTimestampColumnFromSchema(selectedModel?.schema);
  } else {
    timestampColumns = [];
  }

  function removeTimeseries() {
    $metricsInternalRep.updateMetricsParams({
      timeseries: "",
      smallest_time_grain: "",
      default_time_range: "",
    });
  }

  let tooltipText = "";

  $: modelSelected = selectedModel?.name !== undefined;

  $: if (!modelSelected) {
    tooltipText = "Select a model before selecting a timestamp column";
  } else if (!timeColumnIsInModel) {
    // TODO
  } else if (timestampColumns.length === 0) {
    tooltipText = "The selected model has no timestamp columns";
  } else {
    tooltipText = undefined;
  }

  /** state model
   * - field is empty –
   *   - has column options – show dropdown
   *   - doesn't have column options – disable dropdown and say there isn't a timestamp column
   * - field is not empty –
   *   - if field exists in model, show dropdown with field selected
   *   - if field doesn't exist in model,
   *     - show error state
   *     - if timestamp columns present, enable dropdown
   *     - if no timestamp columns present, dont enable dropdown
   */
  let fieldText: string;
  let level: "error" | undefined = undefined;
  let disabled = false;
  let selectable = true;

  const TOOLTIP_WIDTH = "300px";
  const DEFAULT_TOOLTIP_TEXT =
    "Select a timestamp column to see the time series charts on the dashboard";

  $: if (timeColumnEmpty) {
    if (timestampColumns.length > 0) {
      fieldText = "Select a time column";
      tooltipText = "Select a time column";
      disabled = false;
      level = undefined;
      selectable = true;
    } else {
      fieldText = "No time columns";
      tooltipText = "The selected model has no time columns";
      disabled = true;
      selectable = false;
      level = undefined;
    }
  } else {
    fieldText = currentTimestampColumn;
    if (!timeColumnIsInModel) {
      level = "error";
      fieldText = currentTimestampColumn;
      selectable = false;
      tooltipText =
        "the time column in the configuration is not a time column in this model";
    } else {
      level = undefined;
      disabled = false;
      selectable = true;
      tooltipText = DEFAULT_TOOLTIP_TEXT;
    }
  }

  $: metricsConfigErrorStore.update((errors) => {
    errors.timeColumn = level === "error" ? tooltipText : null;
    return errors;
  });

  /** combine options.*/
  $: options = [
    ...(!timeColumnIsInModel && !timeColumnEmpty
      ? [
          {
            key: currentTimestampColumn,
            description: "not in model",
            main: fieldText,
            divider: true,
          },
        ]
      : []),
    // actual existing timestamp options
    ...(timestampColumns.map((columnName) => {
      return {
        key: columnName,
        main: columnName,
      };
    }) || []),
  ];

  let active = false;
</script>

<div
  class:hidden={!modelSelected}
  class={INPUT_ELEMENT_CONTAINER.classes}
  style={INPUT_ELEMENT_CONTAINER.style}
>
  <Tooltip alignment="start" distance={16} location="bottom">
    <div class={CONFIG_TOP_LEVEL_LABEL_CLASSES}>Timestamp</div>
    <TooltipContent maxWidth={TOOLTIP_WIDTH} slot="tooltip-content">
      {DEFAULT_TOOLTIP_TEXT}
    </TooltipContent>
  </Tooltip>
  <div class={SELECTOR_CONTAINER.classes} style={SELECTOR_CONTAINER.style}>
    <Tooltip alignment="start" distance={8} location="bottom" suppress={active}>
      <SelectMenu
        bind:active
        block
        paddingTop={1}
        paddingBottom={1}
        {options}
        {disabled}
        selection={timeColumnSelectedValue}
        tailwindClasses="{CONFIG_SELECTOR.base} {level === 'error'
          ? CONFIG_SELECTOR.error
          : CONFIG_SELECTOR.info}"
        activeTailwindClasses={level === "error"
          ? CONFIG_SELECTOR.activeError
          : CONFIG_SELECTOR.active}
        distance={CONFIG_SELECTOR.distance}
        alignment="start"
        on:select={(evt) => {
          $metricsInternalRep.updateMetricsParams({
            timeseries: evt.detail?.key,
          });
        }}
      >
        <FormattedSelectorText
          value={fieldText}
          selected={timeColumnSelectedValue !== "__DEFAULT_VALUE__" &&
            selectable}
        />
      </SelectMenu>

      <TooltipContent maxWidth={TOOLTIP_WIDTH} slot="tooltip-content">
        {tooltipText}
      </TooltipContent>
    </Tooltip>

    <Tooltip location="right" distance={8} suppress={active}>
      <IconButton
        ariaLabel="Remove timestamp column"
        compact
        rounded
        marginClasses="ml-1"
        on:click={() => {
          if (timeColumnSelectedValue !== "__DEFAULT_VALUE__")
            removeTimeseries();
          else if (timeColumnEmpty && !timestampColumns?.length)
            goto(`/model/${selectedModel.name}`);
          else active = true;
        }}
      >
        <!-- <CancelCircle color="gray" size="16px" /> -->
        <span class="text-gray-600">
          {#if !timeColumnEmpty}
            <span class={level === "error" ? "text-red-800" : ""}>
              <CancelCircle size="16px" />
            </span>
          {:else}
            <InfoCircle size="16px" />
          {/if}
        </span>
      </IconButton>
      <TooltipContent maxWidth={TOOLTIP_WIDTH} slot="tooltip-content">
        {#if !timeColumnEmpty}
          remove the selected timestamp column
        {:else if timestampColumns?.length}
          select a timestamp column
        {:else}
          go to the model and create a timestamp column
        {/if}
      </TooltipContent>
    </Tooltip>
  </div>
</div>
