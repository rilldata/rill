<script lang="ts">
  import ConnectClientButtons from "../ConnectClientButtons.svelte";
  import {
    setConnectClientContext,
    type ConnectClientProvider,
  } from "../connect-client-context";

  export let open: (provider?: ConnectClientProvider) => void;

  setConnectClientContext({ enabled: true, open });
</script>

<ConnectClientButtons />
