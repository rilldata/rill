<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceCreateService,
    createAdminServiceIssueServiceAuthToken,
    createAdminServiceSetProjectMemberServiceRole,
    createAdminServiceListProjectsForOrganization,
    getAdminServiceListServicesQueryKey,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import * as Select from "@rilldata/web-common/components/select";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { CopyIcon, Plus, Trash2Icon } from "lucide-svelte";
  import { ORG_ROLES, PROJECT_ROLES, validateServiceName } from "./utils";

  export let open = false;

  let name = "";
  let orgRole = "viewer";
  let projectAssignments: { project: string; role: string }[] = [];
  let attributes: { key: string; value: string }[] = [];
  let issuedToken = "";
  let step: "form" | "token" = "form";
  let nameError = "";

  $: organization = $page.params.organization;
  $: projectsQuery =
    createAdminServiceListProjectsForOrganization(organization);
  $: allProjects = $projectsQuery.data?.projects ?? [];
  $: assignedProjectNames = new Set(projectAssignments.map((p) => p.project));
  $: availableProjects = allProjects.filter(
    (p) => !assignedProjectNames.has(p.name ?? ""),
  );

  $: nameError = name ? validateServiceName(name) : "";
  $: isValid = name.trim() !== "" && !nameError;

  const queryClient = useQueryClient();
  const createService = createAdminServiceCreateService();
  const issueToken = createAdminServiceIssueServiceAuthToken();
  const setProjectRole = createAdminServiceSetProjectMemberServiceRole();

  function handleReset() {
    name = "";
    orgRole = "viewer";
    projectAssignments = [];
    attributes = [];
    issuedToken = "";
    step = "form";
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
    try {
      // Create the service with the first project assignment (if any)
      const firstProject = projectAssignments[0];
      const attrObj = Object.fromEntries(
        attributes
          .filter((a) => a.key.trim())
          .map((a) => [a.key.trim(), a.value]),
      );

      await $createService.mutateAsync({
        org: organization,
        data: {
          name: name.trim(),
          orgRoleName: orgRole,
          ...(firstProject
            ? {
                project: firstProject.project,
                projectRoleName: firstProject.role,
              }
            : {}),
          ...(Object.keys(attrObj).length > 0 ? { attributes: attrObj } : {}),
        },
      });

      // Add remaining project assignments
      for (let i = 1; i < projectAssignments.length; i++) {
        const pa = projectAssignments[i];
        await $setProjectRole.mutateAsync({
          org: organization,
          project: pa.project,
          name: name.trim(),
          data: { role: pa.role },
        });
      }

      // Auto-issue a token
      const result = await $issueToken.mutateAsync({
        org: organization,
        serviceName: name.trim(),
        data: {},
      });

      issuedToken = result.token ?? "";

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListServicesQueryKey(organization),
      });

      step = "token";
    } catch (error) {
      console.error("Error creating service", error);
      eventBus.emit("notification", {
        message: "Error creating service",
        type: "error",
      });
    }
  }

  function handleClose() {
    open = false;
    handleReset();
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
      <DialogTitle>
        {step === "form" ? "Create service" : "Service created"}
      </DialogTitle>
    </DialogHeader>

    {#if step === "form"}
      <DialogDescription>
        Create a service account to access Rill programmatically.
      </DialogDescription>
      <form
        id="create-service-form"
        class="w-full flex flex-col gap-y-4 max-h-[45vh] overflow-y-auto"
        on:submit|preventDefault={handleSubmit}
      >
        <Input
          bind:value={name}
          id="service-name"
          label="Name"
          placeholder="my-service"
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
            selected={{ value: orgRole, label: orgRole }}
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
          <span class="text-sm font-medium text-fg-primary"
            >Project roles (optional)</span
          >
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
                        disabled={assignedProjectNames.has(
                          project.name ?? "",
                        ) && project.name !== assignment.project}
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
          <span class="text-sm font-medium text-fg-primary"
            >Attributes (optional)</span
          >
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
        <Button type="tertiary" onClick={handleClose}>Cancel</Button>
        <Button
          type="primary"
          form="create-service-form"
          disabled={!isValid || $createService.isPending}
          submitForm
        >
          Create
        </Button>
      </DialogFooter>
    {:else}
      <!-- Token display step -->
      <div class="flex flex-col gap-y-4">
        <p class="text-sm text-fg-tertiary">
          Service <span class="font-medium text-fg-primary">{name}</span> has been
          created. Copy the token below — it will not be shown again.
        </p>
        <div class="flex items-center gap-x-2">
          <code
            class="text-xs bg-surface-subtle border rounded px-2 py-2 flex-1 break-all select-all"
          >
            {issuedToken}
          </code>
          <IconButton on:click={() => copyToClipboard(issuedToken)}>
            <CopyIcon size="14px" />
          </IconButton>
        </div>
        <p class="text-xs text-fg-secondary">
          This token will only be shown once. Make sure to copy it now.
        </p>
      </div>
      <DialogFooter>
        <Button type="primary" onClick={handleClose}>Done</Button>
      </DialogFooter>
    {/if}
  </DialogContent>
</Dialog>
