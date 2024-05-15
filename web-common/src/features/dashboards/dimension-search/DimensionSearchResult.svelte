<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import {
    DropdownMenuGroup,
    DropdownMenuLabel,
    DropdownMenuItem,
  } from "@rilldata/web-common/components/dropdown-menu";

  export let dimension: string;
  export let values: any[];
  export let onSelect: (dimension: string, value: any) => void;

  const SHORT_LIST_COUNT = 10;
  const FULL_LIST_COUNT = 100;

  $: count = values.length;
  let expanded = false;
  $: showExpand = count > SHORT_LIST_COUNT;

  $: shownValues = expanded ? values : values.slice(0, SHORT_LIST_COUNT);
</script>

<DropdownMenuGroup>
  <DropdownMenuLabel class="flex flex-col text-gray-500 text-xs">
    <div class="font-semibold">{dimension.toUpperCase()}</div>
    <div>{count}{count >= FULL_LIST_COUNT ? "+" : ""} results</div>
  </DropdownMenuLabel>
  <div class="flex flex-col">
    {#each shownValues as value}
      <DropdownMenuItem
        class="text-xs"
        on:click={() => onSelect(dimension, value)}
      >
        {value}
      </DropdownMenuItem>
    {/each}
    {#if showExpand}
      <Button
        type="link"
        noStroke
        on:click={() => (expanded = !expanded)}
        class="justify-items-start"
      >
        {expanded ? "See less" : "See more"}
      </Button>
    {/if}
  </div>
</DropdownMenuGroup>
