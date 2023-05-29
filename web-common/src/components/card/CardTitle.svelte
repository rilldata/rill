<script lang="ts">
  export let position: "middle" | "bottom" = "bottom";
</script>

<h3
  class="text-slate-600 font-medium text-sm {position === 'bottom'
    ? 'absolute top-[180px]'
    : ''}"
>
  <slot />
</h3>
