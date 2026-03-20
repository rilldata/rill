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
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import {
    Dialog,
    DialogContent,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import * as Select from "@rilldata/web-common/components/select";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { Plus, Trash2Icon } from "lucide-svelte";
  import { ORG_ROLES, PROJECT_ROLES, validateServiceName } from "./utils";

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

  $: organization = $page.params.organization;
  $: serviceQuery = createAdminServiceGetService(organization, name, {
    query: { enabled: open },
  });
  $: projectsQuery =
    createAdminServiceListProjectsForOrganization(organization);
  $: allProjects = $projectsQuery.data?.projects ?? [];
  $: assignedProjectNames = new Set(projectAssignments.map((p) => p.project));
  $: availableProjects = allProjects.filter(
    (p) => !assignedProjectNames.has(p.name ?? ""),
  );

  // Initialize form when service data loads
  $: if ($serviceQuery.data?.service && open) {
    const svc = $serviceQuery.data.service;
    const memberships = $serviceQuery.data.projectMemberships ?? [];
    if (!newName) {
      newName = svc.name ?? "";
      orgRole = svc.roleName ?? "viewer";
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
    const current = attributes.filter((a) => a.key.trim());
    const initial = initialAttributes.filter((a) => a.key.trim());
    if (current.length !== initial.length) return true;
    return current.some((a, i) => {
      const init = initial[i];
      return !init || a.key !== init.key || a.value !== init.value;
    });
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
  }

  function addProjectAssignment() {
    if (availableProjects.length === 0) return;
    projectAssignments = [
      ...projectAssignments,
      { project: availableProjects[0].name ?? "", role: "viewer" },
    ];
  }

  function removeProjectAssignment(index: number) {
    projectAssignments = projectAssignments.filter((_, i) => i !== index);
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

      // Remove projects no longer assigned
      for (const [project] of initialMap) {
        if (!currentMap.has(project)) {
          await $removeProjectMember.mutateAsync({
            org: organization,
            project,
            name: effectiveName,
          });
        }
      }

      // Add or update project roles
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
    } catch (error) {
      console.error("Error updating service", error);
      eventBus.emit("notification", {
        message: "Error updating service",
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
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Edit service</DialogTitle>
    </DialogHeader>
    <form
      id="edit-service-form"
      class="w-full flex flex-col gap-y-4 max-h-[60vh] overflow-y-auto"
      on:submit|preventDefault={handleSubmit}
    >
      <Input
        bind:value={newName}
        id="service-name"
        label="Name"
        placeholder="Service name"
        errors={nameError || null}
      />

      <div class="flex flex-col gap-y-1">
        <label for="org-role" class="text-sm font-medium text-fg-primary"
          >Organization role</label
        >
        <Select.Root
          onSelectedChange={(v) => {
            if (v) orgRole = v.value;
          }}
          selected={orgRole ? { value: orgRole, label: orgRole } : undefined}
        >
          <Select.Trigger>
            <Select.Value placeholder="Select a role" />
          </Select.Trigger>
          <Select.Content>
            {#each ORG_ROLES as role}
              <Select.Item value={role}>{role}</Select.Item>
            {/each}
          </Select.Content>
        </Select.Root>
      </div>

      <!-- Project assignments -->
      <div class="flex flex-col gap-y-2">
        <span class="text-sm font-medium text-fg-primary">Project access</span>
        {#each projectAssignments as assignment, index}
          <div class="flex items-center gap-x-2">
            <div class="flex-1">
              <Select.Root
                onSelectedChange={(v) => {
                  if (v) projectAssignments[index].project = v.value;
                }}
                selected={{
                  value: assignment.project,
                  label: assignment.project,
                }}
              >
                <Select.Trigger>
                  <Select.Value placeholder="Select project" />
                </Select.Trigger>
                <Select.Content>
                  {#each allProjects as project}
                    <Select.Item
                      value={project.name ?? ""}
                      disabled={assignedProjectNames.has(project.name ?? "") &&
                        project.name !== assignment.project}
                    >
                      {project.name}
                    </Select.Item>
                  {/each}
                </Select.Content>
              </Select.Root>
            </div>
            <div class="w-28">
              <Select.Root
                onSelectedChange={(v) => {
                  if (v) projectAssignments[index].role = v.value;
                }}
                selected={{
                  value: assignment.role,
                  label: assignment.role,
                }}
              >
                <Select.Trigger>
                  <Select.Value placeholder="Role" />
                </Select.Trigger>
                <Select.Content>
                  {#each PROJECT_ROLES as role}
                    <Select.Item value={role}>{role}</Select.Item>
                  {/each}
                </Select.Content>
              </Select.Root>
            </div>
            <IconButton on:click={() => removeProjectAssignment(index)}>
              <Trash2Icon size="14px" class="text-fg-secondary" />
            </IconButton>
          </div>
        {/each}
        {#if availableProjects.length > 0}
          <Button
            type="secondary"
            small
            class="w-fit"
            onClick={addProjectAssignment}
          >
            <Plus size="14px" />
            <span>Add project</span>
          </Button>
        {/if}
      </div>

      <!-- Custom attributes -->
      <div class="flex flex-col gap-y-2">
        <span class="text-sm font-medium text-fg-primary">Attributes</span>
        {#each attributes as attr, index}
          <div class="flex items-center gap-x-2">
            <Input
              bind:value={attr.key}
              id="attr-key-{index}"
              label=""
              placeholder="Key"
            />
            <Input
              bind:value={attr.value}
              id="attr-value-{index}"
              label=""
              placeholder="Value"
            />
            <IconButton
              on:click={() => {
                attributes = attributes.filter((_, i) => i !== index);
              }}
            >
              <Trash2Icon size="14px" class="text-fg-secondary" />
            </IconButton>
          </div>
        {/each}
        <Button
          type="secondary"
          small
          class="w-fit"
          onClick={() => {
            attributes = [...attributes, { key: "", value: "" }];
          }}
        >
          <Plus size="14px" />
          <span>Add attribute</span>
        </Button>
      </div>
    </form>
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
