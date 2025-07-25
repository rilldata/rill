<script lang="ts">
  import { Button } from "../../components/button";
  import { chatActions, chatOpen } from "./chat-store";
</script>

<Button
  compact
  type="secondary"
  onClick={chatActions.toggleChat}
  active={$chatOpen}>AI</Button
>
