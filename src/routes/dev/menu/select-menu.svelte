<script lang="ts">
  import Button from "$lib/components/Button.svelte";
  import List from "$lib/components/icons/List.svelte";

  import SimpleSelectorMenu from "$lib/components/menu/SimpleSelectorMenu.svelte";

  const singleSelectorOptions = [
    { key: 0, main: "option 1", right: "opt1" },
    { key: 1, main: "option 2", right: "opt2" },
    { key: 2, main: "option 3", right: "opt3" },
    { key: 3, main: "option 4", right: "opt4" },
    { key: 4, main: "option 5", right: "opt5" },
  ];
  let singleSelections = [singleSelectorOptions[0]];

  let multipleSelections = [];
  let style = "obvious";
</script>

<section>
  <h2>Select Menus ({style})</h2>

  <button
    on:click={() => {
      style = "obvious";
    }}>obvious</button
  >
  <button
    on:click={() => {
      style = "bare";
    }}>bare</button
  >

  <div class="flex flex-row gap-x-8 content-start items-start">
    <div class="grid grid-flow-row gap-y-4" style:width="600px">
      <div>
        here is a test hm <SimpleSelectorMenu
          {style}
          options={singleSelectorOptions}
          bind:selections={singleSelections}
          let:toggleMenu
          let:active
        />
      </div>

      <div class="pl-4">
        {#each singleSelections as option}
          <div>
            {option.main}
          </div>
        {/each}
      </div>
    </div>

    <div class="grid grid-flow-row gap-y-4">
      <SimpleSelectorMenu
        {style}
        options={singleSelectorOptions}
        bind:selections={singleSelections}
        let:toggleMenu
        let:active
      >
        <Button type="primary" on:click={toggleMenu}
          >single selector (pick one and close menu) <List
            size="16px"
          /></Button
        >
      </SimpleSelectorMenu>

      <div class="pl-4">
        {#each singleSelections as option}
          <div>
            {option.main}
          </div>
        {/each}
      </div>
    </div>

    <div class="grid grid-flow-row gap-y-4">
      <SimpleSelectorMenu
        multiple
        {style}
        options={singleSelectorOptions}
        bind:selections={multipleSelections}
        let:toggleMenu
        let:active
      >
        <Button type="primary" on:click={toggleMenu}
          >multiple selector (pick many and keep menu open) <List
            size="16px"
          /></Button
        >
      </SimpleSelectorMenu>

      <div class="pl-4">
        {#each multipleSelections as option}
          <div>
            {option.main}
          </div>
        {:else}
          <div class="italic text-gray-600">none selected</div>
        {/each}
      </div>
    </div>
  </div>
</section>
