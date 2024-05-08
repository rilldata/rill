<script>
  import Base from "./Base.svelte";
  import { isPercDiff } from "./type-utils";
  export let inTable = false;
  export let dark = false;
  export let customStyle = "";
  export let value;

  let isNull = false;
  $: if (isPercDiff(value)) {
    isNull = true;
  }
</script>

<Base
  {isNull}
  classes="ui-copy-number w-full font-normal {customStyle} {inTable &&
    'block text-right'}"
  {dark}
>
  {value}
</Base>
