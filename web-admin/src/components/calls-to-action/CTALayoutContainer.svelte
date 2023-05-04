<script lang="ts">
  // Components with tall content might want h-4/5
  export let height = "h-3/5";
</script>

<div class="flex flex-col justify-center items-center {height} m-auto">
  <slot />
</div>
