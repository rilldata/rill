<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { Trash2Icon } from "lucide-svelte";
  import { createEventDispatcher } from "svelte";

  export let index: number;
  export let variable: { key: string; value: string };

  const dispatch = createEventDispatcher();
</script>

<div class="flex flex-row items-center gap-2" id={`variable-${index}`}>
  <Input value={variable.key} id="key" label="" placeholder="Key" />
  <Input value={variable.value} id="value" label="" placeholder="Value" />
  <button on:click={() => dispatch("delete", index)}>
    <Trash2Icon size="16px" />
  </button>
</div>
