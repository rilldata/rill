<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceListServices } from "@rilldata/web-admin/client";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import ServiceTokensTable from "@rilldata/web-admin/features/service-tokens/ServiceTokensTable.svelte";
  import CreateServiceDialog from "@rilldata/web-admin/features/service-tokens/CreateServiceDialog.svelte";
  import ServiceDetailDialog from "@rilldata/web-admin/features/service-tokens/ServiceDetailDialog.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";

  let isCreateDialogOpen = $state(false);
  let selectedServiceName = $state("");
  let isDetailDialogOpen = $state(false);

  let organization = $derived($page.params.organization);
  let servicesQuery = $derived(createAdminServiceListServices(organization));
  let services = $derived($servicesQuery.data?.services ?? []);

  function handleSelectService(name: string) {
    selectedServiceName = name;
    isDetailDialogOpen = true;
  }
</script>

<SettingsContainer title="Service Accounts">
  <svelte:fragment slot="body">
    <div class="flex flex-col gap-y-4">
      <div class="flex items-center justify-between">
        <p class="text-sm text-fg-tertiary">
          Service accounts for programmatic access to your organization.
        </p>
        <Button
          type="primary"
          small
          onClick={() => {
            isCreateDialogOpen = true;
          }}
        >
          Create service
        </Button>
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
        <ServiceTokensTable
          data={services}
          onSelectService={handleSelectService}
        />
      {/if}
    </div>
  </svelte:fragment>
</SettingsContainer>

<CreateServiceDialog bind:open={isCreateDialogOpen} />

{#if selectedServiceName}
  <ServiceDetailDialog
    bind:open={isDetailDialogOpen}
    serviceName={selectedServiceName}
  />
{/if}
