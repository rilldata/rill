<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    createAdminServiceGetProject,
    createAdminServiceUpdateProject,
    getAdminServiceGetProjectQueryKey,
    getAdminServiceListProjectsForOrganizationQueryKey,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import { parseUpdateProjectError } from "@rilldata/web-admin/features/projects/settings/errors";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { sanitizeOrgName } from "@rilldata/web-common/features/organization/sanitizeOrgName";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { AxiosError } from "axios";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";

  export let organization: string;
  export let project: string;

  // Reuse org name sanitizer for project names
  const sanitizeProjectName = sanitizeOrgName;

  const initialValues: {
    name: string;
    description: string;
  } = {
    name: "",
    description: "",
  };
  const schema = yup(
    object({
      name: string().required(),
      description: string(),
    }),
  );

  const updateProjectMutation = createAdminServiceUpdateProject();

  const { form, errors, enhance, submit } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        const newProject = sanitizeProjectName(values.name);

        try {
          await $updateProjectMutation.mutateAsync({
            org: organization,
            project: project,
            data: {
              newName: newProject,
              description: values.description,
            },
          });

          await queryClient.invalidateQueries({
            queryKey:
              getAdminServiceListProjectsForOrganizationQueryKey(organization),
          });
        } catch (err) {
          const parsedErr = parseUpdateProjectError(
            err as AxiosError<RpcStatus>,
          );
          if (parsedErr.duplicateProject) {
            form.errors.name = [`The name ${newProject} is already taken`];
          }
          return;
        }

        if (project !== newProject) {
          queryClient.removeQueries({
            queryKey: getAdminServiceGetProjectQueryKey(organization, project),
          });
          setTimeout(() => goto(`/${organization}/${newProject}/-/settings`));
        } else {
          void queryClient.refetchQueries({
            queryKey: getAdminServiceGetProjectQueryKey(organization, project),
          });
        }
        eventBus.emit("notification", {
          message: "Updated project",
        });
      },
      resetForm: false,
    },
  );

  $: projectResp = createAdminServiceGetProject(organization, project);
  $: if ($projectResp.data?.project) {
    $form.name = $projectResp.data.project.name ?? "";
    $form.description = $projectResp.data.project.description ?? "";
  }

  $: changed =
    $projectResp.data?.project?.name !== $form.name ||
    $projectResp.data?.project?.description !== $form.description;

  $: error = parseUpdateProjectError(
    $updateProjectMutation.error as unknown as AxiosError<RpcStatus>,
  );
</script>

<SettingsContainer title="Project">
  <form
    slot="body"
    id="project-update-form"
    on:submit|preventDefault={submit}
    class="update-project-form"
    use:enhance
  >
    <Input
      bind:value={$form.name}
      errors={$errors?.name}
      id="name"
      label="Name"
      description={`Your project URL will be https://ui.rilldata.com/${organization}/${sanitizeProjectName($form.name)}, to comply with our naming rules.`}
      textClass="text-sm"
      alwaysShowError
      additionalClass="max-w-[520px]"
    />
    {#if $form.name && sanitizeProjectName($form.name) !== project}
      <div class="warning-message">
        Renaming this project will invalidate all existing URLs and shared
        links.
      </div>
    {/if}
    <Input
      bind:value={$form.description}
      errors={$errors?.description}
      id="description"
      label="Description"
      placeholder="Describe your project"
      textClass="text-sm"
      additionalClass="max-w-[520px]"
    />
  </form>
  {#if error?.message}
    <div class="text-red-500 text-sm py-px">
      {error.message}
    </div>
  {/if}
  <Button
    onClick={submit}
    type="primary"
    loading={$updateProjectMutation.isPending}
    disabled={!changed}
    slot="action"
  >
    Save
  </Button>
</SettingsContainer>

<style lang="postcss">
  .update-project-form {
    @apply flex flex-col gap-y-5 w-full;
  }

  .warning-message {
    @apply text-sm text-yellow-600 bg-yellow-50 border border-yellow-200 rounded px-3 py-2;
  }
</style>
