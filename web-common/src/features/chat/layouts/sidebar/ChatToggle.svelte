<script lang="ts">
  import Button from "../../../../components/button/Button.svelte";
  import { chatOpen, sidebarActions } from "./sidebar-store";
</script>

<Button
  compact
  type="secondary"
  onClick={sidebarActions.toggleChat}
  active={$chatOpen}
>
  AI
</Button>
