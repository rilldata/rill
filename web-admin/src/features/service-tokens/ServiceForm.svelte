<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import * as Select from "@rilldata/web-common/components/select";
  import { Plus, Trash2Icon } from "lucide-svelte";
  import type { V1Project } from "@rilldata/web-admin/client";
  import {
    DEFAULT_PROJECT_ROLE,
    ORG_ROLES,
    PROJECT_ROLES,
    capitalize,
    formatOrgRole,
  } from "./utils";

  let {
    name = $bindable(),
    orgRole = $bindable(),
    projectAssignments = $bindable(),
    attributes = $bindable(),
    nameError,
    allProjects,
    formId,
    namePlaceholder = "my-service",
    showOptionalLabels = false,
    onSubmit,
  }: {
    name: string;
    orgRole: string;
    projectAssignments: { project: string; role: string }[];
    attributes: { key: string; value: string }[];
    nameError: string;
    allProjects: V1Project[];
    formId: string;
    namePlaceholder?: string;
    showOptionalLabels?: boolean;
    onSubmit: () => void;
  } = $props();

  let assignedProjectNames = $derived(
    new Set(projectAssignments.map((p) => p.project)),
  );
  let availableProjects = $derived(
    allProjects.filter((p) => !assignedProjectNames.has(p.name ?? "")),
  );

  function addProjectAssignment() {
    if (availableProjects.length === 0) return;
    projectAssignments = [
      ...projectAssignments,
      { project: availableProjects[0].name ?? "", role: DEFAULT_PROJECT_ROLE },
    ];
  }

  function removeProjectAssignment(index: number) {
    projectAssignments = projectAssignments.filter((_, i) => i !== index);
  }
</script>

<form
  id={formId}
  class="w-full flex flex-col gap-y-4 max-h-[50vh] overflow-y-auto"
  onsubmit={(e) => {
    e.preventDefault();
    onSubmit();
  }}
>
  <Input
    bind:value={name}
    id="service-name"
    label="Name"
    placeholder={namePlaceholder}
    errors={nameError || null}
  />

  <div class="flex flex-col gap-y-1">
    <span class="text-sm font-medium text-fg-primary">Organization role</span>
    <span class="text-xs text-fg-tertiary"
      >Applies across all projects. Use "None" for project-only access.</span
    >
    <Select.Root
      type="single"
      value={orgRole}
      onValueChange={(v) => {
        if (v) orgRole = v;
      }}
    >
      <Select.Trigger>
        <span
          class="text-sm {orgRole ? 'text-fg-primary' : 'text-fg-secondary'}"
          >{orgRole ? formatOrgRole(orgRole) : "Select a role"}</span
        >
      </Select.Trigger>
      <Select.Content>
        {#each ORG_ROLES as role}
          <Select.Item value={role}>{formatOrgRole(role)}</Select.Item>
        {/each}
      </Select.Content>
    </Select.Root>
  </div>

  <!-- Project assignments -->
  <div class="flex flex-col gap-y-2">
    <div class="flex flex-col gap-y-0.5">
      <span class="text-sm font-medium text-fg-primary"
        >Project access{showOptionalLabels ? " (optional)" : ""}</span
      >
      <span class="text-xs text-fg-tertiary"
        >Grant this service account access to specific projects with a
        designated role.</span
      >
    </div>
    {#each projectAssignments as assignment, index (assignment.project)}
      <div class="flex items-center gap-x-2">
        <div class="flex-1">
          <Select.Root
            type="single"
            value={assignment.project}
            onValueChange={(v) => {
              if (v) projectAssignments[index].project = v;
            }}
          >
            <Select.Trigger>
              <span class="text-sm text-fg-primary"
                >{assignment.project || "Select project"}</span
              >
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
            type="single"
            value={assignment.role}
            onValueChange={(v) => {
              if (v) projectAssignments[index].role = v;
            }}
          >
            <Select.Trigger>
              <span class="text-sm text-fg-primary"
                >{capitalize(assignment.role)}</span
              >
            </Select.Trigger>
            <Select.Content>
              {#each PROJECT_ROLES as role}
                <Select.Item value={role}>{capitalize(role)}</Select.Item>
              {/each}
            </Select.Content>
          </Select.Root>
        </div>
        <IconButton onclick={() => removeProjectAssignment(index)}>
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
    <div class="flex flex-col gap-y-0.5">
      <span class="text-sm font-medium text-fg-primary"
        >Attributes{showOptionalLabels ? " (optional)" : ""}</span
      >
      <span class="text-xs text-fg-tertiary"
        >Key-value pairs passed to security policies for row-level access
        control.</span
      >
    </div>
    {#each attributes as attr, index (index)}
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
          onclick={() => {
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
