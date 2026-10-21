<script lang="ts">
  import type { Conversation } from "../../conversation";
  import FeedbackModal from "../FeedbackModal.svelte";

  let { conversation }: { conversation: Conversation } = $props();

  // Mirrors Messages.svelte: closing the modal clears the pending message id.
  let open = $state(true);
  let messageId = $state<string | null>("assistant-message-1");
</script>

<FeedbackModal
  {open}
  {messageId}
  {conversation}
  agent="analyst_agent"
  onClose={() => {
    open = false;
    messageId = null;
  }}
/>
