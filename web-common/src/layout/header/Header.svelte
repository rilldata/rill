<script lang="ts">
  export let borderBottom: boolean = false;
  export let tinted: boolean = false;
</script>

<header
  class="flex items-center w-full pr-4 pl-2 py-1 border-b"
  class:bg-surface-base={!tinted}
  class:bg-primary-50={tinted}
  class:border-transparent={!borderBottom}
>
  <slot />
</header>
