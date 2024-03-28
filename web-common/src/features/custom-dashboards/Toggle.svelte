<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";

  export let bool: boolean;
  export let text: [string, string];

  function toggleBoolean() {
    bool = !bool;
  }
</script>

<Button on:click={toggleBoolean}>
  {text[Number(bool)]}
</Button>
