<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters";
  import type { V1Expression } from "@rilldata/web-common/runtime-client";

  let {
    advancedFilter,
    onRemove,
  }: {
    advancedFilter: V1Expression;
    // The pill cannot be edited, so removing it clears the whole filter.
    // Read only surfaces leave this unset.
    onRemove?: () => void;
  } = $props();

  let filterText = $derived(convertExpressionToFilterParam(advancedFilter));
</script>

<div
  class="flex flex-none items-center gap-x-1 px-2 py-[3px] max-h-[26px] border bg-surface-subtle border-gray-200 text-fg-primary rounded-2xl"
>
  {#if onRemove}
    <!-- A sibling of the tooltip trigger, which is a button of its own. -->
    <button
      class="text-inherit"
      aria-label={m.filter_advanced_remove()}
      type="button"
      onpointerdown={(e) => e.stopPropagation()}
      onclick={(e) => {
        e.stopPropagation();
        onRemove();
      }}
    >
      <CancelCircle size="16px" />
    </button>
  {/if}
  <Tooltip.Root>
    <Tooltip.Trigger>
      <span class="font-bold mr-1">{m.filter_advanced_beta()}</span>
      <span>{filterText}</span>
    </Tooltip.Trigger>
    <Tooltip.Content sideOffset={8}>
      {m.filter_advanced_warning()}
    </Tooltip.Content>
  </Tooltip.Root>
</div>
