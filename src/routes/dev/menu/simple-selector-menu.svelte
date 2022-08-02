<script lang="ts">
  import SimpleSelectorMenu from "$lib/components/menu/SimpleSelectorMenu.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";

  const singleSelectorOptions = [
    { key: 0, main: "option 1", right: "opt1" },
    { key: 1, main: "option 2", right: "opt2" },
    { key: 2, main: "option 3", right: "opt3" },
    {
      key: 3,
      main: "option 4",
      right: "opt4",
      description: "adding a description",
    },
    { key: 4, main: "option 5", right: "opt5" },
  ];
  let singleSelections = [singleSelectorOptions[0]];

  let level = undefined;
  $: if (singleSelections[0]?.description) {
    level = "error";
  } else {
    level = undefined;
  }
</script>

<section class="grid grid-flow-row gap-y-4">
  <h2>Select Menus</h2>

  <div class="grid grid-flow-row gap-y-4" style:width="600px">
    <div>
      Currently selecting
      <Tooltip>
        <SimpleSelectorMenu
          {level}
          options={singleSelectorOptions}
          bind:selections={singleSelections}
        />
        <TooltipContent slot="tooltip-content">
          {#if level} {level} {:else} a simple selector {/if}
        </TooltipContent>
      </Tooltip>
    </div>

    <div class="pl-4">
      {#each singleSelections as option}
        <div>
          {option.main}
        </div>
      {/each}
    </div>
  </div>

  <div>
    <p>
      use the <code>block</code> prop to establish relationship within tables.
    </p>
    <p>
      This utilizes SimpleSelectorMenu with custom styling and slotting in the
      button copy.
    </p>
    <p>Option 4 will trigger the error state.</p>
  </div>
  <table style:width="400px">
    {#each [0, 0] as column}
      <tr>
        {#each [0, 0] as row}
          <td class="border border-gray-200" style:height="32px">
            <SimpleSelectorMenu
              block
              options={singleSelectorOptions}
              bind:selections={singleSelections}
              {level}
            >
              <div class="flex justify-between w-full gap-x-4">
                {singleSelections[0].main}
                <span
                  class="{level === 'error'
                    ? 'text-red-600'
                    : 'text-gray-500'} italic">{singleSelections[0].right}</span
                >
              </div>
            </SimpleSelectorMenu>
          </td>
        {/each}
      </tr>
    {/each}
  </table>
</section>
