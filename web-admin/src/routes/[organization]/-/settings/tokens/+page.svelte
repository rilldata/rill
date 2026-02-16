<script lang="ts">
  import { page } from "$app/stores";
  import ServiceTokenList from "@rilldata/web-admin/lib/features/tokens/ServiceTokenList.svelte";
  import UserTokenList from "@rilldata/web-admin/lib/features/tokens/UserTokenList.svelte";
  import CreateServiceTokenDialog from "@rilldata/web-admin/lib/features/tokens/CreateServiceTokenDialog.svelte";
  import CreateUserTokenDialog from "@rilldata/web-admin/lib/features/tokens/CreateUserTokenDialog.svelte";
  import DeleteTokenDialog from "@rilldata/web-admin/lib/features/tokens/DeleteTokenDialog.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { Tab, TabList } from "@rilldata/web-common/components/tabs";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    deleteServiceTokenMutation,
    deleteUserTokenMutation,
  } from "@rilldata/web-admin/lib/features/tokens/token-queries";

  const organization = $page.params.organization;

  // Tab state
  type TabType = "service" | "user";
  let activeTab: TabType = "service";

  // Dialog state
  let showCreateServiceTokenDialog = false;
  let showCreateUserTokenDialog = false;
  let showDeleteDialog = false;
  let deleteTokenId = "";
  let deleteTokenName = "";
  let deleteTokenType: "service" | "user" = "service";

  // Mutations for delete
  const deleteServiceToken = deleteServiceTokenMutation(organization);
  const deleteUserToken = deleteUserTokenMutation();

  function handleCreateClick() {
    if (activeTab === "service") {
      showCreateServiceTokenDialog = true;
    } else {
      showCreateUserTokenDialog = true;
    }
  }

  function handleTabChange(tab: TabType) {
    activeTab = tab;
  }

  function handleDeleteRequest(
    event: CustomEvent<{
      tokenId: string;
      tokenName: string;
      tokenType: "service" | "user";
    }>,
  ) {
    deleteTokenId = event.detail.tokenId;
    deleteTokenName = event.detail.tokenName;
    deleteTokenType = event.detail.tokenType;
    showDeleteDialog = true;
  }

  async function handleDeleteConfirm(): Promise<void> {
    try {
      if (deleteTokenType === "service") {
        await $deleteServiceToken.mutateAsync({
          organization,
          tokenId: deleteTokenId,
        });
      } else {
        await $deleteUserToken.mutateAsync({
          tokenId: deleteTokenId,
        });
      }
      showDeleteDialog = false;
      eventBus.emit("notification", {
        message: `Token "${deleteTokenName}" has been revoked`,
        type: "success",
      });
    } catch (e) {
      // Error is handled within DeleteTokenDialog
      throw e;
    }
  }
</script>

<div class="flex flex-col gap-4 w-full">
  <!-- Page header -->
  <div class="flex items-center justify-between">
    <div>
      <h2 class="text-lg font-semibold text-gray-900">Tokens</h2>
      <p class="text-sm text-gray-500 mt-1">
        Manage API tokens for programmatic access to your organization.
      </p>
    </div>
    <Button on:click={handleCreateClick}>
      Create Token
    </Button>
  </div>

  <!-- Tabs -->
  <div class="border-b border-gray-200">
    <TabList>
      <Tab
        active={activeTab === "service"}
        on:click={() => handleTabChange("service")}
      >
        Service Tokens
      </Tab>
      <Tab
        active={activeTab === "user"}
        on:click={() => handleTabChange("user")}
      >
        User Tokens
      </Tab>
    </TabList>
  </div>

  <!-- Tab content -->
  {#if activeTab === "service"}
    <ServiceTokenList
      {organization}
      on:delete={handleDeleteRequest}
    />
  {:else}
    <UserTokenList
      on:delete={handleDeleteRequest}
    />
  {/if}
</div>

<!-- Create Service Token Dialog -->
<CreateServiceTokenDialog
  bind:open={showCreateServiceTokenDialog}
  {organization}
/>

<!-- Create User Token Dialog -->
<CreateUserTokenDialog
  bind:open={showCreateUserTokenDialog}
/>

<!-- Delete Token Dialog -->
<DeleteTokenDialog
  open={showDeleteDialog}
  tokenName={deleteTokenName}
  tokenType={deleteTokenType}
  onConfirm={handleDeleteConfirm}
  on:close={() => {
    showDeleteDialog = false;
  }}
/>