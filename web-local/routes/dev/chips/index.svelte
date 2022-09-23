<script>
  import { ChipContainer, RemovableListChip } from "$lib/components/chip";
  import Add from "$lib/components/icons/Add.svelte";
  import Cancel from "$lib/components/icons/Cancel.svelte";
  import { flip } from "svelte/animate";

  const chipValues = [
    {
      name: "Country",
      values: ["us", "de", "ca", "mx"],
      typeLabel: "dimension",
    },
    {
      name: "OS",
      values: ["mac", "linux", "windows"],
      typeLabel: "environment",
    },
    {
      name: "Architecture",
      values: ["arm-64", "x64", "other"],
      typeLabel: "build",
    },
    {
      name: "Very Long Strings",
      typeLabel: "outcome",
      values: [
        "The agency found that there was no problem to solve, so they're not going to do anything.",
        "This issue was addressed by the agency.",
        "A third very long string value. These should wrap in the popover, and be truncated in the filterable pill.",
      ],
    },
  ];

  let actives;
  function resetActives() {
    actives = chipValues.reduce((obj, v) => {
      obj[v.name] = v.values;
      return obj;
    }, {});
  }
  resetActives();

  function toggleActiveValue(name, value) {
    if (actives[name].includes(value)) {
      actives[name] = [...actives[name].filter((v) => v !== value)];
    } else {
      actives[name] = [...actives[name], value];
    }
  }

  let activeChips = [...chipValues];
</script>

<h1 class="text-xl">Chips</h1>

<p class="mb-4 mt-2">
  This route contains basic "removable list" chips, which are used in the
  dashboard as filters.
</p>

<div class="mb-4 flex items-center gap-x-2">
  <button
    on:click={() => {
      if (activeChips.length > 0) {
        activeChips = [...activeChips.slice(0, activeChips.length - 1)];
      }
    }}><Cancel /></button
  >
  <button
    on:click={() => {
      if (activeChips.length < chipValues.length) {
        activeChips = [
          ...activeChips,
          chipValues.find(
            (chip) => !activeChips.map((c) => c.name).includes(chip.name)
          ),
        ];
      }
    }}><Add /></button
  >
  <button
    on:click={() => {
      activeChips = [...chipValues];
      resetActives();
    }}>reset</button
  >
</div>

<ChipContainer>
  {#each activeChips as { name, typeLabel, tooltipContent } (name)}
    <div animate:flip={{ duration: 200 }}>
      <RemovableListChip
        {name}
        {typeLabel}
        selectedValues={actives[name]}
        on:remove={() => {
          activeChips = [...activeChips.filter((chip) => chip.name !== name)];
        }}
        on:select={(event) => {
          toggleActiveValue(name, event.detail);
        }}
      >
        <svelte:fragment slot="remove-tooltip-content">
          a custom tooltip for removing the {actives[name].length} value{#if actives[name].length !== 1}s{/if}.
          This is a custom tooltip that can be edited as a slot.
        </svelte:fragment>
        <svelte:fragment slot="body-tooltip-content"
          >click to change the values in this list</svelte:fragment
        >
      </RemovableListChip>
    </div>
  {/each}
</ChipContainer>
