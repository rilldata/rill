<script lang="ts">
  import { onDestroy } from "svelte";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import { isClipboardApiSupported } from "@rilldata/web-common/lib/actions/copy-to-clipboard.ts";

  export let copyLabel = "this value";
  export let hideAfter = 3000;

  const clipboardSupported =
    typeof navigator !== "undefined" ? isClipboardApiSupported() : false;
  const disabled = !clipboardSupported;

  let tooltipActive = false;
  let hideTimer: ReturnType<typeof setTimeout> | undefined;

  function clearHideTimer() {
    if (hideTimer) {
      clearTimeout(hideTimer);
      hideTimer = undefined;
    }
  }

  function showTemporarily() {
    if (disabled) return;
    if (hideAfter > 0) {
      clearHideTimer();
      hideTimer = setTimeout(() => {
        tooltipActive = false;
      }, hideAfter);
    }
  }

  $: if (tooltipActive) {
    showTemporarily();
  }

  onDestroy(clearHideTimer);
</script>

<Tooltip.Root bind:open={tooltipActive}>
  <Tooltip.Trigger asChild let:builder {disabled}>
    <slot {builder} />
  </Tooltip.Trigger>

  {#if clipboardSupported && !disabled}
    <Tooltip.Content class="max-w-[280px] bg-popover-foreground">
      <div>
        <StackingWord key="shift">Copy</StackingWord>
        {copyLabel} to clipboard
      </div>
      <Shortcut>
        <span style="font-family: var(--system);">⇧</span> + Click
      </Shortcut>
    </Tooltip.Content>
  {/if}
</Tooltip.Root>
