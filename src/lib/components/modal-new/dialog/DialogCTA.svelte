<!-- a simple two-button CTA -->
<script lang="ts">
  import { Button } from "$lib/components/button";

  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();

  export let disabled = false;
  export let compact = false;
  export let destructiveAction = false;
</script>

<Button type="text" {compact} on:click={() => dispatch("cancel")}
  ><slot name="cancel-body">Cancel</slot></Button
>
<Button
  type="primary"
  status={destructiveAction ? "error" : "info"}
  {compact}
  {disabled}
  on:click={() => dispatch("submit")}
  ><slot name="submit-body">Submit</slot></Button
>
