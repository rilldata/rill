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
    { value: "viewer", label: "Viewers" },
    { value: "admin", label: "Admins" },
  ];
  function onSelect(val: string) {
    value = val;
  }
  $: selected = Options.find((o) => o.value === value);
</script>

<DropdownMenu bind:open typeahead={false}>
  <DropdownMenuTrigger class="w-16 flex flex-row items-center">
    <div>{selected?.label ?? ""}</div>
    {#if open}
      <CaretUpIcon size="16px" />
    {:else}
      <CaretDownIcon size="16px" />
    {/if}
  </DropdownMenuTrigger>
  <DropdownMenuContent>
    {#each Options as { value, label } (value)}
      <DropdownMenuItem on:click={() => onSelect(value)}>
        {label}
      </DropdownMenuItem>
    {/each}
  </DropdownMenuContent>
</DropdownMenu>
