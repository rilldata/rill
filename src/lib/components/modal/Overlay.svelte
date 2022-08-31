<script lang="ts">
  import { fade } from "svelte/transition";
  export let bgColor = "bg-gray-900/30";
</script>

<div
  transition:fade={{ duration: 100 }}
  part="overlay"
  class="{bgColor} fixed top-0 right-0 left-0 bottom-0"
  on:click
  tabindex="-1"
/>
