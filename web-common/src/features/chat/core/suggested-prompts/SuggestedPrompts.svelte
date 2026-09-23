<script lang="ts">
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { V1AIPrompt } from "@rilldata/web-common/runtime-client";
  import { SparklesIcon } from "lucide-svelte";

  export let prompts: V1AIPrompt[];
  export let layout: "sidebar" | "fullpage";

  // ChatInput listens for this event and sends the prompt with the current dashboard context attached.
  function pick(prompt: V1AIPrompt) {
    if (!prompt.prompt) return;
    eventBus.emit("start-chat", prompt.prompt);
  }
</script>

<div
  class="suggested-prompts"
  class:fullpage={layout === "fullpage"}
  role="group"
  aria-label={m.chat_prompts_heading()}
>
  {#each prompts as prompt (prompt.prompt)}
    <button
      type="button"
      class="suggested-prompt"
      title={prompt.prompt}
      on:click={() => pick(prompt)}
    >
      <span class="suggested-prompt-label">
        <SparklesIcon size="14px" class="shrink-0" />
        <span class="truncate">{prompt.label || prompt.prompt}</span>
      </span>
      {#if prompt.label && prompt.label !== prompt.prompt}
        <span class="suggested-prompt-text">{prompt.prompt}</span>
      {/if}
    </button>
  {/each}
</div>

<style lang="postcss">
  .suggested-prompts {
    @apply flex flex-col gap-2 w-full mt-4 text-left;
  }

  .suggested-prompts.fullpage {
    @apply grid grid-cols-1 sm:grid-cols-2 max-w-2xl mt-6;
  }

  .suggested-prompt {
    @apply flex flex-col gap-1 w-full px-3 py-2;
    @apply rounded-md border border-gray-200 dark:border-gray-700;
    @apply bg-surface-base hover:bg-surface-muted;
    @apply text-left transition-colors duration-150;
  }

  .suggested-prompt:focus-visible {
    @apply outline-none ring-1 ring-primary-400;
  }

  .suggested-prompt-label {
    @apply flex items-center gap-1.5;
    @apply text-xs font-medium text-fg-primary;
  }

  .suggested-prompt-text {
    @apply text-xs text-fg-secondary line-clamp-2;
  }
</style>
