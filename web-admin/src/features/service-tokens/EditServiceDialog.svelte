<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceUpdateService,
    createAdminServiceGetService,
    createAdminServiceSetOrganizationMemberServiceRole,
    createAdminServiceSetProjectMemberServiceRole,
    createAdminServiceRemoveProjectMemberService,
    createAdminServiceListProjectsForOrganization,
    getAdminServiceListServicesQueryKey,
    getAdminServiceGetServiceQueryKey,
    type V1ProjectMemberService,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    Dialog,
    DialogContent,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { isAxiosError } from "axios";
  import { DEFAULT_PROJECT_ROLE, validateServiceName } from "./utils";
  import ServiceForm from "./ServiceForm.svelte";

  let {
    open = $bindable(false),
    name,
  }: {
    open: boolean;
    name: string;
  } = $props();

  let newName = $state("");
  let orgRole = $state("");
  let projectAssignments: { project: string; role: string }[] = $state([]);
  let initialProjectAssignments: { project: string; role: string }[] = $state(
    [],
  );
  let attributes: { key: string; value: string }[] = $state([]);
  let initialAttributes: { key: string; value: string }[] = $state([]);
  let saving = $state(false);
  let initialized = $state(false);

  let organization = $derived($page.params.organization);
  let serviceQuery = $derived(
    createAdminServiceGetService(organization, name, {
      query: { enabled: open },
    }),
  );
  let projectsQuery = $derived(
    createAdminServiceListProjectsForOrganization(organization),
  );
  let allProjects = $derived($projectsQuery.data?.projects ?? []);

  // Initialize form when service data loads
  $effect(() => {
    if ($serviceQuery.data?.service && open && !initialized) {
      const svc = $serviceQuery.data.service;
      const memberships = $serviceQuery.data.projectMemberships ?? [];
      initialized = true;
      newName = svc.name ?? "";
      orgRole = svc.roleName ?? "";
      projectAssignments = memberships.map((pm: V1ProjectMemberService) => ({
        project: pm.projectName ?? "",
        role: pm.projectRoleName ?? DEFAULT_PROJECT_ROLE,
      }));
      initialProjectAssignments = projectAssignments.map((pa) => ({ ...pa }));
      const svcAttrs = (svc.attributes as Record<string, unknown>) ?? {};
      attributes = Object.entries(svcAttrs).map(([key, value]) => ({
        key,
        value: String(value ?? ""),
      }));
      initialAttributes = attributes.map((a) => ({ ...a }));
    }
  });

  let nameError = $derived(newName ? validateServiceName(newName) : "");

  let projectAssignmentsChanged = $derived.by(() => {
    const currentMap = new Map(
      projectAssignments.map((pa) => [pa.project, pa.role]),
    );
    const initialMap = new Map(
      initialProjectAssignments.map((pa) => [pa.project, pa.role]),
    );
    if (currentMap.size !== initialMap.size) return true;
    for (const [project, role] of currentMap) {
      if (initialMap.get(project) !== role) return true;
    }
    return false;
  });

  let attributesChanged = $derived.by(() => {
    const current = new Map(
      attributes
        .filter((a) => a.key.trim())
        .map((a) => [a.key.trim(), a.value]),
    );
    const initial = new Map(
      initialAttributes
        .filter((a) => a.key.trim())
        .map((a) => [a.key.trim(), a.value]),
    );
    if (current.size !== initial.size) return true;
    for (const [k, v] of current) {
      if (initial.get(k) !== v) return true;
    }
    return false;
  });

  let hasChanges = $derived(
    (newName.trim() !== name && !nameError) ||
      orgRole !== ($serviceQuery.data?.service?.roleName ?? "") ||
      projectAssignmentsChanged ||
      attributesChanged,
  );

  const queryClient = useQueryClient();
  const updateService = createAdminServiceUpdateService();
  const setOrgRole = createAdminServiceSetOrganizationMemberServiceRole();
  const setProjectRole = createAdminServiceSetProjectMemberServiceRole();
  const removeProjectMember = createAdminServiceRemoveProjectMemberService();

  function handleReset() {
    newName = "";
    orgRole = "";
    projectAssignments = [];
    initialProjectAssignments = [];
    attributes = [];
    initialAttributes = [];
    saving = false;
    initialized = false;
  }

  async function handleSubmit() {
    saving = true;
    try {
      const currentName = name;

      // Update name and/or attributes if changed
      const nameChanged = newName.trim() !== currentName;
      const attrObj = Object.fromEntries(
        attributes
          .filter((a) => a.key.trim())
          .map((a) => [a.key.trim(), a.value]),
      );
      if (nameChanged || attributesChanged) {
        await $updateService.mutateAsync({
          org: organization,
          name: currentName,
          data: {
            ...(nameChanged ? { newName: newName.trim() } : {}),
            ...(attributesChanged ? { attributes: attrObj } : {}),
          },
        });
      }

      const effectiveName = newName.trim() || currentName;

      // Update org role if changed
      const currentOrgRole = $serviceQuery.data?.service?.roleName ?? "";
      if (orgRole !== currentOrgRole) {
        await $setOrgRole.mutateAsync({
          org: organization,
          name: effectiveName,
          data: { role: orgRole },
        });
      }

      // Handle project role changes in parallel. These are not atomic:
      // partial updates are possible if individual calls fail.
      const initialMap = new Map(
        initialProjectAssignments.map((pa) => [pa.project, pa.role]),
      );
      const currentMap = new Map(
        projectAssignments.map((pa) => [pa.project, pa.role]),
      );

      const removals = [...initialMap.keys()]
        .filter((project) => !currentMap.has(project))
        .map((project) =>
          $removeProjectMember.mutateAsync({
            org: organization,
            project,
            name: effectiveName,
          }),
        );

      const upserts = [...currentMap.entries()]
        .filter(([project, role]) => initialMap.get(project) !== role)
        .map(([project, role]) =>
          $setProjectRole.mutateAsync({
            org: organization,
            project,
            name: effectiveName,
            data: { role },
          }),
        );

      const results = await Promise.allSettled([...removals, ...upserts]);
      const failed = results.filter((r) => r.status === "rejected");
      if (failed.length > 0) {
        console.warn(`${failed.length} project role update(s) failed`, failed);
      }

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListServicesQueryKey(organization),
      });
      await queryClient.invalidateQueries({
        queryKey: getAdminServiceGetServiceQueryKey(
          organization,
          effectiveName,
        ),
      });

      eventBus.emit("notification", { message: "Service updated" });
      open = false;
    } catch (e) {
      console.error("Error updating service", e);
      eventBus.emit("notification", {
        message: isAxiosError(e)
          ? (e.response?.data?.message ?? "Error updating service")
          : "Error updating service",
        type: "error",
      });
    } finally {
      saving = false;
    }
  }
</script>

<Dialog
  bind:open
  onOpenChange={(isOpen) => {
    if (!isOpen) handleReset();
  }}
>
  <DialogTrigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </DialogTrigger>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Edit service</DialogTitle>
    </DialogHeader>
    {#if $serviceQuery.isPending}
      <div class="flex items-center justify-center py-8">
        <Spinner status={EntityStatus.Running} size="24px" />
      </div>
    {:else}
      <ServiceForm
        bind:name={newName}
        bind:orgRole
        bind:projectAssignments
        bind:attributes
        {nameError}
        {allProjects}
        formId="edit-service-form"
        namePlaceholder="Service name"
        onSubmit={handleSubmit}
      />
      <DialogFooter>
        <Button
          type="tertiary"
          onClick={() => {
            open = false;
            handleReset();
          }}
        >
          Cancel
        </Button>
        <Button
          type="primary"
          form="edit-service-form"
          disabled={!hasChanges || saving || !!nameError}
          submitForm
        >
          Save
        </Button>
      </DialogFooter>
    {/if}
  </DialogContent>
</Dialog>
