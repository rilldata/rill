<script lang="ts">
  export let borderBottom: boolean = false;
</script>

<header
  class="flex items-center w-full pr-4 pl-2 py-1 bg-surface-base border-b"
  class:border-transparent={!borderBottom}
>
  <slot />
</header>
