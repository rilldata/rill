<script lang="ts">
  import { onDestroy } from "svelte";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import {
    copyToClipboard,
    isClipboardApiSupported,
  } from "@rilldata/web-common/lib/actions/copy-to-clipboard.ts";
  import { builderActions, getAttrs } from "bits-ui";
  import { TOOLTIP_STRING_LIMIT } from "@rilldata/web-common/layout/config.ts";
  import { modified } from "@rilldata/web-common/lib/actions/modified-click.ts";
  import { cellInspectorStore } from "@rilldata/web-common/features/dashboards/stores/cell-inspector-store.ts";

  export let value: string;
  export let type: "dimension" | "measure" | "comparison";
  export let className: string = "";
  export let background: string = "";

  const HideLeaderboardTooltipAfter = 3000;

  const clipboardSupported =
    typeof navigator !== "undefined" ? isClipboardApiSupported() : false;
  const disabled = !clipboardSupported;

  let tooltipActive = false;
  $: if (tooltipActive) {
    showTemporarily();
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

<Tooltip.Root bind:open={tooltipActive} openDelay={1000}>
  <Tooltip.Trigger asChild let:builder {disabled}>
    <td
      role="button"
      tabindex="0"
      {...getAttrs([builder])}
      use:builderActions={{ builders: [builder] }}
      on:click={modified({
        shift: () => shiftClickHandler(value),
      })}
      on:pointerover={() => {
        if (value) {
          // Always update the value in the store, but don't change visibility
          cellInspectorStore.updateValue(value.toString());
        }
      }}
      on:focus={() => {
        if (value) {
          // Always update the value in the store, but don't change visibility
          cellInspectorStore.updateValue(value.toString());
        }
      }}
      style:background
      class="{type}-cell {className}"
    >
      <slot />
    </td>
  </Tooltip.Trigger>

  {#if clipboardSupported && !disabled}
    <Tooltip.Content class="max-w-[280px] bg-popover-foreground">
      <div>
        <StackingWord key="shift">Copy</StackingWord>
        {value} to clipboard
      </div>
      <Shortcut>
        <span style="font-family: var(--system);">⇧</span> + Click
      </Shortcut>
    </Tooltip.Content>
  {/if}
</Tooltip.Root>

<style lang="postcss">
  td {
    @apply text-right p-0;
    @apply px-2 relative;
    height: 22px;
  }

  /*td.comparison-cell {*/
  /*  @apply bg-surface px-1 truncate;*/
  /*}*/

  /*td.dimension-cell {*/
  /*  @apply sticky left-0 z-30 bg-surface;*/
  /*}*/

  :global(tr:hover .dimension-cell),
  :global(tr:hover .measure-cell),
  :global(tr:hover .comparison-cell) {
    @apply bg-gray-100;
  }
</style>
