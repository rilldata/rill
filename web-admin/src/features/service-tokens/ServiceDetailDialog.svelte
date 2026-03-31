<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetService,
    createAdminServiceListServiceAuthTokens,
    createAdminServiceIssueServiceAuthToken,
    createAdminServiceRevokeServiceAuthToken,
    getAdminServiceListServiceAuthTokensQueryKey,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { CopyIcon, Trash2Icon } from "lucide-svelte";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import { capitalize, formatOrgRole, formatServiceDateTime } from "./utils";

  export let open = false;
  export let serviceName: string;

  let newlyIssuedToken = "";
  let confirmRevokeId = "";

  $: organization = $page.params.organization;
  $: serviceQuery = createAdminServiceGetService(organization, serviceName, {
    query: { enabled: open },
  });
  $: service = $serviceQuery.data?.service;
  $: projectMemberships = $serviceQuery.data?.projectMemberships ?? [];
  $: attributes = Object.entries(
    (service?.attributes as Record<string, unknown>) ?? {},
  );
  $: tokensQuery = createAdminServiceListServiceAuthTokens(
    organization,
    serviceName,
    { query: { enabled: open } },
  );
  $: tokens = $tokensQuery.data?.tokens ?? [];

  const queryClient = useQueryClient();
  const issueToken = createAdminServiceIssueServiceAuthToken();
  const revokeToken = createAdminServiceRevokeServiceAuthToken();

  function handleClose() {
    newlyIssuedToken = "";
    confirmRevokeId = "";
  }

  async function handleIssueToken() {
    try {
      const result = await $issueToken.mutateAsync({
        org: organization,
        serviceName,
        data: {},
      });

      newlyIssuedToken = result.token ?? "";

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListServiceAuthTokensQueryKey(
          organization,
          serviceName,
        ),
      });

      eventBus.emit("notification", {
        message: "Token issued",
      });
    } catch (e: any) {
      console.error("Error issuing token", e);
      eventBus.emit("notification", {
        message: e?.response?.data?.message ?? "Error issuing token",
        type: "error",
      });
    }
  }

  async function handleRevokeToken(tokenId: string) {
    try {
      await $revokeToken.mutateAsync({ tokenId });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListServiceAuthTokensQueryKey(
          organization,
          serviceName,
        ),
      });

      confirmRevokeId = "";
      eventBus.emit("notification", {
        message: "Token revoked",
      });
    } catch (e: any) {
      console.error("Error revoking token", e);
      eventBus.emit("notification", {
        message: e?.response?.data?.message ?? "Error revoking token",
        type: "error",
      });
    }
  }
</script>

<Dialog
  bind:open
  onOpenChange={(isOpen) => {
    if (!isOpen) handleClose();
  }}
>
  <DialogTrigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </DialogTrigger>
  <DialogContent class="max-w-2xl">
    <DialogHeader>
      <DialogTitle>{serviceName}</DialogTitle>
    </DialogHeader>

    <div class="flex flex-col gap-y-5 max-h-[45vh] overflow-y-auto">
      <!-- Service info -->
      <div class="grid grid-cols-2 gap-y-2 text-sm">
        <span class="text-fg-tertiary">Organization access</span>
        <span class="text-fg-primary">{formatOrgRole(service?.roleName)}</span>
        <span class="text-fg-tertiary">Created</span>
        <span class="text-fg-primary"
          >{formatServiceDateTime(service?.createdOn)}</span
        >
      </div>

      <!-- Project memberships -->
      {#if projectMemberships.length > 0}
        <div class="flex flex-col gap-y-2">
          <span class="text-sm font-medium text-fg-primary">Project access</span
          >
          <div class="flex flex-col gap-y-1 text-sm">
            {#each projectMemberships as pm}
              <div class="flex justify-between">
                <span class="text-fg-primary">{pm.projectName}</span>
                <span class="text-fg-tertiary"
                  >{capitalize(pm.projectRoleName ?? "")}</span
                >
              </div>
            {/each}
          </div>
        </div>
      {/if}

      <!-- Custom attributes -->
      {#if attributes.length > 0}
        <div class="flex flex-col gap-y-2">
          <span class="text-sm font-medium text-fg-primary"
            >Custom attributes</span
          >
          <div class="flex flex-col border rounded divide-y">
            {#each attributes as [key, value]}
              <div class="flex items-center justify-between px-3 py-2 text-sm">
                <span class="text-fg-secondary">{key}</span>
                <code
                  class="text-xs text-fg-primary bg-surface-subtle rounded px-1.5 py-0.5"
                >
                  {String(value ?? "")}
                </code>
              </div>
            {/each}
          </div>
        </div>
      {/if}

      <!-- Newly issued token -->
      {#if newlyIssuedToken}
        <div class="flex flex-col gap-y-2 p-3 rounded border bg-surface-subtle">
          <span class="text-sm font-medium text-fg-primary"
            >New token issued</span
          >
          <div class="flex items-center gap-x-2">
            <code
              class="text-xs bg-surface-base border rounded px-2 py-1 flex-1 break-all"
            >
              {newlyIssuedToken}
            </code>
            <IconButton on:click={() => copyToClipboard(newlyIssuedToken)}>
              <CopyIcon size="14px" />
            </IconButton>
          </div>
          <span class="text-xs text-fg-secondary">
            Copy this token now. It will not be shown again.
          </span>
        </div>
      {/if}

      <!-- Tokens section -->
      <div class="flex flex-col gap-y-2">
        <div class="flex items-center justify-between">
          <span class="text-sm font-medium text-fg-primary">Tokens</span>
          <Button
            type="secondary"
            small
            onClick={handleIssueToken}
            disabled={$issueToken.isPending}
          >
            Issue token
          </Button>
        </div>

        {#if tokens.length === 0}
          <span class="text-sm text-fg-tertiary">No tokens issued</span>
        {:else}
          <div class="flex flex-col border rounded divide-y">
            {#each tokens as token}
              <div class="flex items-center justify-between px-3 py-2 text-sm">
                <div class="flex flex-col gap-y-0.5">
                  <span class="font-mono text-xs text-fg-primary"
                    >{token.prefix}...</span
                  >
                  <span class="text-xs text-fg-tertiary">
                    Created {formatServiceDateTime(token.createdOn)}
                    {#if token.expiresOn}
                      · Expires {formatServiceDateTime(token.expiresOn)}
                    {/if}
                  </span>
                </div>
                {#if confirmRevokeId === token.id}
                  <div class="flex items-center gap-x-1">
                    <Button
                      type="tertiary"
                      small
                      onClick={() => {
                        confirmRevokeId = "";
                      }}
                    >
                      Cancel
                    </Button>
                    <Button
                      type="destructive"
                      small
                      onClick={() => handleRevokeToken(token.id ?? "")}
                    >
                      Confirm
                    </Button>
                  </div>
                {:else}
                  <IconButton
                    on:click={() => {
                      confirmRevokeId = token.id ?? "";
                    }}
                  >
                    <Trash2Icon size="14px" class="text-fg-secondary" />
                  </IconButton>
                {/if}
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  </DialogContent>
</Dialog>
