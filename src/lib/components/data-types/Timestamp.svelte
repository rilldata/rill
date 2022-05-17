<script>
  import Base from "./Base.svelte";
  export let isNull = false;
  export let inTable = false;
  export const textAlign = "text-right";
  export let dark;
</script>

<Base
  {isNull}
  classes="italic font-semibold {dark
    ? 'text-slate-200'
    : 'text-slate-500'} {inTable && 'text-left'}"
  {dark}
>
  <slot />
</Base>
