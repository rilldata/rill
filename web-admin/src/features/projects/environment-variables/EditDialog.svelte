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
  import { EnvironmentType, type VariableNames } from "./types";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { onMount } from "svelte";
  import {
    getCurrentEnvironment,
    getEnvironmentLabel,
    isDuplicateKey,
  } from "./utils";

  export let open = false;
  export let id: string;
  export let environment: string;
  export let name: string;
  export let value: string;
  export let variableNames: VariableNames = [];

  let initialEnvironment: {
    isDevelopment: boolean;
    isProduction: boolean;
  };
  let isDevelopment: boolean;
  let isProduction: boolean;
  let isKeyAlreadyExists = false;
  let inputErrors: { [key: number]: boolean } = {};

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: isEnvironmentSelected = isDevelopment || isProduction;
  $: hasNewChanges =
    $form.key !== initialValues.key ||
    $form.value !== initialValues.value ||
    initialEnvironment?.isDevelopment !== isDevelopment ||
    initialEnvironment?.isProduction !== isProduction;
  $: hasExistingKeys = Object.values(inputErrors).some((error) => error);

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
      // FIXME: after https://github.com/rilldata/rill/pull/6121
      key: string().optional(),
      value: string().optional(),
    }),
  );

  const {
    form,
    enhance,
    formId,
    submit,
    errors,
    allErrors,
    submitting,
    reset,
  } = superForm(defaults(initialValues, schema), {
    // See: https://superforms.rocks/concepts/multiple-forms#setting-id-on-the-client
    id: id,
    SPA: true,
    validators: schema,
    async onUpdate({ form }) {
      if (!form.valid) return;
      const values = form.data;

      const flatVariable = {
        [values.key]: values.value,
      };

      try {
        await handleUpdateProjectVariables(flatVariable);
        open = false;
      } catch (error) {
        console.error(error);
      }
    },
  });

  // function getCurrentEnvironment() {
  //   if (isDevelopment && isProduction) {
  //     return EnvironmentType.UNDEFINED;
  //   }

  //   if (isDevelopment) {
  //     return EnvironmentType.DEVELOPMENT;
  //   }

  //   if (isProduction) {
  //     return EnvironmentType.PRODUCTION;
  //   }

  //   return EnvironmentType.UNDEFINED;
  // }

  async function handleUpdateProjectVariables(
    flatVariable: AdminServiceUpdateProjectVariablesBodyVariables,
  ) {
    if ($form.key !== initialValues.key) {
      checkForExistingKeys();
    }

    try {
      // If the key has changed, remove the old key
      if (initialValues.key !== $form.key) {
        // Unset the old key
        await $updateProjectVariables.mutateAsync({
          organization,
          project,
          data: {
            environment: initialValues.environment,
            unsetVariables: [initialValues.key],
          },
        });

        // Update with the new key
        await $updateProjectVariables.mutateAsync({
          organization,
          project,
          data: {
            environment: getCurrentEnvironment(isDevelopment, isProduction),
            variables: flatVariable,
          },
        });
      }

      // If the key remains the same, update the environment or value
      if (initialValues.key === $form.key) {
        // If environment has changed, remove the old key and add the new key
        if (
          initialValues.environment !==
          getCurrentEnvironment(isDevelopment, isProduction)
        ) {
          await $updateProjectVariables.mutateAsync({
            organization,
            project,
            data: {
              environment: initialValues.environment,
              unsetVariables: [initialValues.key],
            },
          });
        }

        await $updateProjectVariables.mutateAsync({
          organization,
          project,
          data: {
            environment: getCurrentEnvironment(isDevelopment, isProduction),
            variables: flatVariable,
          },
        });
      }

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

  function handleKeyChange(event: Event) {
    const target = event.target as HTMLInputElement;
    $form.key = target.value;
  }

  function handleValueChange(event: Event) {
    const target = event.target as HTMLInputElement;
    $form.value = target.value;
  }

  function handleReset() {
    reset();
    isKeyAlreadyExists = false;
  }

  function checkForExistingKeys() {
    inputErrors = {};
    isKeyAlreadyExists = false;

    const existingKey = {
      environment: getEnvironmentLabel(
        getCurrentEnvironment(isDevelopment, isProduction),
      ),
      name: $form.key,
    };

    const variableEnvironment = existingKey.environment;
    const variableKey = existingKey.name;

    if (isDuplicateKey(variableEnvironment, variableKey, variableNames)) {
      inputErrors[0] = true;
      isKeyAlreadyExists = true;
    }
  }

  function setInitialCheckboxState() {
    if ($form.environment === EnvironmentType.DEVELOPMENT) {
      isDevelopment = true;
    }

    if ($form.environment === EnvironmentType.PRODUCTION) {
      isProduction = true;
    }

    if ($form.environment === EnvironmentType.UNDEFINED) {
      isDevelopment = true;
      isProduction = true;
    }
  }

  onMount(() => {
    setInitialCheckboxState();
    initialEnvironment = {
      isDevelopment,
      isProduction,
    };
  });
</script>

<Dialog
  bind:open
  onOpenChange={() => handleReset()}
  onOutsideClick={() => handleReset()}
>
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
              id="development"
              label="Development"
            />
            <Checkbox
              bind:checked={isProduction}
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
                textClass={inputErrors[0] || $allErrors[0]
                  ? "error-input-wrapper"
                  : ""}
                placeholder="Key"
                on:input={(e) => handleKeyChange(e)}
                onBlur={() => {
                  if ($form.key !== initialValues.key) {
                    checkForExistingKeys();
                  }
                }}
              />
              <Input
                bind:value={$form.value}
                label=""
                id={`edit-${value}`}
                placeholder="Value"
                on:input={(e) => handleValueChange(e)}
              />
            </div>
            {#if $errors.key}
              <div class="mt-1">
                <p class="text-xs text-red-600 font-normal">
                  {$errors.key}
                </p>
              </div>
            {/if}
            {#if isKeyAlreadyExists}
              <div class="mt-1">
                <p class="text-xs text-red-600 font-normal">
                  This key already exists for this project.
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
      <Button
        type="primary"
        form={$formId}
        disabled={$submitting ||
          !hasNewChanges ||
          !isEnvironmentSelected ||
          hasExistingKeys ||
          $allErrors.length > 0}
        submitForm>Edit</Button
      >
    </DialogFooter>
  </DialogContent>
</Dialog>
