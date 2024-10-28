<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import CopyIcon from "@rilldata/web-common/components/icons/CopyIcon.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";

  export let message: string;
</script>

{#if message}
  <div class="container">
    <Tooltip distance={8}>
      <span class="error-message">
        {message}
      </span>
      <TooltipContent slot="tooltip-content">
        {message}
      </TooltipContent>
    </Tooltip>
    <IconButton
      rounded
      compact
      on:click={() =>
        copyToClipboard(message, "Copied error message to clipboard")}
    >
      <CopyIcon size="16px" />
    </IconButton>
  </div>
{:else}
  <span>-</span>
{/if}

<style lang="postcss">
  .container {
    @apply flex items-center justify-between gap-2;
  }

  .error-message {
    @apply text-red-600 line-clamp-3;
  }
</style>
