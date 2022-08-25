<script lang="ts">
  export let side: "left" | "right";
</script>

<div
  class="px-4 flex flex-row items-center gap-x-2 justify-{side === 'left'
    ? 'start'
    : 'end'}"
  style:height="var(--header-height)"
  style="margin-{side === 'left' ? 'right' : 'left'}: 28px;"
>
  <slot />
</div>
