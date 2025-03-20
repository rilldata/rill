<script lang="ts">
  export let show: boolean;
  export let vertical = false;
  export let horizontal = false;
</script>

<span
  class:bg-indigo-300={show}
  style:width={vertical ? "3px" : "100%"}
  style:height={horizontal ? "3px" : "100%"}
  class="rounded-full pointer-events-none group-hover:bg-indigo-300 opacity-80"
/>
