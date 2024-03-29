<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import type { SelectMenuItem } from "../types";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";

  export let options: SelectMenuItem[];
  export let selections: Array<string | number>;

  const dispatch = createEventDispatcher();

  function handleClick(index: number) {
    // if (!multiSelect) selection = event.detail;
    const selection = options?.[index];

    if (!options) return;

    dispatch("select", selection);
  }
</script>

<DropdownMenu.Content class="min-w-44 max-h-96 overflow-y-scroll" align="start">
  {#each options as option, i (option.key)}
    {@const selected = selections.includes(option.key)}
    <DropdownMenu.CheckboxItem
      role="menuitem"
      disabled={option.disabled}
      class="text-xs cursor-pointer rounded-none"
      checked={selected}
      on:click={() => handleClick(i)}
    >
      <div class="flex flex-col">
        <div class:text-gray-400={option.disabled} class:font-bold={selected}>
          {option.main}
        </div>
        {#if option.description}
          <p class="ui-copy-muted" style:font-size="11px">
            {option.description}
          </p>
        {/if}
      </div>
      {option.right || ""}
    </DropdownMenu.CheckboxItem>
    {#if option.divider}
      <DropdownMenu.Separator />
    {/if}
  {/each}
</DropdownMenu.Content>
