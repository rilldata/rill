<script>
  import Base from "./Base.svelte";
  export let isNull = false;
  export let inTable = false;
  export const textAlign = "text-right";
  export let customStyle = "";
  export let dark;
</script>

<Base
  {isNull}
  classes="ui-copy-number font-semibold {customStyle} {dark
    ? 'text-slate-200'
    : 'text-slate-500'} {inTable && 'text-left'}"
  {dark}
>
  <slot />
</Base>
