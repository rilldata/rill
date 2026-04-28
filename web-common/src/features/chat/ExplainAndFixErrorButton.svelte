<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { SparklesIcon } from "lucide-svelte";
  import { sidebarActions } from "./layouts/sidebar/sidebar-store";

  export let filePath: string;
  export let large = false;
  export let variant: "compact" | "cta" = "compact";

  function handleClick() {
    sidebarActions.startChat(`Fix the errors in \`${filePath}\``);
  }
</script>

{#if variant === "cta"}
  <Button type="secondary" onClick={handleClick}>
    <SparklesIcon size="16px" />
    <span>Explain and fix</span>
  </Button>
{:else}
  <button
    class={large ? "explain-error-btn large" : "explain-error-btn"}
    on:click|stopPropagation={handleClick}
    aria-label="Explain and fix this error with AI"
    title="Explain and fix"
  >
    <SparklesIcon size={large ? "16px" : "14px"} />
    <span>Explain and fix</span>
  </button>
{/if}

<style lang="postcss">
  .explain-error-btn {
    @apply inline-flex items-center gap-1 px-2 py-1;
    @apply text-[11px] font-medium;
    @apply text-fg-primary;
    @apply bg-surface-muted hover:bg-gray-200 dark:hover:bg-gray-600;
    @apply rounded-[2px] cursor-pointer;
    @apply border border-gray-300 dark:border-gray-500;
    @apply transition-colors duration-150;
    @apply flex-shrink-0 self-start;
  }

  .explain-error-btn.large {
    @apply gap-2 px-3;
    @apply h-7 min-h-[28px] text-xs;
  }

  .explain-error-btn:focus-visible {
    @apply outline-none ring-1 ring-gray-400 ring-offset-1;
  }
</style>
