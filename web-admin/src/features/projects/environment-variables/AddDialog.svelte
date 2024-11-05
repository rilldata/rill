<script lang="ts">
  import { page } from "$app/stores";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
    DialogDescription,
    DialogFooter,
  } from "@rilldata/web-common/components/dialog-v2";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import { Plus } from "lucide-svelte";
  import ErrorMessage from "./ErrorMessage.svelte";
  import KeyValueItem from "./KeyValueItem.svelte";
  import {
    createAdminServiceUpdateProjectVariables,
    getAdminServiceGetProjectVariablesQueryKey,
  } from "@rilldata/web-admin/client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string, array } from "yup";
  import { EnvironmentType } from "./types";

  export let open = false;

  let errorMessage = "";
  let isDevelopment = false;
  let isProduction = false;
  let newVariables: { key: string; value: string }[] = [
    { key: "CLIENT_KEY", value: "test_value" },
  ];

  $: console.log(newVariables);

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  const queryClient = useQueryClient();
  const updateProjectVariables = createAdminServiceUpdateProjectVariables();

  const formId = "add-environment-variables-form";

  const initialValues = {
    newVariables,
  };

  const schema = yup(
    object({
      newVariables: array(
        object({
          key: string().required("Key is required"),
          value: string().required("Value is required"),
        }),
      ),
    }),
  );

  const { form, enhance, submit, errors, submitting } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      // https://superforms.rocks/concepts/nested-data
      dataType: "json",
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        const flatVariables = Object.fromEntries(
          values.newVariables.map(({ key, value }) => [key, value]),
        );

        console.log("flatVariables: ", flatVariables);

        try {
          await handleUpdateProjectVariables(flatVariables);
          open = false;
        } catch (error) {
          console.error(error);
        }
      },
    },
  );

  function processEnvironment() {
    return isDevelopment
      ? EnvironmentType.DEVELOPMENT
      : isProduction
        ? EnvironmentType.PRODUCTION
        : // If empty, the variable(s) will be used as defaults for all environments.
          undefined;
  }

  async function handleUpdateProjectVariables(flatVariables: {
    [key: string]: string;
  }) {
    try {
      await $updateProjectVariables.mutateAsync({
        organization,
        project,
        data: {
          environment: processEnvironment(),
          variables: flatVariables,
        },
      });

      await queryClient.invalidateQueries(
        getAdminServiceGetProjectVariablesQueryKey(organization, project),
      );

      eventBus.emit("notification", {
        message: "Environment variables updated",
      });
    } catch (error) {
      console.error("Error updating project variables", error);
      eventBus.emit("notification", {
        message: "Error updating project variables",
        type: "error",
      });
    }
  }

  function handleDelete(index: number) {
    newVariables = newVariables.filter((_, i) => i !== index);
  }
</script>

<Dialog bind:open>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent class="translate-y-[-200px]">
    <DialogHeader>
      <DialogTitle>Add environment variables</DialogTitle>
    </DialogHeader>
    <DialogDescription>
      For help, see <a
        href="https://docs.rilldata.com/tutorials/administration/project/credential-envvariable-mangement"
        target="_blank">documentation</a
      >
    </DialogDescription>
    <form
      id={formId}
      class="w-full"
      on:submit|preventDefault={submit}
      use:enhance
    >
      <div class="flex flex-col gap-y-5">
        <!-- TODO: onclick to trigger file upload, parse the content -->
        <Button type="secondary" small class="w-fit">
          <span>Import .env file</span>
        </Button>
        <div class="flex flex-col items-start gap-1">
          <div class="text-sm font-medium text-gray-800">Environment</div>
          <div class="flex flex-row gap-4 mt-1">
            <!-- TODO: check the usage before changing the label color to text-gray-800 -->
            <Checkbox
              inverse
              bind:checked={isDevelopment}
              id="development"
              label="Development"
            />
            <Checkbox
              inverse
              bind:checked={isProduction}
              id="production"
              label="Production"
            />
          </div>
        </div>
        <div class="flex flex-col items-start gap-1">
          <div class="text-sm font-medium text-gray-800">Variables</div>
          <div class="flex flex-col gap-y-4 w-full">
            {#each newVariables as variable, index}
              <KeyValueItem
                {variable}
                {index}
                on:delete={() => handleDelete(index)}
              />
            {/each}
          </div>
          <Button
            type="dashed"
            class="w-full mt-4"
            on:click={() => {
              newVariables = [...newVariables, { key: "", value: "" }];
            }}
          >
            <Plus size="16px" />
            <span>Add variable</span>
          </Button>
          {#if errorMessage}
            <div class="mt-1">
              <ErrorMessage />
            </div>
          {/if}
        </div>
      </div>
    </form>

    <DialogFooter>
      <Button
        type="plain"
        on:click={() => {
          open = false;
        }}>Cancel</Button
      >
      <Button type="primary" form={formId} disabled={$submitting} submitForm
        >Create</Button
      >
    </DialogFooter>
  </DialogContent>
</Dialog>
