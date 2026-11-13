<!--
  Prompts the viewer to sign in when a query fails because they haven't authorized a per-user OAuth connector.
  Rendered once per project, so one prompt covers every dashboard surface.
-->
<script lang="ts">
  import { onMount } from "svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { openMcpConnectPopup } from "@rilldata/web-common/features/chat/connect/mcp-servers";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { connectorOAuthConnectUrl } from "./connect-url";
  import {
    connectorSignInRequired,
    watchConnectorSignInErrors,
  } from "./connector-oauth-error";

  let {
    organization,
    project,
    branch,
  }: {
    organization: string;
    project: string;
    branch?: string;
  } = $props();

  let signingIn = $state(false);
  let failure = $state<string | null>(null);

  onMount(() => watchConnectorSignInErrors());

  // Reset on project change, so the button never starts a flow for another project with the old connector.
  $effect(() => {
    void organization;
    void project;
    return dismiss;
  });

  function dismiss() {
    connectorSignInRequired.set(null);
    failure = null;
  }

  async function signIn() {
    const connector = $connectorSignInRequired;
    if (!connector) return;
    signingIn = true;
    failure = null;
    const result = await openMcpConnectPopup(
      connectorOAuthConnectUrl(organization, project, connector, branch),
      connector,
    );
    signingIn = false;
    if (result.status === "connected") {
      dismiss();
      void queryClient.refetchQueries({
        predicate: (q) => q.state.status === "error",
      });
    } else if (result.status === "error") {
      failure = m.connector_sign_in_failed({ error: result.error ?? "" });
    }
  }
</script>

<Dialog.Root
  bind:open={
    () => $connectorSignInRequired !== null,
    (open) => {
      if (!open) dismiss();
    }
  }
>
  <Dialog.Content class="max-w-md">
    <Dialog.Header>
      <Dialog.Title>{m.connector_sign_in_title()}</Dialog.Title>
      <Dialog.Description
        >{m.connector_sign_in_description()}</Dialog.Description
      >
    </Dialog.Header>
    {#if failure}
      <p class="text-sm text-destructive">{failure}</p>
    {/if}
    <Dialog.Footer>
      <Button
        type="primary"
        onClick={signIn}
        loading={signingIn}
        disabled={signingIn}
      >
        {m.connector_sign_in_button()}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
