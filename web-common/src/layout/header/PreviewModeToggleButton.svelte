<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { Pencil, Play, UserRoundSearch } from "lucide-svelte";

  export let mode: "Preview" | "Edit" = "Preview";
  export let href: string;
  export let disabled = false;
  export let dropdownOpen = false;
  /** When true, render the trailing UserRoundSearch dropdown trigger and a
   *  `dropdown` slot for the user-list popover. */
  export let showViewAs = false;
  /** Fires alongside navigation when the user clicks the Preview/Edit
   *  half. Use this to reset session state (chat, impersonation, etc.)
   *  on an explicit mode swap — distinct from picking a user from the
   *  dropdown, which intentionally preserves impersonation. */
  export let onPreviewClick: (() => void) | null = null;
</script>

<div class="split-button" class:disabled>
  <a class="left" {href} aria-label={mode} on:click={() => onPreviewClick?.()}>
    {#if mode === "Preview"}
      <Play size={14} />
    {:else}
      <Pencil size={14} />
    {/if}
    <span>{mode}</span>
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
    @apply flex items-stretch h-7 border border-primary-500 bg-surface-base shadow-sm;
  }
  .split-button.disabled {
    @apply opacity-50 pointer-events-none;
  }
  .left {
    @apply flex items-center gap-x-1.5 px-3 text-primary-600 text-xs font-medium;
    @apply hover:bg-primary-50 transition-colors;
  }
  .right {
    @apply flex items-center justify-center px-2 border-l border-primary-500;
    @apply text-primary-600 hover:bg-primary-50 transition-colors;
  }
  .right.active {
    @apply bg-primary-50;
  }
</style>
