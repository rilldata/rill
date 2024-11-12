<script lang="ts" context="module">
  const gradients = {
    hue: createGradientString(
      Array.from({ length: 360 }).map((_, i) => `hsl(${i}, 100%, 50%)`),
    ),
    saturation: createGradientString(
      Array.from({ length: 50 }).map(
        (_, i) => `hsl(var(--hue), ${i * 2}%, 50%)`,
      ),
    ),
    lightness: createGradientString(
      Array.from({ length: 50 }).map(
        (_, i) => `hsl(var(--hue), 100%, ${i * 2}%)`,
      ),
    ),
  };

  function createGradientString(array: string[]) {
    return `linear-gradient(to right, ${array.join(", ")})`;
  }
</script>

<script lang="ts">
  export let value: number;
  export let hue = 0;
  export let mode: "hue" | "saturation" | "lightness";
  export let onChange: () => void;
</script>

<div class="size-full flex flex-none gap-x-2 items-center">
  <input
    style:--hue={hue}
    style:background-image={gradients[mode]}
    type="range"
    min="0"
    max={mode === "hue" ? 360 : 100}
    bind:value
    on:change={onChange}
  />

  <input
    type="number"
    min="0"
    max={mode === "hue" ? 360 : 100}
    class="border rounded-sm pl-1 w-[50px]"
    bind:value
    on:change={onChange}
  />
</div>

<style lang="postcss">
  input:focus {
    @apply outline-none;
  }

  input[type="range"] {
    @apply rounded-full w-full;
    -webkit-appearance: none;
    box-shadow:
      rgba(255, 255, 255, 0.25) 0 1px 1px inset,
      rgba(0, 0, 0, 0.6) 0 0 0 1px;
    height: 18px;
    margin: 0;
  }

  input[type="range"]::-webkit-slider-thumb {
    @apply rounded-full;
    -webkit-appearance: none;
    box-shadow:
      rgba(0, 0, 0, 0.6) 0 0 0 1px inset,
      rgb(235, 235, 235) 0 0 0 2.5px,
      rgba(0, 0, 0, 0.6) 0 0 0 3.5px;
    height: 16px;
    width: 16px;
    background: transparent;
    cursor: pointer;
    margin-top: -0px;
  }

  input[type="range"]::-moz-range-thumb {
    @apply rounded-full;
    -webkit-appearance: none;
    box-shadow:
      rgba(0, 0, 0, 0.6) 0 0 0 1px inset,
      rgb(235, 235, 235) 0 0 0 2.5px,
      rgba(0, 0, 0, 0.6) 0 0 0 3.5px;
    height: 16px;
    width: 16px;
    background: transparent;
    cursor: pointer;
    margin-top: -0px;
  }

  input[type="range"]:hover::-webkit-slider-thumb {
    box-shadow:
      rgba(0, 0, 0, 0.6) 0 0 0 1px inset,
      rgba(255, 255, 255, 1) 0 0 0 2.5px,
      rgba(0, 0, 0, 0.6) 0 0 0 3.5px;
  }

  input[type="range"]:focus::-webkit-slider-thumb {
    box-shadow:
      rgba(0, 0, 0, 0.6) 0 0 0 1px inset,
      rgba(255, 255, 255, 1) 0 0 0 2.5px,
      hsl(var(--hue), 100%, 50%) 0 0 0 4.5px,
      rgba(0, 0, 0, 0.6) 0 0 0 5.5px;
  }

  input[type="range"]:hover::-moz-range-thumb {
    box-shadow:
      rgba(0, 0, 0, 0.6) 0 0 0 1px inset,
      rgba(255, 255, 255, 1) 0 0 0 2.5px,
      rgba(0, 0, 0, 0.6) 0 0 0 3.5px;
  }

  input[type="range"]:focus::-moz-range-thumb {
    box-shadow:
      rgba(0, 0, 0, 0.6) 0 0 0 1px inset,
      rgba(255, 255, 255, 1) 0 0 0 2.5px,
      hsl(var(--hue), 100%, 50%) 0 0 0 4.5px,
      rgba(0, 0, 0, 0.6) 0 0 0 5.5px;
  }
</style>
