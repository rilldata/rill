<script lang="ts">
  export let borderBottom: boolean = false;
  export let tinted: boolean = false;

  // Editing surface uses a lavender wash in light mode and a deep
  // purple in dark mode (theme-primary/100 → theme-primary/900 in
  // Figma).
  $: bgClass = tinted ? "bg-primary-50 dark:bg-primary-900" : "bg-surface-base";
</script>

<header
  class="flex items-center w-full pr-4 pl-2 py-1 border-b {bgClass}"
  class:border-transparent={!borderBottom}
>
  <slot />
</header>
