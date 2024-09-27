<script lang="ts">
  export let checked: boolean | number = false;
  export let disabled = false;
  export let onChange: (checked: boolean) => void;
  export let size: number = 16;
  export let className: string = "";

  $: fullyChecked = checked === 1 || checked === true;
</script>

<button
  on:click={() => onChange(!fullyChecked)}
  {disabled}
  class={className}
  class:checked={fullyChecked}
  style:width="{size}px"
  style:height="{size}px"
>
  {#if fullyChecked}
    <svg
      class="size-full"
      viewBox="-0.4 0 11 10"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        fill-rule="evenodd"
        clip-rule="evenodd"
        d="M10.4604 1.35316C10.6622 1.52943 10.6892 1.83341 10.5216 2.04246L4.29328 9.81186L0.355406 5.86535C0.15948 5.669 0.16084 5.35068 0.358435 5.15601L1.25035 4.27729C1.44559 4.08493 1.7593 4.08561 1.9537 4.27881L4.19416 6.50532L8.80174 0.604638C8.97835 0.378467 9.30866 0.347021 9.52477 0.535807L10.4604 1.35316Z"
        fill="white"
      />
    </svg>
  {:else if typeof checked === "number" && checked > 0}
    <div class="bg-primary-400 size-2" />
  {/if}
</button>

<style lang="postcss">
  button {
    @apply border rounded-[2px] bg-gray-50 grid place-content-center p-[1.5px];
  }

  button:hover {
    @apply border-primary-400;
  }
  .checked {
    @apply bg-primary-400 border-primary-400;
  }

  .checked:hover {
    @apply bg-primary-300 border-primary-300;
  }
</style>
