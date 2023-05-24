<script lang="ts">
  export let bgClasses = "";
</script>

<div
  class="w-[240px] h-[240px] rounded-md
  flex flex-col items-center justify-center gap-y-2
  relative {bgClasses} 
  transition duration-300 ease-out custom-shadow custom-shadow-hover cursor-pointer"
  on:click
>
  <slot />
</div>

<style>
  .custom-shadow {
    box-shadow: 0px 2px 3px rgba(15, 23, 42, 0.06),
      0px 1px 3px rgba(15, 23, 42, 0.08), 0px 0px 0px 1px rgba(15, 23, 42, 0.12);
  }

  .custom-shadow-hover:hover {
    box-shadow: 0px 2px 3px rgba(99, 102, 241, 0.2),
      0px 1px 3px rgba(15, 23, 42, 0.08), 0px 0px 0px 1px rgba(15, 23, 42, 0.12),
      0px 4px 6px rgba(15, 23, 42, 0.12);
  }
</style>
