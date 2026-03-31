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
  import { useQueryClient } from "@tanstack/svelte-query";
  import { validateServiceName } from "./utils";
  import ServiceForm from "./ServiceForm.svelte";

  export let open = false;
  export let name: string;

  let newName = "";
  let orgRole = "";
  let projectAssignments: { project: string; role: string }[] = [];
  let initialProjectAssignments: { project: string; role: string }[] = [];
  let attributes: { key: string; value: string }[] = [];
  let initialAttributes: { key: string; value: string }[] = [];
  let saving = false;
  let nameError = "";
  let initialized = false;

  $: organization = $page.params.organization;
  $: serviceQuery = createAdminServiceGetService(organization, name, {
    query: { enabled: open },
  });
  $: projectsQuery =
    createAdminServiceListProjectsForOrganization(organization);
  $: allProjects = $projectsQuery.data?.projects ?? [];

  // Initialize form when service data loads
  $: if ($serviceQuery.data?.service && open && !initialized) {
    const svc = $serviceQuery.data.service;
    const memberships = $serviceQuery.data.projectMemberships ?? [];
    initialized = true;
    newName = svc.name ?? "";
    orgRole = svc.roleName ?? "";
    projectAssignments = memberships.map((pm: V1ProjectMemberService) => ({
      project: pm.projectName ?? "",
      role: pm.projectRoleName ?? "viewer",
    }));
    initialProjectAssignments = projectAssignments.map((pa) => ({ ...pa }));
    const svcAttrs = (svc.attributes as Record<string, unknown>) ?? {};
    attributes = Object.entries(svcAttrs).map(([key, value]) => ({
      key,
      value: String(value ?? ""),
    }));
    initialAttributes = attributes.map((a) => ({ ...a }));
  }

  $: nameError = newName ? validateServiceName(newName) : "";

  $: hasChanges =
    (newName.trim() !== name && !nameError) ||
    orgRole !== ($serviceQuery.data?.service?.roleName ?? "") ||
    projectAssignmentsChanged ||
    attributesChanged;

  $: projectAssignmentsChanged = (() => {
    if (projectAssignments.length !== initialProjectAssignments.length)
      return true;
    return projectAssignments.some((pa, i) => {
      const initial = initialProjectAssignments[i];
      return (
        !initial || pa.project !== initial.project || pa.role !== initial.role
      );
    });
  })();

  $: attributesChanged = (() => {
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
  })();

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
    nameError = "";
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

      // Handle project role changes
      const initialMap = new Map(
        initialProjectAssignments.map((pa) => [pa.project, pa.role]),
      );
      const currentMap = new Map(
        projectAssignments.map((pa) => [pa.project, pa.role]),
      );

      for (const [project] of initialMap) {
        if (!currentMap.has(project)) {
          await $removeProjectMember.mutateAsync({
            org: organization,
            project,
            name: effectiveName,
          });
        }
      }

      for (const [project, role] of currentMap) {
        const initialRole = initialMap.get(project);
        if (initialRole !== role) {
          await $setProjectRole.mutateAsync({
            org: organization,
            project,
            name: effectiveName,
            data: { role },
          });
        }
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
    } catch (e: any) {
      console.error("Error updating service", e);
      eventBus.emit("notification", {
        message: e?.response?.data?.message ?? "Error updating service",
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
  </DialogContent>
</Dialog>
