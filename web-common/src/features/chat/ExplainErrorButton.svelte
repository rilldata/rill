<script lang="ts">
  import { SparklesIcon } from "lucide-svelte";
  import { sidebarActions } from "./layouts/sidebar/sidebar-store";
  import { composeErrorPrompt } from "./error-prompt-composer";

  export let errorMessage: string;
  export let filePath: string;
  export let fileContent: string | null | undefined = undefined;
  export let lineNumber: number | undefined = undefined;
  export let additionalErrorCount: number | undefined = undefined;
  export let large = false;

  function handleClick() {
    const prompt = composeErrorPrompt({
      errorMessage,
      filePath,
      fileContent,
      lineNumber,
      additionalErrorCount,
    });
    sidebarActions.startChat(prompt);
  }
</script>

<button
  class={large ? "explain-error-btn large" : "explain-error-btn"}
  on:click|stopPropagation={handleClick}
  aria-label="Explain this error with AI"
  title="Explain this error"
>
  <SparklesIcon size={large ? "16px" : "14px"} />
  <span>Explain this error</span>
</button>

<style lang="postcss">
  .explain-error-btn {
    @apply inline-flex items-center gap-1 px-2 py-1;
    @apply text-[11px] font-medium;
    @apply text-fg-primary;
    @apply bg-surface-muted hover:bg-gray-200 dark:hover:bg-gray-600;
    @apply rounded-[2px] cursor-pointer;
    @apply border border-gray-300 dark:border-gray-500;
    @apply transition-colors duration-150;
    @apply flex-shrink-0;
  }

  .explain-error-btn.large {
    @apply gap-2 px-3;
    @apply h-7 min-h-[28px] text-xs;
  }

  .explain-error-btn:focus-visible {
    @apply outline-none ring-1 ring-gray-400 ring-offset-1;
  }
</style>
