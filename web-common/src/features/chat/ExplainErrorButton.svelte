<script lang="ts">
  import { SparklesIcon } from "lucide-svelte";
  import { sidebarActions } from "./layouts/sidebar/sidebar-store";
  import { composeErrorPrompt } from "./error-prompt-composer";

  export let errorMessage: string;
  export let filePath: string;
  export let fileContent: string | null | undefined = undefined;
  export let lineNumber: number | undefined = undefined;
  export let additionalErrorCount: number | undefined = undefined;

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
  class="explain-error-btn"
  on:click|stopPropagation={handleClick}
  aria-label="Explain this error with AI"
  title="Explain this error"
>
  <SparklesIcon size="12px" />
  <span>Explain this error</span>
</button>

<style lang="postcss">
  .explain-error-btn {
    @apply inline-flex items-center gap-1 px-2 py-0.5;
    @apply text-[11px] font-medium;
    @apply text-accent-primary-action hover:text-fg-accent;
    @apply bg-transparent hover:bg-surface-hover;
    @apply rounded-sm cursor-pointer;
    @apply border border-transparent hover:border-accent-primary-action/30;
    @apply transition-colors duration-150;
    @apply flex-shrink-0;
  }

  .explain-error-btn:focus-visible {
    @apply outline-none ring-1 ring-accent-primary-action ring-offset-1;
  }
</style>
