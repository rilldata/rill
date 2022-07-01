<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import Trash from "$lib/components/icons/Trash.svelte";
  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";

  const dispatch = createEventDispatcher();

  export let value;
</script>

<ContextButton
  id={value}
  tooltipText="delete"
  location="left"
  on:click={() => dispatch("delete", value)}
  ><Trash size="1.2em" /></ContextButton
>
