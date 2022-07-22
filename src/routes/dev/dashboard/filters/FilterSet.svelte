<script lang="ts">
  /** for tailwind css: flex-row and flex-col */
  export let direction: "row" | "col" = "row";
  export let height: string = undefined;
  export let width: string = undefined;
  /** tailwind: flex inline-flex */
  export let style: "flex" | "inline-flex" = "flex";
</script>

<div
  style:height
  style:width
  style:outline="1px solid black"
  class="filter-set {style} flex-{direction} flex-wrap justify-items-start items-start h-full"
>
  <slot />
</div>
