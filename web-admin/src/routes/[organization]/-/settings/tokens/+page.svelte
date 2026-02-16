<script lang="ts">
  import { page } from "$app/stores";
  import { Button } from "@rilldata/web-common/components/button";
  import ServiceTokenList from "@rilldata/web-admin/lib/features/tokens/ServiceTokenList.svelte";
  // import UserTokenList from "@rilldata/web-admin/lib/features/tokens/UserTokenList.svelte";
  // import CreateServiceTokenDialog from "@rilldata/web-admin/lib/features/tokens/CreateServiceTokenDialog.svelte";
  // import DeleteTokenDialog from "@rilldata/web-admin/lib/features/tokens/DeleteTokenDialog.svelte";

  type TabType = "service" | "user";

  let activeTab: TabType = "service";
  let createDialogOpen = false;
  let deleteDialogOpen = false;
  let tokenToDelete: { id: string; name: string; type: TabType } | null = null;

  $: organization = $page.params.organization;

  function handleCreateToken() {
    createDialogOpen = true;
  }

  function handleCloseCreateDialog() {
    createDialogOpen = false;
  }

  function handleRequestDelete(event: CustomEvent<{ id: string; name: string }>) {
    tokenToDelete = {
      id: event.detail.id,
      name: event.detail.name,
      type: activeTab,
    };
    deleteDialogOpen = true;
  }

  function handleCloseDeleteDialog() {
    deleteDialogOpen = false;
    tokenToDelete = null;
  }

  function setActiveTab(tab: TabType) {
    // Only allow switching to "service" for now; "user" is Phase 2
    if (tab === "user") return;
    activeTab = tab;
  }
</script>

<svelte:head>
  <title>Tokens - Settings - {organization}</title>
</svelte:head>

<div class="flex flex-col gap-4 w-full">
  <!-- Page header -->
  <div class="flex items-center justify-between">
    <div class="flex flex-col gap-1">
      <h2 class="text-lg font-semibold text-gray-900">Tokens</h2>
      <p class="text-sm text-gray-500">
        Manage service tokens and personal access tokens for API and CLI access.
      </p>
    </div>
    <Button on:click={handleCreateToken} type="primary">
      Create Token
    </Button>
  </div>

  <!-- Tab bar -->
  <div class="flex border-b border-gray-200">
    <button
      class="tab-button"
      class:active={activeTab === "service"}
      on:click={() => setActiveTab("service")}
    >
      Service Tokens
    </button>
    <button
      class="tab-button tab-disabled"
      disabled
      title="Coming soon"
      on:click={() => setActiveTab("user")}
    >
      User Tokens
      <span class="ml-1.5 rounded-full bg-gray-100 px-2 py-0.5 text-[10px] font-medium text-gray-400 uppercase">
        Soon
      </span>
    </button>
  </div>

  <!-- Tab content -->
  <div class="flex-1">
    {#if activeTab === "service"}
      <ServiceTokenList
        {organization}
        on:create={handleCreateToken}
        on:delete={handleRequestDelete}
      />
    {:else if activeTab === "user"}
      <!-- Phase 2: UserTokenList will be rendered here -->
      <div class="flex items-center justify-center py-16 text-sm text-gray-400">
        User token management coming soon.
      </div>
    {/if}
  </div>
</div>

<!-- Create Service Token Dialog (Sprint 3) -->
<!-- {#if createDialogOpen}
  {#if activeTab === "service"}
    <CreateServiceTokenDialog
      {organization}
      open={createDialogOpen}
      on:close={handleCloseCreateDialog}
    />
  {/if}
{/if} -->

<!-- Delete Token Confirmation Dialog (Sprint 3) -->
<!-- {#if deleteDialogOpen && tokenToDelete}
  <DeleteTokenDialog
    tokenName={tokenToDelete.name}
    tokenType={tokenToDelete.type}
    open={deleteDialogOpen}
    on:close={handleCloseDeleteDialog}
    on:confirm={async () => {
      // Deletion logic handled by the dialog component via mutation
      handleCloseDeleteDialog();
    }}
  />
{/if} -->

<style lang="postcss">
  .tab-button {
    @apply relative px-4 py-2.5 text-sm font-medium text-gray-500 transition-colors;
    @apply hover:text-gray-700 focus:outline-none cursor-pointer;
    @apply border-b-2 border-transparent -mb-px;
  }

  .tab-button.active {
    @apply text-primary-600 border-primary-600;
  }

  .tab-disabled {
    @apply cursor-not-allowed opacity-60 hover:text-gray-500;
  }
</style>