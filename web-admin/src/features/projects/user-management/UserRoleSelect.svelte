<script lang="ts">
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { PROJECT_ROLES_OPTIONS } from "../constants";

  export let value: string;
  export let width = "w-18";

  let open = false;

  function onSelect(val: string) {
    value = val;
  }

  $: selected = PROJECT_ROLES_OPTIONS.find((o) => o.value === value);
</script>

<DropdownMenu bind:open typeahead={false}>
  <DropdownMenuTrigger
    class="{width} flex flex-row gap-1 items-center rounded-sm {open
      ? 'bg-surface-active'
      : 'hover:bg-surface-hover'} px-2 py-1"
  >
    <div class="text-xs">{selected?.label ?? ""}</div>
    {#if open}
      <CaretUpIcon size="12px" />
    {:else}
      <CaretDownIcon size="12px" />
    {/if}
  </DropdownMenuTrigger>
  <DropdownMenuContent
    side="bottom"
    align="end"
    class="w-[240px]"
    strategy="fixed"
  >
    {#each PROJECT_ROLES_OPTIONS as { value, label, description } (value)}
      <DropdownMenuItem
        on:click={() => onSelect(value)}
        class="text-xs hover:bg-surface-hover {selected?.value ===
        value
          ? 'bg-surface-active'
          : ''}"
      >
        <div class="flex flex-col">
          <div class="text-xs font-medium text-fg-primary">{label}</div>
          <div class="text-fg-secondary text-[11px]">{description}</div>
        </div>
      </DropdownMenuItem>
    {/each}
  </DropdownMenuContent>
</DropdownMenu>
