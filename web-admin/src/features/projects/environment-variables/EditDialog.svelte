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
  import {
    createAdminServiceUpdateProjectVariables,
    getAdminServiceGetProjectVariablesQueryKey,
  } from "@rilldata/web-admin/client";
  import { type AdminServiceUpdateProjectVariablesBodyVariables } from "@rilldata/web-admin/client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import { EnvironmentType, type EnvironmentTypes } from "./types";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { onMount } from "svelte";

  export let open = false;
  export let id: string;
  export let environment: string;
  export let name: string;
  export let value: string;

  let isDevelopment = false;
  let isProduction = false;
  let processedEnvironment: EnvironmentTypes;

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  const queryClient = useQueryClient();
  const updateProjectVariables = createAdminServiceUpdateProjectVariables();

  const initialValues = {
    environment: environment,
    key: name,
    value: value,
  };

  const schema = yup(
    object({
      environment: string().optional(),
      key: string().required("Key is required"),
      value: string().optional(),
    }),
  );

  const { form, enhance, formId, submit, errors, submitting } = superForm(
    defaults(initialValues, schema),
    {
      // See: https://superforms.rocks/concepts/multiple-forms#setting-id-on-the-client
      id: id,
      SPA: true,
      validators: schema,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        const flatVariables = {
          [values.key]: values.value,
        };

        try {
          await handleUpdateProjectVariables(flatVariables);
          open = false;
        } catch (error) {
          console.error(error);
        }
      },
    },
  );

  function processFormEnvironment() {
    if (!$form.environment && !$form.environment) {
      return "";
    } else if ($form.environment === EnvironmentType.DEVELOPMENT) {
      return EnvironmentType.DEVELOPMENT;
    } else if ($form.environment === EnvironmentType.PRODUCTION) {
      return EnvironmentType.PRODUCTION;
    }
  }
  $: processedEnvironment = processFormEnvironment();

  async function handleUpdateProjectVariables(
    flatVariable: AdminServiceUpdateProjectVariablesBodyVariables,
  ) {
    try {
      await $updateProjectVariables.mutateAsync({
        organization,
        project,
        data: {
          environment: processedEnvironment,
          variables: flatVariable,
        },
      });

      await queryClient.invalidateQueries(
        getAdminServiceGetProjectVariablesQueryKey(organization, project),
      );

      eventBus.emit("notification", {
        message: "Environment variable updated",
      });
    } catch (error) {
      console.error("Error updating project variable", error);
      eventBus.emit("notification", {
        message: "Error updating project variable",
        type: "error",
      });
    }
  }

  function handleValueChange(e: any) {
    $form.value = e.target.value;
  }

  function handleReset() {
    $form = initialValues;

    isDevelopment = true;
    isProduction = true;
  }

  onMount(() => {
    if ($form.environment === "") {
      isDevelopment = true;
      isProduction = true;
    }
  });

  // $: hasChanges =
  //   $form.value !== value || $form.environment !== processedEnvironment;

  // $: console.log("$form: ", $form);
</script>

<Dialog bind:open onOutsideClick={() => handleReset()}>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent class="translate-y-[-200px]">
    <DialogHeader>
      <DialogTitle>Edit environment variable</DialogTitle>
    </DialogHeader>
    <DialogDescription>
      For help, see <a
        href="https://docs.rilldata.com/tutorials/administration/project/credentials-env-variable-management"
        target="_blank">documentation</a
      >
    </DialogDescription>
    <form
      id={$formId}
      class="w-full"
      on:submit|preventDefault={submit}
      use:enhance
    >
      <div class="flex flex-col gap-y-5">
        <div class="flex flex-col items-start gap-1">
          <div class="text-sm font-medium text-gray-800">Environment</div>
          <div class="flex flex-row gap-4 mt-1">
            <Checkbox
              bind:checked={isDevelopment}
              onCheckedChange={(checked) => {
                // if (checked) {
                //   isDevelopment = true;
                // } else {
                //   isDevelopment = false;
                // }
                isDevelopment = checked;
              }}
              id="development"
              label="Development"
            />
            <Checkbox
              bind:checked={isProduction}
              onCheckedChange={(checked) => {
                // if (checked) {
                //   isProduction = true;
                // } else {
                //   isProduction = false;
                // }
                isProduction = checked;
              }}
              id="production"
              label="Production"
            />
          </div>
        </div>
        <div class="flex flex-col items-start gap-1">
          <div class="text-sm font-medium text-gray-800">Variable</div>
          <div class="flex flex-col w-full">
            <div class="flex flex-row items-center gap-2">
              <Input
                bind:value={$form.key}
                label=""
                id={`edit-${name}`}
                placeholder="Key"
                readonly
              />
              <Input
                bind:value={$form.value}
                label=""
                id={`edit-${value}`}
                placeholder="Value"
                on:input={(e) => handleValueChange(e.target.value)}
              />
            </div>
            {#if $errors.key || $errors.value || $errors.environment}
              <div class="mt-1">
                <p class="text-xs text-red-600 font-normal">
                  {$errors.key || $errors.value || $errors.environment}
                </p>
              </div>
            {/if}
          </div>
        </div>
      </div>
    </form>

    <DialogFooter>
      <Button
        type="plain"
        on:click={() => {
          open = false;
          handleReset();
        }}>Cancel</Button
      >
      <Button type="primary" form={$formId} disabled={$submitting} submitForm
        >Edit</Button
      >
    </DialogFooter>
  </DialogContent>
</Dialog>
