<script lang="ts">
  import Button from "../../../../components/button/Button.svelte";
  import { chatOpen, sidebarActions } from "./sidebar-store";

  export let beta = false;
</script>

<Button
  compact
  type="secondary"
  onClick={sidebarActions.toggleChat}
  active={$chatOpen}
>
  AI {#if beta}(Beta){/if}
</Button>
