<script lang="ts">
  import InputV2 from "@rilldata/web-common/components/forms/InputV2.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import {
    generateColorPaletteUsingDarken,
    generateColorPaletteUsingScale,
    generateColorPaletteUsingScaleDifferentStartAndEnd,
  } from "@rilldata/web-common/features/themes/color-palette";
  import { copySaturationAndLightness } from "@rilldata/web-common/features/themes/theme-actions";
  import chroma, { Color } from "chroma-js";

  let input = "darkred";

  let palettes: Array<{
    title: string;
    palette: Array<Color>;
  }>;
  $: if (input) {
    try {
      palettes = [
        {
          title: "Copy saturation and lightness from blue palette (default)",
          palette: copySaturationAndLightness(chroma(input)),
        },
        {
          title: "Lighten and darken with fixed position (mode=v1)",
          palette: generateColorPaletteUsingScale(chroma(input)),
        },
        {
          title: `Palette based on lightness scale from white -> ${input} -> black (mode=v2)`,
          palette: generateColorPaletteUsingScale(chroma(input)),
        },
        {
          title: `Palette based on lightness scale from non-white -> ${input} -> -non-black`,
          palette: generateColorPaletteUsingScaleDifferentStartAndEnd(
            chroma(input)
          ),
        },
        {
          title: `Palette using chroma.darken()`,
          palette: generateColorPaletteUsingDarken(chroma(input)),
        },
      ];
      console.log(palettes);
    } catch (err) {}
  }
</script>

<div>
  <InputV2 bind:value={input} error="" />
</div>
<Spacer />

{#if palettes}
  <table>
    <tbody>
      {#each palettes as entry}
        <tr>
          <td class="w-60 text-lg p-5">{entry.title}</td>
          <td class="flex flex-row gap-1">
            {#each entry.palette as color}
              <div
                class="w-24 h-40 text-center"
                style="background-color: {color.hex()}"
              >
                <span class="text-lg text-white">{color.hex()}</span>
                <span class="text-lg text-black">{color.hex()}</span>
              </div>
            {/each}
          </td>
        </tr>
        <Spacer />
      {/each}
    </tbody>
  </table>
{/if}
