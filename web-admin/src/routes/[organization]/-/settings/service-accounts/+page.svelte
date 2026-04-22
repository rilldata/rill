<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceListServices } from "@rilldata/web-admin/client";
  import ServiceTokensTable from "@rilldata/web-admin/features/service-tokens/ServiceTokensTable.svelte";
  import CreateServiceDialog from "@rilldata/web-admin/features/service-tokens/CreateServiceDialog.svelte";
  import ServiceDetailDialog from "@rilldata/web-admin/features/service-tokens/ServiceDetailDialog.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { Plus } from "lucide-svelte";

  let isCreateDialogOpen = $state(false);
  let selectedServiceName = $state("");
  let isDetailDialogOpen = $state(false);
  let searchText = $state("");

  let organization = $derived($page.params.organization);
  let servicesQuery = $derived(createAdminServiceListServices(organization));
  let services = $derived($servicesQuery.data?.services ?? []);

  let filteredServices = $derived(
    searchText
      ? services.filter((s) =>
          (s.name ?? "").toLowerCase().includes(searchText.toLowerCase()),
        )
      : services,
  );

  function handleSelectService(name: string) {
    selectedServiceName = name;
    isDetailDialogOpen = true;
  }
</script>

<div class="flex flex-col gap-y-6">
  <div class="flex flex-col">
    <h2 class="text-lg font-semibold text-fg-primary">Service Accounts</h2>
    <p class="text-sm text-fg-tertiary font-medium">
      Manage your service accounts here. <a
        href="https://docs.rilldata.com/guide/administration/access-tokens/service-tokens"
        target="_blank"
        class="text-primary-600 hover:text-primary-700 active:text-primary-800"
      >
        Learn more ->
      </a>
    </p>
  </div>

  {#if $servicesQuery.isLoading}
    <div class="flex items-center justify-center py-8">
      <Spinner status={EntityStatus.Running} size="24px" />
    </div>
  {:else if $servicesQuery.isError}
    <div
      class="text-sm text-red-500 border border-red-200 rounded p-3 bg-red-50"
    >
      Failed to load service accounts. Please try again.
    </div>
  {:else}
    <div class="flex flex-row gap-x-4">
      <Search
        placeholder="Search"
        bind:value={searchText}
        large
        autofocus={false}
        showBorderOnFocus={false}
        disabled={services.length === 0}
      />
      <Button type="primary" large onClick={() => (isCreateDialogOpen = true)}>
        <Plus size="16px" />
      </Button>
    </div>
    <ServiceTokensTable
      data={filteredServices}
      onSelectService={handleSelectService}
    />
  {/if}
</div>

<CreateServiceDialog bind:open={isCreateDialogOpen} />

{#if selectedServiceName}
  <ServiceDetailDialog
    bind:open={isDetailDialogOpen}
    serviceName={selectedServiceName}
  />
{/if}
