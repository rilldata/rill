<!-- a simple two-button CTA -->
<script lang="ts">
  import { Button } from "$lib/components/button";

  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();

  export let disabled = false;
  export let compact = false;
  export let destructiveAction = false;
  export let showSecondary = false;
</script>

<Button type="text" {compact} on:click={() => dispatch("cancel")}
  ><slot name="cancel-body">Cancel</slot></Button
>
<div class={!showSecondary ? "hidden" : ""}>
  <Button
    type="text"
    on:click={() => {
      dispatch("secondary-action");
    }}><slot name="secondary-action-body" /></Button
  >
</div>
<Button
  type="primary"
  status={destructiveAction ? "error" : "info"}
  {compact}
  {disabled}
  on:click={() => {
    dispatch("primary-action");
  }}><slot name="primary-action-body">Submit</slot></Button
>
