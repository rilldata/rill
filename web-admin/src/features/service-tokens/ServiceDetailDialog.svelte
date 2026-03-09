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
    } catch (error) {
      console.error("Error issuing token", error);
      eventBus.emit("notification", {
        message: "Error issuing token",
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
    } catch (error) {
      console.error("Error revoking token", error);
      eventBus.emit("notification", {
        message: "Error revoking token",
        type: "error",
      });
    }
  }

  function formatDate(value: string | undefined) {
    if (!value) return "-";
    return new Date(value).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "numeric",
      minute: "numeric",
    });
  }
</script>

<Dialog
  bind:open
  onOpenChange={(isOpen) => {
    if (!isOpen) handleClose();
  }}
>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent class="max-w-lg">
    <DialogHeader>
      <DialogTitle>{serviceName}</DialogTitle>
    </DialogHeader>

    <div class="flex flex-col gap-y-5">
      <!-- Service info -->
      <div class="grid grid-cols-2 gap-y-2 text-sm">
        <span class="text-fg-tertiary">Org role</span>
        <span class="text-fg-primary">{service?.roleName ?? "-"}</span>
        <span class="text-fg-tertiary">Created</span>
        <span class="text-fg-primary">{formatDate(service?.createdOn)}</span>
        <span class="text-fg-tertiary">Updated</span>
        <span class="text-fg-primary">{formatDate(service?.updatedOn)}</span>
      </div>

      <!-- Project memberships -->
      {#if projectMemberships.length > 0}
        <div class="flex flex-col gap-y-2">
          <span class="text-sm font-medium text-fg-primary"
            >Project roles</span
          >
          <div class="flex flex-col gap-y-1 text-sm">
            {#each projectMemberships as pm}
              <div class="flex justify-between">
                <span class="text-fg-primary">{pm.projectName}</span>
                <span class="text-fg-tertiary">{pm.projectRoleName}</span>
              </div>
            {/each}
          </div>
        </div>
      {/if}

      <!-- Newly issued token -->
      {#if newlyIssuedToken}
        <div
          class="flex flex-col gap-y-2 p-3 rounded border border-yellow-300 bg-yellow-50"
        >
          <span class="text-sm font-medium text-fg-primary"
            >New token issued</span
          >
          <div class="flex items-center gap-x-2">
            <code
              class="text-xs bg-white border rounded px-2 py-1 flex-1 break-all"
            >
              {newlyIssuedToken}
            </code>
            <IconButton
              on:click={() => copyToClipboard(newlyIssuedToken)}
            >
              <CopyIcon size="14px" />
            </IconButton>
          </div>
          <span class="text-xs text-yellow-700">
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
                    Created {formatDate(token.createdOn)}
                    {#if token.expiresOn}
                      · Expires {formatDate(token.expiresOn)}
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
