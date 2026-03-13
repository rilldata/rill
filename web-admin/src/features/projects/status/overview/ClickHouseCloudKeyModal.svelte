<script lang="ts">
  import {
    createAdminServiceUpdateProjectVariables,
    getAdminServiceGetProjectVariablesQueryKey,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import { AXIOS_INSTANCE } from "@rilldata/web-admin/client/http-client";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
  } from "@rilldata/web-common/components/dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";
  import type { AxiosError } from "axios";

  export let open = false;
  export let organization: string;
  export let project: string;
  export let connectorHost: string | undefined = undefined;
  export let onDismiss: () => void;
  export let onSave: ((memoryGb?: number) => void) | undefined = undefined;

  let keyId = "";
  let keySecret = "";
  let validating = false;

  const queryClient = useQueryClient();
  const updateVariables = createAdminServiceUpdateProjectVariables();

  $: canSubmit =
    keyId.trim() !== "" &&
    keySecret.trim() !== "" &&
    !$updateVariables.isPending &&
    !validating;

  async function lookupChcService(): Promise<number | undefined> {
    if (!connectorHost) {
      console.log("[CHC Lookup] No connectorHost, skipping lookup");
      return undefined;
    }
    try {
      console.log("[CHC Lookup] Calling /v1/clickhouse-cloud/lookup", {
        host: connectorHost,
        org: organization,
        project,
      });
      const resp = await AXIOS_INSTANCE.post("/v1/clickhouse-cloud/lookup", {
        key_id: keyId.trim(),
        key_secret: keySecret.trim(),
        host: connectorHost,
        org: organization,
        project,
      });
      console.log("[CHC Lookup] Response:", resp.data);
      return resp.data?.max_memory_gb;
    } catch (err) {
      console.error("[CHC Lookup] Failed:", err);
      return undefined;
    }
  }

  async function handleSubmit() {
    try {
      console.log(
        "[CHC Lookup] handleSubmit called, connectorHost:",
        connectorHost,
      );
      // First, validate the key by looking up the service
      validating = true;
      const memoryGb = await lookupChcService();
      validating = false;
      console.log("[CHC Lookup] memoryGb result:", memoryGb);

      // Save the key as project variables
      await $updateVariables.mutateAsync({
        org: organization,
        project,
        data: {
          environment: "prod",
          variables: {
            CLICKHOUSE_CLOUD_API_KEY_ID: keyId.trim(),
            CLICKHOUSE_CLOUD_API_KEY_SECRET: keySecret.trim(),
          },
        },
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceGetProjectVariablesQueryKey(
          organization,
          project,
        ),
      });

      eventBus.emit("notification", {
        message: "ClickHouse Cloud API key saved",
      });
      keyId = "";
      keySecret = "";
      open = false;
      onSave?.(memoryGb);
    } catch (err) {
      validating = false;
      const axiosError = err as AxiosError<RpcStatus>;
      eventBus.emit("notification", {
        message:
          axiosError.response?.data?.message ??
          "Failed to save ClickHouse Cloud API key",
        type: "error",
      });
    }
  }

  function handleRemindLater() {
    onDismiss();
    open = false;
  }
</script>

<Dialog
  bind:open
  onOpenChange={(isOpen) => {
    if (!isOpen) {
      keyId = "";
      keySecret = "";
    }
  }}
>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Connect to ClickHouse Cloud</DialogTitle>
    </DialogHeader>
    <DialogDescription>
      We detected your project is using ClickHouse Cloud. Enter your Admin API
      key to unlock cluster monitoring, auto-detect scaling settings, and view
      service status directly in Rill.
    </DialogDescription>

    <div class="flex flex-col gap-4 mt-2">
      <Input
        bind:value={keyId}
        id="chc-key-id"
        label="API Key ID"
        placeholder="Key ID"
        secret
      />
      <Input
        bind:value={keySecret}
        id="chc-key-secret"
        label="API Key Secret"
        placeholder="Key Secret"
        secret
      />
      <p class="text-xs text-fg-tertiary">
        You can create an API key in your <a
          href="https://clickhouse.cloud/settings/api-keys"
          target="_blank"
          rel="noopener noreferrer"
          class="text-primary-500 hover:underline">ClickHouse Cloud console</a
        >. The key needs read access to the Cloud API.
      </p>
    </div>

    <DialogFooter>
      <Button type="tertiary" onClick={handleRemindLater}>
        Remind me later
      </Button>
      <Button type="primary" disabled={!canSubmit} onClick={handleSubmit}>
        {#if validating}
          Validating...
        {:else if $updateVariables.isPending}
          Saving...
        {:else}
          Save API Key
        {/if}
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
