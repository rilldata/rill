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
  export let color: string;
  export let mode: "hue" | "saturation" | "lightness";
  export let onChange: () => void;
</script>

<div class="size-full flex flex-none gap-x-2 items-center">
  <input
    style:--hue={hue}
    style:--color={color}
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
  * {
    --focus: rgba(255, 255, 255, 1) 0 0 0 2.5px, var(--color) 0 0 0 4.5px;
    --hover: rgba(255, 255, 255, 1) 0 0 0 2.5px,
      rgba(0, 0, 0, 0.2) 0 0 3px 3.5px;
  }

  input:focus {
    @apply outline-none;
  }

  input[type="range"] {
    @apply rounded-full w-full;
    -webkit-appearance: none;

    height: 13px;
    margin: 0;
  }

  input[type="range"]::-webkit-slider-thumb {
    @apply rounded-full;
    -webkit-appearance: none;
    box-shadow:
      rgba(255, 255, 255, 1) 0 0 0 2.5px,
      rgba(0, 0, 0, 0.2) 0 0 3px 3.5px;
    height: 14px;
    width: 14px;
    background: var(--color);
    cursor: pointer;
    margin-top: -0px;
  }

  input[type="range"]::-moz-range-thumb {
    @apply rounded-full;
    -webkit-appearance: none;
    box-shadow:
      rgba(255, 255, 255, 1) 0 0 0 2.5px,
      rgba(0, 0, 0, 0.2) 0 0 3px 3.5px;
    height: 16px;
    width: 16px;
    background: transparent;
    cursor: pointer;
    margin-top: -0px;
  }

  input[type="range"]:hover::-webkit-slider-thumb {
    box-shadow: var(--hover);
  }

  input[type="range"]:focus::-webkit-slider-thumb {
    box-shadow: var(--focus);
  }

  input[type="range"]:hover::-moz-range-thumb {
    box-shadow: var(--hover);
  }

  input[type="range"]:focus::-moz-range-thumb {
    box-shadow: var(--focus);
  }
</style>
