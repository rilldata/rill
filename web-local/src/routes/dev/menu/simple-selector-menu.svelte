<script lang="ts">
  import { SelectMenu } from "@rilldata/web-common/components/menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  const singleSelectorOptions = [
    { key: 0, main: "option 1", right: "opt1" },
    { key: 1, main: "option 2", right: "opt2" },
    { key: 2, main: "option 3", right: "opt3" },
    {
      key: 3,
      main: "option 4",
      right: "opt4",
      description: "triggers error state",
    },
    { key: 4, main: "option 5", right: "opt5" },
  ];
  let singleSelection = singleSelectorOptions[0];

  let level = undefined;
  $: if (singleSelection?.description) {
    level = "error";
  } else {
    level = undefined;
  }

  let dark = false;
</script>

<section class="grid grid-flow-row gap-y-4">
  <h1 class="text-lg">Select Menus</h1>

  <div>
    <button
      class="px-2 py-1 rounded {!dark && 'bg-gray-100'}"
      on:click={() => {
        dark = false;
      }}
    >
      light
    </button>
    <button
      class="px-2 py-1 rounded {dark && 'bg-gray-100'}"
      on:click={() => {
        dark = true;
      }}
    >
      dark
    </button>
  </div>

  <p>
    the <code>SelectMenu</code> component takes care of the basic cases covered
    by the {`<select>`}
    element. It has a few additional bells and whistles:
  </p>

  <ul>
    <li>you can add descriptions & right-text to the menu items</li>
    <li>you can change the styling of the overall element</li>
    <li>it's easy to make a block-level select element</li>
    <li>
      tooltips attached to the top-level element will suppress when active
    </li>
    <li>
      you can add a level prop, which currently accepts "error" and turns the
      element red
    </li>
  </ul>

  <h2 class="text-md">basic inline select</h2>

  <div class="grid grid-flow-row gap-y-4" style:width="600px">
    <div>
      Currently selecting
      <Tooltip distance={16}>
        <SelectMenu
          {dark}
          {level}
          options={singleSelectorOptions}
          bind:selection={singleSelection}
        />
        <TooltipContent slot="tooltip-content">
          {#if level}
            {level}
          {:else}
            a simple selector. This will disappear on click
          {/if}
        </TooltipContent>
      </Tooltip>
    </div>

    <div class="pl-4">
      <div>
        {singleSelection?.main}
      </div>
    </div>
  </div>

  <h2 class="text-md">
    using the block prop for tables (along with custom slot value for menu text)
  </h2>

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
            <SelectMenu
              block
              {dark}
              options={singleSelectorOptions}
              bind:selection={singleSelection}
              {level}
            >
              <div class="flex justify-between w-full gap-x-4">
                {singleSelection?.main}
                <span
                  class={level === "error" ? "text-red-600" : "text-gray-500"}
                  >{singleSelection?.right}</span
                >
              </div>
            </SelectMenu>
          </td>
        {/each}
      </tr>
    {/each}
  </table>
</section>
