<script lang="ts">
  import { WithTween } from "@rilldata/web-common/components/data-graphic/functional-components";
  import CrossIcon from "@rilldata/web-common/components/icons/CrossIcon.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Spinner from "@rilldata/web-common/features/temp/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/lib/entity";
  import { crossfade, fly } from "svelte/transition";
  import {
    humanizeDataType,
    NicelyFormattedTypes,
  } from "../../../../util/humanize-numbers";

  export let value: number;
  export let status: EntityStatus;
  export let description: string = undefined;
  export let withTimeseries = true;
  export let formatPreset: string; // workaround, since unable to cast `string` to `NicelyFormattedTypes` within MetricsTimeSeriesCharts.svelte's `#each` block

  $: formatPresetEnum =
    (formatPreset as NicelyFormattedTypes) || NicelyFormattedTypes.HUMANIZE;
  $: valusIsPresent = value !== undefined && value !== null;

  const [send, receive] = crossfade({ fallback: fly });
</script>

<div class="flex flex-col {withTimeseries ? 'mt-2' : 'justify-between'}">
  <Tooltip distance={16} location="top">
    <h2
      class="break-words line-clamp-2"
      style:font-size={withTimeseries ? "" : "0.8rem"}
    >
      <slot name="name" />
    </h2>
    <TooltipContent slot="tooltip-content">
      {description}
    </TooltipContent>
  </Tooltip>
  <div
    class="ui-copy-muted"
    style:font-size={withTimeseries ? "1.5rem" : "1.8rem"}
    style:font-weight="light"
  >
    <!-- the default slot will be a tweened number that uses the formatter. One can optionally
    override this by filling the slot in the consuming component. -->
    <slot name="value">
      <div>
        {#if valusIsPresent && status === EntityStatus.Idle}
          <div>
            <WithTween {value} tweenProps={{ duration: 500 }} let:output>
              {#if formatPresetEnum !== NicelyFormattedTypes.NONE}
                {humanizeDataType(output, formatPresetEnum)}
              {:else}
                {output}
              {/if}
            </WithTween>
          </div>
        {:else if status === EntityStatus.Error}
          <CrossIcon />
        {:else if status === EntityStatus.Running}
          <div
            class="absolute p-2"
            in:receive|local={{ key: "spinner" }}
            out:send|local={{ key: "spinner" }}
          >
            <Spinner status={EntityStatus.Running} />
          </div>
        {/if}
      </div>
    </slot>
  </div>
</div>
