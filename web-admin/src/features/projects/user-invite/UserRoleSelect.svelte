<script lang="ts">
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";

  export let value: string;

  let open = false;

  const Options = [
    { value: "viewer", label: "can view" },
    { value: "admin", label: "can edit" },
  ];
  function onSelect(val: string) {
    value = val;
  }
  $: selected = Options.find((o) => o.value === value);
</script>

<DropdownMenu bind:open typeahead={false}>
  <DropdownMenuTrigger class="w-16">
    {selected?.label ?? ""}
  </DropdownMenuTrigger>
  <DropdownMenuContent>
    {#each Options as { value, label } (value)}
      <DropdownMenuItem on:click={() => onSelect(value)}>
        {label}
      </DropdownMenuItem>
    {/each}
  </DropdownMenuContent>
</DropdownMenu>
