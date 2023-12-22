<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { buttonClasses } from "./classes";

  export let type: "primary" | "secondary" | "highlighted" | "text" | "dashed" =
    "primary";
  export let status: "info" | "error" = "info";
  export let disabled = false;
  export let compact = false;
  export let submitForm = false;
  export let form = "";
  export let label: string | undefined = undefined;

  const dispatch = createEventDispatcher();

  const handleClick = (event: MouseEvent) => {
    if (!disabled) {
      dispatch("click", event);
    }
  };
</script>

<button
  {disabled}
  class={buttonClasses({ type, compact, status })}
  on:click={handleClick}
  type={submitForm ? "submit" : "button"}
  form={submitForm ? form : undefined}
  aria-label={label}
>
  <slot />
</button>
