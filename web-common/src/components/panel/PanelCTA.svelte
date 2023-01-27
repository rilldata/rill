<script lang="ts">
  export let side: "left" | "right";
</script>

<div
  class="flex flex-row items-center gap-x-2 justify-{side === 'left'
    ? 'start'
    : 'end'}"
  style:height="var(--header-height)"
>
  <slot />
</div>
