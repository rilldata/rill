<script lang="ts">
  export let side: "left" | "right";
</script>

<div
  class="sticky top-0 bg-white z-10 px-4 flex flex-row items-center gap-x-2 justify-{side ===
  'left'
    ? 'start'
    : 'end'}"
  style:height="var(--header-height)"
  style="padding-{side === 'left' ? 'right' : 'left'}: 40px;"
>
  <slot />
</div>
