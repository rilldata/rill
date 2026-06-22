<script lang="ts">
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let disabled = false;
  export let selected = false;
  export let href: string | undefined = undefined;
  export let theme = false;
  export let onclick: ((e: MouseEvent) => void) | undefined = undefined;
</script>

<!-- Use snippets when transitioning to Svelte 5 -->
{#if href}
  <a class="relative group" class:theme class:selected {href} {onclick}>
    <slot />
    {#if disabled}
      <div class="disabled group-hover:block">
        <TooltipContent>{m.dashboards_tab_coming_soon()}</TooltipContent>
      </div>
    {/if}
  </a>
{:else}
  <button class="relative group" class:selected {onclick}>
    <slot />
    {#if disabled}
      <div class="disabled group-hover:block">
        <TooltipContent>{m.dashboards_tab_coming_soon()}</TooltipContent>
      </div>
    {/if}
  </button>
{/if}

<style lang="postcss">
  a,
  button {
    @apply border-b-2 border-transparent;
    @apply flex items-center relative cursor-pointer;
    @apply p-1 gap-x-2;
    @apply font-medium text-xs text-fg-muted;
  }

  .disabled {
    @apply absolute top-full translate-y-2 z-10 w-fit;
    @apply -translate-x-1/2 left-1/2 hidden;
  }

  .selected.theme {
    @apply border-b-2 border-theme-600 text-theme-600;
  }

  .selected {
    @apply border-b-2 border-primary-600 text-primary-600;
  }
</style>
