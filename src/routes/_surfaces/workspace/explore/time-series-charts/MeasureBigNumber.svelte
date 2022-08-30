<script lang="ts">
  import { EntityStatus } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { WithTween } from "$lib/components/data-graphic/functional-components";
  import CrossIcon from "$lib/components/icons/CrossIcon.svelte";
  import Spinner from "$lib/components/Spinner.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import {
    humanizeDataType,
    NicelyFormattedTypes,
  } from "$lib/util/humanize-numbers";
  import { crossfade, fly } from "svelte/transition";

  export let value: number;
  export let status: EntityStatus;
  export let description: string = undefined;
  export let formatPreset: NicelyFormattedTypes;

  $: valusIsPresent = value !== undefined && value !== null;

  const [send, receive] = crossfade({ fallback: fly });
</script>

<div>
  <Tooltip location="top" distance={16}>
    <h2>
      <slot name="name" />
    </h2>
    <TooltipContent slot="tooltip-content">
      {description}
    </TooltipContent>
  </Tooltip>
  <div style:font-size="1.5rem" style:font-weight="light" class="text-gray-600">
    <!-- the default slot will be a tweened number that uses the formatter. One can optionally
    override this by filling the slot in the consuming component. -->
    <slot name="value">
      <div>
        {#if valusIsPresent && status === EntityStatus.Idle}
          <div>
            <WithTween {value} tweenProps={{ duration: 500 }} let:output>
              {#if formatPreset !== NicelyFormattedTypes.NONE}
                {humanizeDataType(output, formatPreset)}
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
