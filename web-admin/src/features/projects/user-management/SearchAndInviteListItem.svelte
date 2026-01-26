<script lang="ts">
  import Avatar from "@rilldata/web-common/components/avatar/Avatar.svelte";
  import { Chip } from "@rilldata/web-common/components/chip";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { getRandomBgColor } from "@rilldata/web-common/features/themes/color-config";

  export let result: any;
  export let resultIndex: number;
  export let isSelected: boolean;
  export let highlightedIndex: number;
  export let keyboardNavigationActive: boolean;
  export let onSelect: (result: any) => void;
  export let onHighlight: (index: number) => void;
  export let onClearHighlight: () => void;

  function getInitials(name: string) {
    return name.charAt(0).toUpperCase();
  }
</script>

<button
  type="button"
  class:highlighted={resultIndex === highlightedIndex}
  class:selected={isSelected}
  class="dropdown-item"
  on:click={(e) => {
    e.preventDefault();
    onSelect(result);
  }}
  on:keydown={(e) => {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      onSelect(result);
    }
  }}
  on:pointerdown={(e) => {
    e.preventDefault();
  }}
  on:pointerenter={() => {
    if (!keyboardNavigationActive) {
      onHighlight(resultIndex);
    }
  }}
  on:pointerleave={() => {
    if (!keyboardNavigationActive) {
      onClearHighlight();
    }
  }}
>
  <div class="flex items-center gap-2">
    {#if result.type === "group"}
      <div
        class={cn(
          "h-7 w-7 rounded-sm flex items-center justify-center",
          getRandomBgColor(result.identifier),
        )}
      >
        <span class="text-sm text-fg-primary font-semibold"
          >{getInitials(result.identifier)}</span
        >
      </div>
      <div class="flex flex-col text-left">
        <span class="text-sm font-medium text-fg-primary"
          >{result.identifier}</span
        >
        {#if result.groupCount !== undefined}
          <span class="text-xs text-fg-secondary">
            {result.groupCount} user{result.groupCount > 1 ? "s" : ""}
          </span>
        {/if}
      </div>
    {:else}
      <Avatar
        avatarSize="h-7 w-7"
        fontSize="text-xs"
        src={result.photoUrl}
        alt={result.invitedBy ? undefined : result.name}
        bgColor={getRandomBgColor(result.identifier)}
      />
      <div class="flex flex-col text-left">
        {#if result.type === "user" && result.orgRoleName === OrgUserRoles.Guest}
          <span
            class="text-sm font-medium text-fg-primary flex flex-row items-center gap-x-1"
          >
            {result.identifier}
            <Chip type="amber" label="Guest" compact readOnly>
              <svelte:fragment slot="body">Guest</svelte:fragment>
            </Chip>
          </span>
        {:else}
          <span class="text-sm font-medium text-fg-primary">
            {result.identifier}
          </span>
        {/if}
        <span class="text-xs text-fg-secondary"
          >{result.invitedBy ? "Pending invitation" : result.name}</span
        >
      </div>
    {/if}
  </div>
  {#if isSelected}
    <Check size="16px" className="text-fg-secondary" />
  {/if}
</button>

<style lang="postcss">
  .dropdown-item {
    @apply flex items-center justify-between px-3 py-2 cursor-pointer w-full text-left border-none bg-transparent;
    scroll-margin: 8px;
    transition: background-color 150ms ease-in-out;
  }

  .dropdown-item:hover {
    @apply bg-surface-hover;
  }

  .dropdown-item.highlighted {
    @apply bg-surface-hover;
    scroll-snap-align: start;
  }

  .dropdown-item.selected {
    @apply bg-surface-active;
  }

  .dropdown-item.selected:hover {
    @apply bg-surface-hover;
  }
</style>
