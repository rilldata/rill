<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import {
    DropdownMenuGroup,
    DropdownMenuItem,
    DropdownMenuLabel,
  } from "@rilldata/web-common/components/dropdown-menu";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let dimension: string;
  export let values: any[];
  export let onSelect: (dimension: string, value: any) => void;

  const SHORT_LIST_COUNT = 10;
  const FULL_LIST_COUNT = 100;

  $: count = values.length;
  $: countLabel = count >= FULL_LIST_COUNT ? `${count}+` : `${count}`;
  let expanded = false;
  $: showExpand = count > SHORT_LIST_COUNT;

  $: shownValues = expanded ? values : values.slice(0, SHORT_LIST_COUNT);
</script>

<DropdownMenuGroup>
  <DropdownMenuLabel class="flex flex-col text-fg-secondary">
    <div class="font-semibold text-[10px] h-4">{dimension.toUpperCase()}</div>
    <div class="font-normal text-[11px] h-4">
      {m.dashboards_dim_search_results({ label: countLabel, n: count })}
    </div>
  </DropdownMenuLabel>
  <div class="flex flex-col">
    {#each shownValues as value}
      <DropdownMenuItem
        class="text-xs"
        onclick={() => onSelect(dimension, value)}
      >
        {value}
      </DropdownMenuItem>
    {/each}
    {#if showExpand}
      <Button
        type="link"
        noStroke
        onClick={() => (expanded = !expanded)}
        class="justify-items-start"
      >
        {expanded
          ? m.dashboards_dim_search_see_less()
          : m.dashboards_dim_search_see_more()}
      </Button>
    {/if}
  </div>
</DropdownMenuGroup>
