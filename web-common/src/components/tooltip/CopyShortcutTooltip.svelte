<script lang="ts">
  import { onDestroy } from "svelte";
  import Tooltip from "./Tooltip.svelte";
  import TooltipContent from "./TooltipContent.svelte";
  import TooltipShortcutContainer from "./TooltipShortcutContainer.svelte";
  import TooltipTitle from "./TooltipTitle.svelte";
  import Shortcut from "./Shortcut.svelte";
  import StackingWord from "./StackingWord.svelte";
  import FormattedDataType from "../data-types/FormattedDataType.svelte";
  import { isClipboardApiSupported } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import type { Location } from "@rilldata/web-common/lib/place-element";

  export let value:
    | string
    | number
    | boolean
    | null
    | undefined
    | Record<string, unknown>;
  export let type = "VARCHAR";
  export let label = "this value";
  export let location: Location = "top";
  export let distance = 16;
  export let hideAfter = 0;
  export let maxWidth = "360px";
  export let truncate = false;
  export let disabled = false;

  const clipboardSupported = isClipboardApiSupported();

  let suppressTooltip = true;
  let hideTimer: ReturnType<typeof setTimeout> | undefined;
  let hasTooltipTitleSlot = false;

  // Svelte injects $$slots at compile time.
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  // @ts-ignore
  $: hasTooltipTitleSlot = !!$$slots["tooltip-title"];

  function clearHideTimer() {
    if (!hideTimer) return;
    clearTimeout(hideTimer);
    hideTimer = undefined;
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
    suppressTooltip = true;
  }

  onDestroy(() => {
    clearHideTimer();
  });
</script>

<Tooltip
  {location}
  {distance}
  activeDelay={0}
  hoverIntentTimeout={0}
  suppress={!clipboardSupported || disabled || suppressTooltip}
>
  <div
    class="contents"
    on:pointerenter={showTemporarily}
    on:pointerleave={hideTooltip}
    on:focus={showTemporarily}
    on:blur={hideTooltip}
  >
    <slot />
  </div>
  <TooltipContent slot="tooltip-content" {maxWidth}>
    {#if hasTooltipTitleSlot}
      <slot name="tooltip-title" />
    {:else if value !== undefined}
      <TooltipTitle>
        <FormattedDataType slot="name" {type} {value} {truncate} />
      </TooltipTitle>
    {/if}
    <TooltipShortcutContainer>
      <div>
        <StackingWord key="shift">Copy</StackingWord> {label} to clipboard
      </div>
      <Shortcut>
        <span style="font-family: var(--system);">⇧</span> + Click
      </Shortcut>
    </TooltipShortcutContainer>
  </TooltipContent>
</Tooltip>
