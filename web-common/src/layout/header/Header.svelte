<script lang="ts">
  export let borderBottom: boolean = false;
</script>

<header
  class="flex items-center w-full h-11 pr-4 pl-2 py-1 bg-surface-base"
  class:border-b={borderBottom}
>
  <slot />
</header>
