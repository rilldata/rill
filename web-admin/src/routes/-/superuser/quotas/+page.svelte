<script lang="ts">
  import OrgPicker from "@rilldata/web-admin/features/superuser/shared/OrgPicker.svelte";
  import ConfirmActionDialog from "@rilldata/web-admin/features/superuser/dialogs/ConfirmActionDialog.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    getOrgForQuotas,
    createUpdateOrgQuotasMutation,
  } from "@rilldata/web-admin/features/superuser/quotas/selectors";
  import { getAdminServiceGetOrganizationQueryKey } from "@rilldata/web-admin/client";
  import { useQueryClient } from "@tanstack/svelte-query";

  let activeOrg = "";

  // Save quotas dialog state
  let saveDialogOpen = false;
  let saveLoading = false;

  const queryClient = useQueryClient();
  const updateOrgQuotas = createUpdateOrgQuotasMutation();

  // Quota fields for editing
  let projects = "";
  let deployments = "";
  let slotsTotal = "";
  let slotsPerDeployment = "";
  let outstandingInvites = "";
  let storageLimitBytes = "";

  $: orgQuery = getOrgForQuotas(activeOrg);

  // Populate fields when org data loads
  $: if ($orgQuery.data?.organization?.quotas) {
    const q = $orgQuery.data.organization.quotas;
    projects = q.projects != null ? String(q.projects) : "";
    deployments = q.deployments != null ? String(q.deployments) : "";
    slotsTotal = q.slotsTotal != null ? String(q.slotsTotal) : "";
    slotsPerDeployment =
      q.slotsPerDeployment != null ? String(q.slotsPerDeployment) : "";
    outstandingInvites =
      q.outstandingInvites != null ? String(q.outstandingInvites) : "";
    storageLimitBytes = q.storageLimitBytesPerDeployment ?? "";
  }

  async function doSaveQuotas() {
    saveLoading = true;
    try {
      await $updateOrgQuotas.mutateAsync({
        data: {
          org: activeOrg,
          projects: projects ? Number(projects) : undefined,
          deployments: deployments ? Number(deployments) : undefined,
          slotsTotal: slotsTotal ? Number(slotsTotal) : undefined,
          slotsPerDeployment: slotsPerDeployment
            ? Number(slotsPerDeployment)
            : undefined,
          outstandingInvites: outstandingInvites
            ? Number(outstandingInvites)
            : undefined,
          storageLimitBytesPerDeployment: storageLimitBytes
            ? storageLimitBytes
            : undefined,
        },
      });
      eventBus.emit("notification", {
        type: "success",
        message: `Quotas updated for org: ${activeOrg}`,
      });
      await queryClient.invalidateQueries({
        queryKey: getAdminServiceGetOrganizationQueryKey(activeOrg),
      });
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to update quotas: ${err}`,
      });
      throw err;
    } finally {
      saveLoading = false;
    }
  }
</script>

<p class="text-sm text-fg-secondary mb-4">
  View and adjust resource quotas for organizations.
</p>

<div class="flex flex-col gap-6">
  <section class="p-5 rounded-lg border">
    <div class="flex gap-3 items-center flex-wrap mb-4">
      <div class="w-64">
        <OrgPicker bind:value={activeOrg} />
      </div>
    </div>

    {#if activeOrg && $orgQuery.isFetching}
      <p class="text-sm text-fg-secondary py-4">Loading quotas...</p>
    {:else if activeOrg && $orgQuery.data?.organization}
      <div class="grid grid-cols-2 lg:grid-cols-3 gap-4">
        <div class="flex flex-col gap-1">
          <label class="text-sm font-medium text-fg-secondary" for="projects"
            >Projects</label
          >
          <input
            id="projects"
            type="number"
            class="px-3 py-2 text-sm rounded-md border bg-input text-fg-primary placeholder:text-fg-muted focus:outline-none focus:ring-2 focus:ring-primary-500"
            bind:value={projects}
          />
        </div>
        <div class="flex flex-col gap-1">
          <label class="text-sm font-medium text-fg-secondary" for="deployments"
            >Deployments</label
          >
          <input
            id="deployments"
            type="number"
            class="px-3 py-2 text-sm rounded-md border bg-input text-fg-primary placeholder:text-fg-muted focus:outline-none focus:ring-2 focus:ring-primary-500"
            bind:value={deployments}
          />
        </div>
        <div class="flex flex-col gap-1">
          <label class="text-sm font-medium text-fg-secondary" for="slotsTotal"
            >Total Slots</label
          >
          <input
            id="slotsTotal"
            type="number"
            class="px-3 py-2 text-sm rounded-md border bg-input text-fg-primary placeholder:text-fg-muted focus:outline-none focus:ring-2 focus:ring-primary-500"
            bind:value={slotsTotal}
          />
        </div>
        <div class="flex flex-col gap-1">
          <label
            class="text-sm font-medium text-fg-secondary"
            for="slotsPerDeployment">Slots per Deployment</label
          >
          <input
            id="slotsPerDeployment"
            type="number"
            class="px-3 py-2 text-sm rounded-md border bg-input text-fg-primary placeholder:text-fg-muted focus:outline-none focus:ring-2 focus:ring-primary-500"
            bind:value={slotsPerDeployment}
          />
        </div>
        <div class="flex flex-col gap-1">
          <label
            class="text-sm font-medium text-fg-secondary"
            for="outstandingInvites">Outstanding Invites</label
          >
          <input
            id="outstandingInvites"
            type="number"
            class="px-3 py-2 text-sm rounded-md border bg-input text-fg-primary placeholder:text-fg-muted focus:outline-none focus:ring-2 focus:ring-primary-500"
            bind:value={outstandingInvites}
          />
        </div>
        <div class="flex flex-col gap-1">
          <label
            class="text-sm font-medium text-fg-secondary"
            for="storageLimitBytes">Storage Limit (bytes)</label
          >
          <input
            id="storageLimitBytes"
            type="text"
            class="px-3 py-2 text-sm rounded-md border bg-input text-fg-primary placeholder:text-fg-muted focus:outline-none focus:ring-2 focus:ring-primary-500"
            bind:value={storageLimitBytes}
          />
        </div>
      </div>

      <div class="mt-4">
        <Button
          large
          class="font-normal"
          type="primary"
          onClick={() => (saveDialogOpen = true)}>Save Quotas</Button
        >
      </div>
    {/if}
  </section>
</div>

<ConfirmActionDialog
  bind:open={saveDialogOpen}
  title="Update Quotas"
  description={`This will update the resource quotas for "${activeOrg}". This change takes effect immediately.`}
  loading={saveLoading}
  onConfirm={doSaveQuotas}
/>
