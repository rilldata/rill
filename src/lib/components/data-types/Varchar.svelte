<script>
  import Base from "./Base.svelte";
  export let isNull = false;
  export let inTable = false;
  export let dark = false;
</script>

<Base {isNull} classes={inTable && "text-left"} {dark}>
  <slot />
</Base>
