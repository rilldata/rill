<script lang="ts">
  import { Copy } from "lucide-svelte";
  import CrossIcon from "@rilldata/web-common/components/icons/CrossIcon.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";

  export let error: string | undefined;
</script>

<div class="component-error">
  <div class="error-icon">
    <CrossIcon size="12px" />
  </div>
  <div class="error-message">
    {error ?? "Component Error"}
  </div>
  {#if error}
    <button
      class="p-2 rounded hover:bg-destructive/50 transition-colors"
      on:click={() => copyToClipboard(error, "Copied error to clipboard")}
      title="Copy error to clipboard"
    >
      <Copy size="14px" />
    </button>
  {/if}
</div>

<style lang="postcss">
  .component-error {
    @apply flex flex-col gap-y-3 w-full h-full items-center justify-center p-4;
    @apply text-fg-secondary bg-destructive/30;
  }

  .error-icon {
    @apply p-2 bg-destructive/60 rounded-full text-fg-inverse;
  }

  .error-message {
    @apply text-base font-medium text-center max-w-md;
  }
</style>
