<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceListServices } from "@rilldata/web-admin/client";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import ServiceTokensTable from "@rilldata/web-admin/features/service-tokens/ServiceTokensTable.svelte";
  import CreateServiceDialog from "@rilldata/web-admin/features/service-tokens/CreateServiceDialog.svelte";
  import ServiceDetailDialog from "@rilldata/web-admin/features/service-tokens/ServiceDetailDialog.svelte";
  import { Button } from "@rilldata/web-common/components/button";

  let isCreateDialogOpen = false;
  let selectedServiceName = "";
  let isDetailDialogOpen = false;

  $: organization = $page.params.organization;
  $: servicesQuery = createAdminServiceListServices(organization);
  $: services = $servicesQuery.data?.services ?? [];

  function handleSelectService(name: string) {
    selectedServiceName = name;
    isDetailDialogOpen = true;
  }
</script>

<SettingsContainer title="Service Account Tokens">
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
      <ServiceTokensTable
        data={services}
        onSelectService={handleSelectService}
      />
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
