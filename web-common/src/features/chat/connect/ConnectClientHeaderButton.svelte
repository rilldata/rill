<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import { Plug } from "lucide-svelte";
  import { getConnectClientContext } from "./connect-client-context";

  const connectClient = getConnectClientContext();
  const label = "Connect Claude, ChatGPT & more";
</script>

<IconButton ariaLabel={label} bgGray onclick={() => connectClient.open()}>
  <Plug size="16px" class="text-fg-muted" />
  <svelte:fragment slot="tooltip-content">{label}</svelte:fragment>
</IconButton>
