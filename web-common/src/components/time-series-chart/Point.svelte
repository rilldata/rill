<script lang="ts">
  export let y: number;
  export let x: number;
  export let color: string;
</script>

<path
  d="M {x} {y} l 0.0001 0"
  stroke-linecap="round"
  stroke-width="9"
  stroke="white"
  vector-effect="non-scaling-stroke"
/>

<path
  d="M {x} {y} l 0.0001 0"
  stroke-linecap="round"
  stroke-width="6.5"
  stroke={color}
  vector-effect="non-scaling-stroke"
/>

<style>
  * {
    pointer-events: none;
  }
</style>
