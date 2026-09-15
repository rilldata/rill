<script lang="ts">
  import {
    createAdminServiceGetProject,
    createAdminServiceUpdateProject,
    getAdminServiceGetProjectQueryKey,
    getAdminServiceListDeploymentsQueryKey,
    getAdminServiceListProjectsForOrganizationQueryKey,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import ClusterSize from "@rilldata/web-admin/features/projects/status/overview/ClusterSize.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { AxiosError } from "axios";

  let {
    organization,
    project,
    showDevSlots = false,
  }: {
    organization: string;
    project: string;
    showDevSlots?: boolean;
  } = $props();

  type SlotField = "prodSlots" | "devSlots";
  let fields = $derived<{ field: SlotField; label: string }[]>([
    { field: "prodSlots", label: m.settings_slots_title() },
    ...(showDevSlots
      ? [{ field: "devSlots" as const, label: m.settings_dev_slots_title() }]
      : []),
  ]);
  const updateProjectMutation = createAdminServiceUpdateProject();
  let projectResp = $derived(
    createAdminServiceGetProject(organization, project),
  );
  let projectData = $derived($projectResp.data?.project);
  // Preserve unsaved edits across background project refreshes.
  let editedSlots = $state<Partial<Record<SlotField, string>>>({});
  let slots = $derived({
    prodSlots: Number(editedSlots.prodSlots ?? projectData?.prodSlots),
    devSlots: Number(editedSlots.devSlots ?? projectData?.devSlots),
  });
  let updates = $derived.by(() => {
    const data: Partial<Record<SlotField, string>> = {};
    for (const { field } of fields) {
      if (slots[field] !== Number(projectData?.[field])) {
        data[field] = String(slots[field]);
      }
    }
    return data;
  });
  function validSlots(value: number) {
    return Number.isSafeInteger(value) && value > 0;
  }
  let valid = $derived(fields.every(({ field }) => validSlots(slots[field])));
  let canManage = $derived(
    !!$projectResp.data?.projectPermissions?.manageProject,
  );
  let saving = $state(false);
  let isPending = $derived(saving || $updateProjectMutation.isPending);
  let canSave = $derived(
    canManage && valid && Object.keys(updates).length > 0 && !isPending,
  );
  let error = $state<string | undefined>();

  async function saveSlots() {
    if (!canSave) return;
    error = undefined;
    saving = true;
    try {
      await $updateProjectMutation.mutateAsync({
        org: organization,
        project,
        data: updates,
      });
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: getAdminServiceGetProjectQueryKey(organization, project),
        }),
        queryClient.invalidateQueries({
          queryKey: getAdminServiceListDeploymentsQueryKey(
            organization,
            project,
          ),
        }),
        queryClient.invalidateQueries({
          queryKey:
            getAdminServiceListProjectsForOrganizationQueryKey(organization),
        }),
      ]);
      editedSlots = {};
      eventBus.emit("notification", {
        message: m.settings_deployment_slots_saved(),
      });
    } catch (err) {
      error =
        (err as AxiosError<RpcStatus>).response?.data?.message ??
        m.settings_deployment_slots_failed();
    } finally {
      saving = false;
    }
  }
</script>

<SettingsContainer title={m.settings_deployment_slots_title()}>
  <form
    id="project-slots-form"
    aria-label={m.settings_deployment_slots_title()}
    onsubmit={(event) => {
      event.preventDefault();
      void saveSlots();
    }}
    class="flex flex-col gap-4"
  >
    <p>{m.settings_deployment_slots_description()}</p>
    <div class="grid grid-cols-1 gap-5 {showDevSlots ? 'md:grid-cols-2' : ''}">
      {#each fields as { field, label } (field)}
        <div class="flex flex-col gap-2">
          <Input
            id={field}
            {label}
            inputType="number"
            value={editedSlots[field] ?? projectData?.[field] ?? ""}
            oninput={(event: Event) => {
              editedSlots[field] = (
                event.currentTarget as HTMLInputElement
              ).value;
              error = undefined;
            }}
            disabled={!canManage || isPending}
            errors={editedSlots[field] !== undefined &&
            !validSlots(slots[field])
              ? m.settings_slots_invalid()
              : undefined}
            alwaysShowError
            textClass="text-sm"
            additionalClass="max-w-[260px]"
          />
          {#if validSlots(slots[field])}<ClusterSize
              slots={slots[field]}
            />{/if}
        </div>
      {/each}
    </div>
    <p>{m.settings_slots_restart_description()}</p>
    {#if error}<p role="alert" class="text-red-500">{error}</p>{/if}
  </form>
  {#snippet action()}
    <Button
      type="primary"
      submitForm
      form="project-slots-form"
      disabled={!canSave}
      loading={isPending}
    >
      {m.settings_save_button()}
    </Button>
  {/snippet}
</SettingsContainer>
