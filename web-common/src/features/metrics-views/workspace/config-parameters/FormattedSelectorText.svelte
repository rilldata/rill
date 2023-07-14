<script lang="ts">
  import { SELECTOR_BUTTON_TEXT_CLASSES } from "../styles";

  export let value: string;
  export let selected = true;
</script>

<span
  style:max-width="14em"
  class={selected
    ? SELECTOR_BUTTON_TEXT_CLASSES.selected
    : SELECTOR_BUTTON_TEXT_CLASSES.unselected}
>
  {value}
</span>
