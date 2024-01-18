<script lang="ts">
  import {
    applyShiftToPalette,
    getPaletteAndShift,
  } from "@rilldata/web-common/features/themes/palette-generator";
  import chroma from "chroma-js";
  import type { Color } from "chroma-js";
  import InputV2 from "@rilldata/web-common/components/forms/InputV2.svelte";

  let input: string;

  let inputPalette: Color[];
  let matchedPalette: Color[];
  $: if (input) {
    try {
      const inputColor = chroma(input);
      const [palette, shift] = getPaletteAndShift(inputColor);
      matchedPalette = palette;
      inputPalette = applyShiftToPalette(palette, shift);
    } catch (err) {
      // no-op
    }
  }
</script>

<div class="p-5">
  <InputV2 bind:value={input} error="" placeholder="Input color" />
  <br />

  {#if inputPalette}
    <div>Input Palette</div>
    <div class="flex flex-row gap-1">
      {#each inputPalette as color}
        <div style="background-color:{color.name()}" class="h-20 w-10"></div>
      {/each}
    </div>
  {/if}

  {#if matchedPalette}
    <div>Matched Palette</div>
    <div class="flex flex-row gap-1">
      {#each matchedPalette as color}
        <div style="background-color:{color.name()}" class="h-20 w-10"></div>
      {/each}
    </div>
  {/if}
</div>
