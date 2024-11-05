<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { Trash2Icon } from "lucide-svelte";
  import { createEventDispatcher } from "svelte";

  export let index: number;
  export let variable: { key: string; value: string };
  export let isLast: boolean;

  $: console.log("variable: ", variable);

  const dispatch = createEventDispatcher();
</script>

<div class="flex flex-row items-center gap-2" id={`variable-${index}`}>
  <Input
    bind:value={variable.key}
    id={`key-${index}`}
    label=""
    placeholder="Key"
  />
  <Input
    bind:value={variable.value}
    id={`value-${index}`}
    label=""
    placeholder="Value"
  />
  <IconButton
    on:click={() => {
      if (!isLast) {
        dispatch("delete", index);
      }
    }}
  >
    <Trash2Icon size="16px" />
  </IconButton>
</div>
