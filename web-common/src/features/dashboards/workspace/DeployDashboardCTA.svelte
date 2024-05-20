<script lang="ts">
  import { createEventDispatcher, onMount } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";
  import { Button } from "../../../components/button";
  import Dialog from "../../../components/dialog/Dialog.svelte";
  import InputV2 from "../../../components/forms/InputV2.svelte";
  import Tooltip from "../../../components/tooltip/Tooltip.svelte";
  import TooltipContent from "../../../components/tooltip/TooltipContent.svelte";
  import {
    DeployResponse,
    DeployValidationResponse,
  } from "../../../proto/gen/rill/local/v1/api_pb";
  import { localServiceClient } from "../../../runtime-client/local-service/client";

  const dispatch = createEventDispatcher();

  export let open: boolean;

  let deployValidationResponse: DeployValidationResponse;
  let deployResponse: DeployResponse;
  let deployError: string;

  onMount(() => {
    void validate();
  });

  const { form, errors, handleSubmit } = createForm({
    initialValues: {
      orgName: "Default Org Name",
      projectName: "Default Project Name",
    },
    validationSchema: yup.object({
      orgName: yup.string().required("Required"),
      projectName: yup.string().required("Required"),
    }),
    onSubmit: async (values) => {
      try {
        deployResponse = await localServiceClient.deploy({
          rillOrg: values.orgName,
          rillProjectName: values.projectName,
        });
        await validate();
      } catch (e) {
        deployError = e.message;
      }
    },
  });

  async function validate() {
    deployValidationResponse = await localServiceClient.deployValidation({});

    // Set default values for the form
    $form.orgName = deployValidationResponse.rillOrgExistsAsGitUserName
      ? "Default Org Name"
      : deployValidationResponse.gitUserName;
    $form.projectName = deployValidationResponse.localProjectName;
  }

  async function pushToGit() {
    await localServiceClient.pushToGit({});
    await validate();
  }

  function close() {
    dispatch("close");
  }
</script>

<Dialog titleMarginBottomOverride="mb-4" on:close {open}>
  <svelte:fragment slot="title">
    Deploy your project to Rill Cloud
  </svelte:fragment>

  <div class="body" slot="body">
    <section>
      <div class="flex gap-x-2">
        <Button type="primary" on:click={validate}>Validate</Button>
        <Tooltip
          suppress={!deployValidationResponse?.isGithubConnected}
          distance={8}
          location="right"
        >
          <Button
            type="primary"
            disabled={deployValidationResponse?.isGithubConnected}
            on:click={pushToGit}>Push to Git</Button
          >
          <TooltipContent slot="tooltip-content">
            Disabled when the project has already been pushed to Git.
          </TooltipContent>
        </Tooltip>
      </div>
      {#if deployValidationResponse}
        <div class="json-container">
          {JSON.stringify(deployValidationResponse, null, 2)}
        </div>
      {/if}
    </section>
    <section>
      <form id="deploy-project-form" on:submit|preventDefault={handleSubmit}>
        <InputV2
          bind:value={$form["orgName"]}
          error={$errors["orgName"]}
          id="orgName"
          label="Org Name"
        />
        <InputV2
          bind:value={$form["projectName"]}
          error={$errors["projectName"]}
          id="projectName"
          label="Project Name"
        />
        <div>
          <Button type="primary" form="deploy-project-form" submitForm>
            Deploy
          </Button>
        </div>
      </form>
      {#if deployResponse}
        <div class="json-container">
          {JSON.stringify(deployResponse, null, 2)}
        </div>
      {/if}
      {#if deployError}
        <div class="text-red-500 mt-2">
          {deployError}
        </div>
      {/if}
    </section>
  </div>

  <svelte:fragment slot="footer">
    <div class="flex">
      <div class="grow" />
      <Button type="secondary" on:click={close}>Close</Button>
    </div>
  </svelte:fragment>
</Dialog>

<style lang="postcss">
  .body {
    @apply flex flex-col w-full;
  }

  section {
    @apply mb-4;
  }

  .json-container {
    @apply bg-gray-100 p-2 rounded w-full;
    @apply mt-4;
    @apply whitespace-pre-wrap border border-gray-300;
  }

  form {
    @apply flex flex-col gap-4;
  }
</style>
