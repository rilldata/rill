<script lang="ts">
  import * as Popover from "@rilldata/web-common/components/popover";
  import InputLabel from "../forms/InputLabel.svelte";
  import ColorSlider from "./ColorSlider.svelte";

  export let stringColor: string | undefined;
  export let label: string;
  export let disabled = false;
  export let onChange: (color: string) => void;
  export let small = false;

  let open = false;

  $: ({ h: hue, s: saturation, l: lightness } = stringColorToHsl(stringColor));

  $: hsl = `hsl(${hue}, ${saturation}%, ${lightness}%)`;

  function extractHSL(color: string) {
    const [, hue, saturation, lightness] = color.match(
      /hsl\((\d+),\s*(\d+)%,\s*(\d+)%\)/,
    ) || [null, 0, 100, 50];

    return {
      h: +hue,
      s: +saturation,
      l: +lightness,
    };
  }

  function stringColorToHsl(string: string | undefined) {
    if (!string) {
      return {
        h: 0,
        s: 100,
        l: 50,
      };
    }
    if (string.startsWith("hsl")) {
      return extractHSL(string);
    }

    const color = getComputedColor(
      hexColorWithoutPound(string) ? `#${string}` : string,
    );

    if (!color)
      return {
        h: 0,
        s: 100,
        l: 50,
      };

    if (color.startsWith("rgba")) {
      return rgbaToHSL(color);
    }

    const hsl = hexToHSL(color);

    return hsl;
  }

  function rgbaToHSL(string: string) {
    const matches = string.match(
      /rgba\((\d+),\s*(\d+),\s*(\d+),\s*(\d+(\.\d+)?)\)/,
    );

    if (!matches) {
      return {
        h: 0,
        s: 0,
        l: 0,
      };
    }

    const [, red, green, blue] = matches;

    const r = +red / 255;
    const g = +green / 255;
    const b = +blue / 255;

    let max = Math.max(r, g, b);
    let min = Math.min(r, g, b);

    let h = (max + min) / 2;
    let s = h;
    let l = h;

    if (max === min) {
      return { h: 0, s: 0, l };
    }

    const d = max - min;
    s = l >= 0.5 ? d / (2 - (max + min)) : d / (max + min);
    switch (max) {
      case r:
        h = ((g - b) / d + 0) * 60;
        break;
      case g:
        h = ((b - r) / d + 2) * 60;
        break;
      case b:
        h = ((r - g) / d + 4) * 60;
        break;
    }

    return {
      h: Math.round(h),
      s: Math.round(s * 100),
      l: Math.round(l * 100),
    };
  }

  function hexColorWithoutPound(color: string) {
    return /^[0-9A-F]{6}$/i.test(color);
  }

  function getComputedColor(color: string) {
    const canvas = document.createElement("canvas").getContext("2d");
    if (!canvas) return;
    canvas.fillStyle = color;
    return canvas.fillStyle;
  }

  function hexToHSL(hex: string) {
    const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);

    let h = 0;
    let s = 100;
    let l = 50;

    if (!result)
      return {
        h,
        s,
        l,
      };

    let r = parseInt(result[1], 16);
    let g = parseInt(result[2], 16);
    let b = parseInt(result[3], 16);

    r /= 255;
    g /= 255;
    b /= 255;

    let max = Math.max(r, g, b),
      min = Math.min(r, g, b);
    h = s = l = (max + min) / 2;

    if (max == min) {
      h = s = 0;
    } else {
      var d = max - min;
      s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
      switch (max) {
        case r:
          h = (g - b) / d + (g < b ? 6 : 0);
          break;
        case g:
          h = (b - r) / d + 2;
          break;
        case b:
          h = (r - g) / d + 4;
          break;
      }

      h /= 6;
    }

    h = Math.round(h * 360);
    s = Math.round(s * 100);
    l = Math.round(l * 100);

    return { h, s, l };
  }
</script>

<svelte:window
  on:keydown={(e) => {
    if (e.key === "Escape" || e.key === "Enter") {
      open = false;
    }
  }}
/>

<div
  class="color-wrapper"
  class:small
  class:pointer-events-none={disabled}
  class:bg-gray-50={disabled}
  class:text-gray-400={disabled}
>
  <Popover.Root bind:open>
    <Popover.Trigger asChild let:builder>
      <button
        class="trigger"
        use:builder.action
        {...builder}
        style:background-color={hsl}
      />
    </Popover.Trigger>

    <Popover.Content class="w-[270px] space-y-2" align="start" sideOffset={10}>
      <div class="space-y-1">
        <InputLabel label="Hue" id="hue" />
        <ColorSlider
          mode="hue"
          bind:value={hue}
          {hue}
          color={hsl}
          onChange={() => {
            stringColor = `hsl(${hue}, ${saturation}%, 50%)`;
            onChange(stringColor);
          }}
        />
      </div>
      <div class="space-y-1">
        <InputLabel label="Saturation" id="saturation" />
        <ColorSlider
          mode="saturation"
          bind:value={saturation}
          {hue}
          color={hsl}
          onChange={() => {
            stringColor = `hsl(${hue}, ${saturation}%, 50%)`;
            onChange(stringColor);
          }}
        />
      </div>
    </Popover.Content>
  </Popover.Root>

  <input
    class:small
    bind:value={stringColor}
    {disabled}
    on:keydown={(e) => {
      if (e.key === "Enter") {
        e.currentTarget.blur();
      }
    }}
    on:blur={() => {
      if (stringColor) {
        onChange(stringColor);
      }
    }}
  />

  <p class:small class:text-gray-500={!disabled}>{label}</p>
</div>

<style lang="postcss">
  .trigger {
    @apply h-full aspect-square;
    @apply rounded-[2px];
  }

  .color-wrapper {
    @apply py-[5px] px-[5px] pr-3;
    @apply h-8 w-full border border-gray-300 rounded-[2px];
    @apply flex gap-x-3;
  }

  .color-wrapper.small {
    @apply h-6 text-xs;
  }

  .color-wrapper:focus-within {
    @apply border-primary-500 ring-2 ring-primary-100;
  }

  p {
    @apply text-sm;
  }

  p.small {
    @apply text-xs;
  }

  input {
    @apply w-full text-sm;
    @apply outline-none border-0 bg-transparent;
  }

  input.small {
    @apply text-xs;
  }
</style>
