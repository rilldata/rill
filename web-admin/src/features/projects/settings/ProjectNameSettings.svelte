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
  import { sanitizeSlug } from "@rilldata/web-common/lib/string-utils";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { AxiosError } from "axios";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";

  export let organization: string;
  export let project: string;

  // Track if form has been initialized to prevent overwriting user edits
  let formInitialized = false;

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

        const newProject = sanitizeSlug(values.name);

        try {
          await $updateProjectMutation.mutateAsync({
            org: organization,
            project,
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

  // Only sync server data to form on initial load or when form hasn't been modified
  $: if ($projectResp.data?.project && !formInitialized) {
    $form.name = $projectResp.data.project.name ?? "";
    $form.description = $projectResp.data.project.description ?? "";
    formInitialized = true;
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
      description={`Your project will be available at https://ui.rilldata.com/${organization}/${sanitizeSlug($form.name)}.`}
      textClass="text-sm"
      alwaysShowError
      additionalClass="max-w-[520px]"
    />
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
</style>
