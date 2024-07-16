<script lang="ts">
  import { WithTogglableFloatingElement } from "../floating-element";
  import Check from "../icons/Check.svelte";
  import InfoCircle from "../icons/InfoCircle.svelte";
  import Spacer from "../icons/Spacer.svelte";
  import { Menu, MenuItem } from "../menu";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import { createEventDispatcher } from "svelte";

  export let id = "";
  export let label = "";
  export let placeholder = "";
  export let options: { value: string; label?: string }[];
  export let selectValues: { value: string; label?: string }[] = [];
  export let hint = "";
  export let readonly = true;

  const dispatch = createEventDispatcher();

  function toggleOption(option: { value: string; label?: string }) {
    const index = selectValues.findIndex(
      (selectedOption) => selectedOption.value === option.value,
    );

    if (index === -1) {
      selectValues.push(option);
    } else {
      selectValues.splice(index, 1);
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
      class="bg-white rounded-sm border border-gray-300 px-3 py-[5px] h-8 cursor-pointer focus:outline-primary-500 w-full text-xs"
      {readonly}
      on:click={() => {
        toggleFloatingElement();
      }}
      value={selectValues && selectValues.length
        ? selectValues.map((m) => (m.label ? m.label : m.value)).join(", ")
        : null}
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
      {#if options.length > 0}
        {#each options as option}
          <MenuItem
            icon
            focusOnMount={false}
            animateSelect={false}
            on:select={() => {
              toggleOption(option);
              dispatch("change", selectValues);
              toggleFloatingElement();
            }}
          >
            <svelte:fragment slot="icon">
              {#if selectValues.find((val) => val.value === option.value)}
                <Check size="20px" color="#15141A" />
              {:else}
                <Spacer size="20px" />
              {/if}
            </svelte:fragment>
            {option.label || option.value}
          </MenuItem>
        {/each}
      {:else}
        <MenuItem focusOnMount={false} disabled>No options</MenuItem>
      {/if}
    </Menu>
  </WithTogglableFloatingElement>
</div>
