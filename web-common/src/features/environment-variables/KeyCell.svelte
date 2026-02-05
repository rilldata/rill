<script lang="ts">
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let name: string;
  export let subtitle: string | undefined = undefined;

  let copied = false;
  function onCopy() {
    copyToClipboard(name, undefined, false);
    copied = true;

    setTimeout(() => {
      copied = false;
    }, 2_000);
  }
</script>

<div class="truncate flex flex-col">
  <Tooltip distance={6} location="top">
    <button on:click={onCopy} class="truncate text-start" title={name}>
      <span class="source-code text-sm text-fg-primary font-medium truncate">
        {name}
      </span>
    </button>

    <TooltipContent slot="tooltip-content">
      {copied ? "Copied!" : "Click to copy"}
    </TooltipContent>
  </Tooltip>

  {#if subtitle}
    <span class="text-xs text-fg-muted font-normal truncate">
      {subtitle}
    </span>
  {/if}
</div>

<style lang="postcss">
  .source-code {
    font-family: "Source Code Variable", monospace;
  }
</style>
