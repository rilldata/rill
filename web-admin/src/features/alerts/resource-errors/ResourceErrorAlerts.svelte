<script lang="ts">
  import { createAdminServiceDeleteAlert } from "@rilldata/web-admin/client";
  import { useAlerts } from "@rilldata/web-admin/features/alerts/selectors";
  import { Button } from "@rilldata/web-common/components/button";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import {
    getRuntimeServiceListResourcesQueryKey,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { untrack } from "svelte";
  import ResourceErrorAlertDialog from "./ResourceErrorAlertDialog.svelte";
  import { getFormValuesFromAlertSpec, isResourceErrorAlert } from "./utils";

  let {
    org,
    project,
    initialResource = undefined,
  }: {
    org: string;
    project: string;
    // Opens the create dialog with this "kind/name" resource pre-selected.
    initialResource?: string;
  } = $props();

  const runtimeClient = useRuntimeClient();
  const queryClient = useQueryClient();
  const alertsQuery = useAlerts(runtimeClient);
  const deleteAlert = createAdminServiceDeleteAlert();

  let createOpen = $state(untrack(() => !!initialResource));
  let editing = $state<V1Resource | undefined>(undefined);
  let editOpen = $state(false);

  let alerts = $derived(
    ($alertsQuery.data?.resources ?? []).filter((r) =>
      isResourceErrorAlert(r.alert?.spec),
    ),
  );

  async function handleDelete(name: string) {
    await $deleteAlert.mutateAsync({ org, project, name });
    await queryClient.invalidateQueries({
      queryKey: getRuntimeServiceListResourcesQueryKey(
        runtimeClient.instanceId,
      ),
    });
  }
</script>

<div class="flex flex-col gap-y-4">
  <div class="flex items-start justify-between gap-x-4">
    <div class="flex flex-col gap-y-1">
      <h2 class="text-lg font-medium text-fg-primary">
        {m.alert_errors_settings_title()}
      </h2>
      <p class="text-sm text-fg-secondary">
        {m.alert_errors_settings_description()}
      </p>
    </div>
    <Button type="primary" onClick={() => (createOpen = true)}>
      {m.alert_errors_add()}
    </Button>
  </div>

  {#if $alertsQuery.isPending}
    <DelayedSpinner isLoading size="1rem" />
  {:else if alerts.length === 0}
    <p class="text-sm text-fg-secondary">{m.alert_errors_empty()}</p>
  {:else}
    <ul class="flex flex-col border rounded-sm divide-y">
      {#each alerts as alert (alert.meta?.name?.name)}
        {@const name = alert.meta?.name?.name ?? ""}
        {@const values = getFormValuesFromAlertSpec(alert.alert?.spec ?? {})}
        <li class="flex items-center gap-x-4 px-4 py-3">
          <div class="flex flex-col gap-y-0.5 grow min-w-0">
            <span class="text-sm font-medium text-fg-primary truncate">
              {values.name}
            </span>
            <span class="text-xs text-fg-secondary truncate">
              {values.resources.length
                ? m.alert_errors_resources_selected({
                    count: values.resources.length,
                  })
                : m.alert_errors_scope_all()}
              ·
              {[
                ...values.emailRecipients,
                ...values.slackChannels,
                ...values.slackUsers,
              ].join(", ")}
            </span>
          </div>
          <Button
            type="secondary"
            onClick={() => {
              editing = alert;
              editOpen = true;
            }}
          >
            {m.alert_errors_edit()}
          </Button>
          <Button
            type="secondary"
            onClick={() => handleDelete(name)}
            disabled={$deleteAlert.isPending}
          >
            {m.common_delete()}
          </Button>
        </li>
      {/each}
    </ul>
  {/if}
</div>

{#if createOpen}
  <ResourceErrorAlertDialog
    {org}
    {project}
    {initialResource}
    bind:open={createOpen}
  />
{/if}

{#if editOpen && editing}
  <ResourceErrorAlertDialog
    {org}
    {project}
    alertName={editing.meta?.name?.name}
    alertSpec={editing.alert?.spec}
    bind:open={editOpen}
  />
{/if}
