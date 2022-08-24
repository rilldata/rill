<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { buttonClasses } from "./classes";

  export let type: "primary" | "secondary" | "text";
  export let disabled = false;
  export let compact = false;

  const dispatch = createEventDispatcher();

  const handleClick = (event: MouseEvent) => {
    if (!disabled) {
      dispatch("click", event);
    }
  };
</script>

<button
  {disabled}
  class={buttonClasses({ type, compact })}
  on:click={handleClick}
>
  <slot />
</button>
