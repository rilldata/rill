<script lang="ts">
  import UpArrow from "./UpArrow.svelte";

  export let size: string | number = "1em";
  export let disabled = false;

  $: backgroundClass = disabled ? "fill-gray-300" : "fill-primary-400";
  $: arrowClass = "text-gray-100";
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  xmlns="http://www.w3.org/2000/svg"
>
  <circle cx="12" cy="12" r="10" class={backgroundClass} />
  <UpArrow size={24} className={arrowClass} />
</svg>
