<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import { cellInspectorStore } from "@rilldata/web-common/features/dashboards/stores/cell-inspector-store.ts";
  import { TOOLTIP_STRING_LIMIT } from "@rilldata/web-common/layout/config.ts";
  import {
    copyToClipboard,
    isClipboardApiSupported,
  } from "@rilldata/web-common/lib/actions/copy-to-clipboard.ts";
  import { modified } from "@rilldata/web-common/lib/actions/modified-click.ts";
  import { Tooltip as TooltipPrimitive } from "bits-ui";
  import { onDestroy } from "svelte";

  export let value: string;
  export let tooltipValue: string = value;
  export let cellType: "dimension" | "measure" | "comparison";
  export let className: string = "";
  export let background: string = "";

  const HideLeaderboardTooltipAfter = 3000;

  const clipboardSupported =
    typeof navigator !== "undefined" ? isClipboardApiSupported() : false;
  const disabled = !clipboardSupported;

  let tooltipActive = false;
  $: if (tooltipActive) {
    showTemporarily();
  } else {
    clearHideTimer();
  }

  let hideTimer: ReturnType<typeof setTimeout> | undefined;

  function clearHideTimer() {
    if (hideTimer) {
      clearTimeout(hideTimer);
      hideTimer = undefined;
    }
  }

  function showTemporarily() {
    if (disabled) return;
    clearHideTimer();
    hideTimer = setTimeout(() => {
      tooltipActive = false;
    }, HideLeaderboardTooltipAfter);
  }

  function shiftClickHandler(label: string) {
    let truncatedLabel = label?.toString();
    if (truncatedLabel?.length > TOOLTIP_STRING_LIMIT) {
      truncatedLabel = `${truncatedLabel.slice(0, TOOLTIP_STRING_LIMIT)}...`;
    }
    copyToClipboard(
      label,
      `copied dimension value "${truncatedLabel}" to clipboard`,
    );
  }

  onDestroy(clearHideTimer);
</script>

<Tooltip.Root
  bind:open={tooltipActive}
  delayDuration={1000}
  disableCloseOnTriggerClick
>
  <TooltipPrimitive.Trigger {disabled}>
    {#snippet child({ props })}
      <td
        {...props}
        role="button"
        tabindex="0"
        onclick={modified({
          shift: () => shiftClickHandler(value),
        })}
        onpointerover={() => cellInspectorStore.updateValue(value)}
        onfocus={() => cellInspectorStore.updateValue(value)}
        onmouseleave={() => (tooltipActive = false)}
        style:background
        class="{cellType}-cell {className}"
      >
        <slot />
      </td>
    {/snippet}
  </TooltipPrimitive.Trigger>

  {#if clipboardSupported && !disabled}
    <Tooltip.Content
      class="flex flex-col max-w-[280px] gap-y-2 p-2 shadow-md bg-tooltip text-fg-inverse"
      sideOffset={16}
    >
      <span class="font-semibold !text-fg-inverse">{tooltipValue}</span>
      <div class="flex flex-row gap-x-6 items-baseline text-fg-disabled">
        <div>
          <StackingWord key="shift">Copy</StackingWord>
          this value to clipboard
        </div>
        <Shortcut>
          <span style="font-family: var(--system);">⇧</span> + Click
        </Shortcut>
      </div>
    </Tooltip.Content>
  {/if}
</Tooltip.Root>

<style lang="postcss">
  td {
    @apply text-right p-0;
    @apply px-2 relative;
    height: 22px;
  }

  td.comparison-cell {
    @apply bg-transparent px-1 truncate;
  }

  td.dimension-cell {
    @apply sticky left-0 z-30 bg-surface-background;
  }

  :global(tr:hover) td.dimension-cell {
    @apply bg-popover-accent;
  }
</style>
