<script lang="ts">
  import { onDestroy } from "svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import { isClipboardApiSupported } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import type {
    Alignment,
    Location,
  } from "@rilldata/web-common/lib/place-element";

  export let copyLabel = "this value";
  export let hideAfter = 0;
  export let location: Location = "top";
  export let alignment: Alignment = "middle";
  export let distance = 4;
  export let pad = 8;
  export let suppress = false;
  export let disabled = false;
  export let hoverIntentThreshold = 5;
  export let hoverIntentTimeout = 100;
  export let activeDelay = 200;
  export let hideDelay = 0;

  const clipboardSupported =
    typeof navigator !== "undefined" ? isClipboardApiSupported() : false;

  let tooltipActive = false;
  let suppressTooltip = false;
  let hideTimer: ReturnType<typeof setTimeout> | undefined;

  function clearHideTimer() {
    if (hideTimer) {
      clearTimeout(hideTimer);
      hideTimer = undefined;
    }
  }

  function showTemporarily() {
    if (!clipboardSupported || disabled) return;
    suppressTooltip = false;
    if (hideAfter > 0) {
      clearHideTimer();
      hideTimer = setTimeout(() => {
        suppressTooltip = true;
      }, hideAfter);
    }
  }

  function hideTooltip() {
    clearHideTimer();
    suppressTooltip = false;
  }

  $: if (tooltipActive) {
    showTemporarily();
  } else {
    hideTooltip();
  }

  $: computedSuppress =
    suppress || suppressTooltip || disabled || !clipboardSupported;

  onDestroy(clearHideTimer);
</script>

<Tooltip
  {location}
  {alignment}
  {distance}
  {pad}
  {hoverIntentThreshold}
  {hoverIntentTimeout}
  {activeDelay}
  {hideDelay}
  suppress={computedSuppress}
  bind:active={tooltipActive}
>
  <slot />

  {#if clipboardSupported && !disabled}
    <TooltipContent slot="tooltip-content" maxWidth="280px">
      <TooltipShortcutContainer>
        <div>
          <StackingWord key="shift">Copy</StackingWord>
          {copyLabel} to clipboard
        </div>
        <Shortcut>
          <span style="font-family: var(--system);">⇧</span> + Click
        </Shortcut>
      </TooltipShortcutContainer>
    </TooltipContent>
  {/if}
</Tooltip>
