<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { Eye, Pencil, Play, UserRoundSearch, X } from "lucide-svelte";

  export let mode: "Preview" | "Edit" = "Preview";
  export let href: string;
  export let disabled = false;
  /** Email (or label) of the user being impersonated. When set, the left half
   *  swaps from "Preview / Edit" to "Viewing as <label>" and exposes a clear
   *  affordance via the `onClearViewAs` callback. */
  export let activeViewAsLabel: string | null = null;
  export let onClearViewAs: (() => void) | null = null;
  export let dropdownOpen = false;
  /** When true, render the trailing UserRoundSearch dropdown trigger and a
   *  `dropdown` slot for the user-list popover. */
  export let showViewAs = false;

  $: isViewing = activeViewAsLabel !== null;
</script>

<div class="split-button" class:disabled>
  <a
    class="left"
    class:viewing={isViewing}
    {href}
    aria-label={isViewing ? `Viewing as ${activeViewAsLabel}` : mode}
  >
    {#if isViewing}
      <Eye size={14} />
      <span class="truncate">
        Viewing as <span class="font-semibold">{activeViewAsLabel}</span>
      </span>
      {#if onClearViewAs}
        <button
          type="button"
          class="clear-btn"
          aria-label="Clear view"
          on:click|preventDefault|stopPropagation={onClearViewAs}
        >
          <X size={12} />
        </button>
      {/if}
    {:else if mode === "Preview"}
      <Play size={14} />
      <span>Preview</span>
    {:else}
      <Pencil size={14} />
      <span>Edit</span>
    {/if}
  </a>

  {#if showViewAs}
    <DropdownMenu.Root bind:open={dropdownOpen}>
      <DropdownMenu.Trigger>
        {#snippet child({ props })}
          <button
            {...props}
            type="button"
            class="right"
            class:active={dropdownOpen}
            aria-label="View as another user"
          >
            <UserRoundSearch size={14} />
          </button>
        {/snippet}
      </DropdownMenu.Trigger>
      <DropdownMenu.Content
        align="end"
        class="flex flex-col min-w-[220px] max-w-[320px]"
      >
        <slot name="dropdown" />
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  {/if}
</div>

<style lang="postcss">
  .split-button {
    @apply flex items-stretch h-7 rounded-sm border border-primary-500 bg-surface-base shadow-sm overflow-hidden;
  }
  .split-button.disabled {
    @apply opacity-50 pointer-events-none;
  }
  .left {
    @apply flex items-center gap-x-1.5 px-3 text-primary-600 text-xs font-medium;
    @apply hover:bg-primary-50 transition-colors;
  }
  .left.viewing {
    @apply max-w-[280px];
  }
  .right {
    @apply flex items-center justify-center px-2 bg-primary-50 border-l border-primary-500;
    @apply text-primary-600 hover:bg-primary-100 transition-colors;
  }
  .right.active {
    @apply bg-primary-100;
  }
  .clear-btn {
    @apply flex items-center justify-center rounded-full p-0.5 ml-1;
    @apply text-primary-600 hover:bg-primary-100;
  }
</style>
