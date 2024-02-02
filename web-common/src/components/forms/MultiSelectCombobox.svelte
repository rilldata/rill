<script lang="ts">
  import { slide } from "svelte/transition";
  import { WithTogglableFloatingElement } from "../floating-element";
  import Check from "../icons/Check.svelte";
  import InfoCircle from "../icons/InfoCircle.svelte";
  import Spacer from "../icons/Spacer.svelte";
  import { Menu, MenuItem } from "../menu";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";

  export let id = "";
  export let label = "";
  export let error: string;
  export let placeholder = "";
  export let options: string[] = [];
  export let selectedValues: string[] = [];
  export let hint = "";

  let inputValue = "";

  $: filteredOptions = options.filter((option) =>
    option.toLowerCase().includes(inputValue.toLowerCase()),
  );

  function toggleOption(option: string) {
    const index = selectedValues.indexOf(option);
    if (index === -1) {
      selectedValues = [...selectedValues, option];
    } else {
      selectedValues = [
        ...selectedValues.slice(0, index),
        ...selectedValues.slice(index + 1),
      ];
    }
  }

  let inputEl: HTMLElement;
</script>

<div class="flex flex-col gap-y-2">
  <div class="flex gap-x-1 items-center">
    <label for={id} class="text-gray-800 text-sm font-medium w-fit">
      {label}
    </label>
    {#if hint}
      <Tooltip location="right" alignment="middle" distance={8}>
        <div class="text-gray-500" style="transform:translateY(-.5px)">
          <InfoCircle size="13px" />
        </div>
        <TooltipContent maxWidth="400px" slot="tooltip-content">
          {hint}
        </TooltipContent>
      </Tooltip>
    {/if}
  </div>
  <WithTogglableFloatingElement
    let:active
    let:toggleFloatingElement
    distance={8}
    alignment="start"
  >
    <input
      bind:this={inputEl}
      {id}
      name={id}
      {placeholder}
      type="text"
      class="bg-white rounded-sm border border-gray-300 px-3 py-[5px] h-8 cursor-pointer focus:outline-primary-500 w-full text-xs {error &&
        'border-red-500'}"
      bind:value={inputValue}
      on:input={() => {
        if (!active) toggleFloatingElement();
      }}
    />
    <Menu
      slot="floating-element"
      let:active
      let:toggleFloatingElement
      focusOnMount={false}
      minWidth={`${inputEl.clientWidth}px`}
      maxHeight="120px"
      on:click-outside={() => {
        if (active) toggleFloatingElement();
      }}
      on:escape={toggleFloatingElement}
    >
      {#if filteredOptions.length > 0}
        {#each filteredOptions as option}
          <MenuItem
            icon
            focusOnMount={false}
            animateSelect={false}
            on:select={() => {
              toggleOption(option);
              toggleFloatingElement();
              inputValue = "";
            }}
          >
            <svelte:fragment slot="icon">
              {#if selectedValues.includes(option)}
                <Check size="20px" color="#15141A" />
              {:else}
                <Spacer size="20px" />
              {/if}
            </svelte:fragment>
            {option}
          </MenuItem>
        {/each}
      {:else}
        <MenuItem focusOnMount={false} disabled>No options</MenuItem>
      {/if}
    </Menu>
  </WithTogglableFloatingElement>
  {#if error}
    <div in:slide={{ duration: 200 }} class="text-red-500 text-sm py-px">
      {error}
    </div>
  {/if}
</div>
