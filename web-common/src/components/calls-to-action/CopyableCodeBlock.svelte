<script lang="ts">
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import CopyIcon from "@rilldata/web-common/components/icons/CopyIcon.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";

  export let code: string;
  export let message: string = "Copied to clipboard";

  let copied = false;

  function handleCopy() {
    copyToClipboard(code, message, false);
    copied = true;
    setTimeout(() => (copied = false), 2000);
  }
</script>

<button class="command-box" title={code} on:click={handleCopy}>
  <code class="text-xs truncate">{code}</code>
  <span class="text-fg-muted shrink-0">
    {#if copied}
      <Check size="14px" color="#22c55e" />
    {:else}
      <CopyIcon size="14px" />
    {/if}
  </span>
</button>

<style lang="postcss">
  .command-box {
    @apply flex items-center gap-x-2;
    @apply bg-surface-subtle border border-gray-200 rounded px-2 py-1;
    @apply font-mono text-fg-primary text-left;
    @apply cursor-pointer;
  }

  .command-box:hover {
    @apply bg-surface-hover;
  }
</style>
