<script lang="ts">
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";

  export let value: string;

  let open = false;

  const Options = [
    { value: "admin", label: "Admin" },
    { value: "editor", label: "Editor" },
    { value: "viewer", label: "Viewer" },
  ];
  function onSelect(val: string) {
    value = val;
  }
  $: selected = Options.find((o) => o.value === value);
</script>

<DropdownMenu bind:open typeahead={false}>
  <DropdownMenuTrigger
    class="w-18 flex flex-row gap-1 items-center rounded-sm {open
      ? 'bg-slate-200'
      : 'hover:bg-slate-100'} px-2 py-1"
  >
    <div class="text-xs">{selected?.label ?? ""}</div>
    {#if open}
      <CaretUpIcon size="12px" />
    {:else}
      <CaretDownIcon size="12px" />
    {/if}
  </DropdownMenuTrigger>
  <DropdownMenuContent side="bottom" align="end">
    {#each Options as { value, label } (value)}
      <DropdownMenuItem on:click={() => onSelect(value)} class="text-xs">
        {label}
      </DropdownMenuItem>
    {/each}
  </DropdownMenuContent>
</DropdownMenu>
