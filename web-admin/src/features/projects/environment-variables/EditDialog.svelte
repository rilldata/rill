<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceUpdateProjectVariables,
    getAdminServiceGetProjectVariablesQueryKey,
    type AdminServiceUpdateProjectVariablesBodyVariables,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { debounce } from "lodash";
  import { onMount } from "svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import { EnvironmentType, type VariableNames } from "./types";
  import { getCurrentEnvironment, isDuplicateKey } from "./utils";

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
  let isDevelopment = false;
  let isProduction = false;
  let isKeyAlreadyExists = false;
  let inputErrors: { [key: number]: boolean } = {};
  let showEnvironmentError = false;

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: hasNewChanges =
    $form.key !== initialValues.key ||
    $form.value !== initialValues.value ||
    initialEnvironment?.isDevelopment !== isDevelopment ||
    initialEnvironment?.isProduction !== isProduction;
  $: hasExistingKeys = Object.values(inputErrors).some((error) => error);
  $: hasNoEnvironment = showEnvironmentError && !isDevelopment && !isProduction;

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
      key: string()
        .optional()
        .matches(
          /^[a-zA-Z_][a-zA-Z0-9_.]*$/,
          // See: https://github.com/rilldata/rill/pull/6121/files#diff-04140a6ac071a4bac716371f8b66a56c89c9d52cfbf2b05ea1e14ee8d4e301e7R12
          "Key must start with a letter or underscore and can only contain letters, digits, underscores, and dots",
        ),
      value: string().optional(),
    }),
  );

  const { form, enhance, formId, submit, errors, allErrors, submitting } =
    superForm(defaults(initialValues, schema), {
      id: id,
      SPA: true,
      validators: schema,
      resetForm: false,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        checkForExistingKeys();
        if (isKeyAlreadyExists) return;

        const flatVariable = {
          [values.key]: values.value,
        };

        try {
          await handleUpdateProjectVariables(flatVariable);
          open = false;
          handleReset();
        } catch (error) {
          console.error(error);
        }
      },
    });

  async function handleUpdateProjectVariables(
    flatVariable: AdminServiceUpdateProjectVariablesBodyVariables,
  ) {
    // Check if the key has changed, if so, check for existing keys
    if ($form.key !== initialValues.key) {
      checkForExistingKeys();
    }

    try {
      // If the key has changed, remove the old key
      if (initialValues.key !== $form.key) {
        // Unset the old key
        await $updateProjectVariables.mutateAsync({
          org: organization,
          project,
          data: {
            environment: initialValues.environment,
            unsetVariables: [initialValues.key],
          },
        });

        // Update with the new key
        await $updateProjectVariables.mutateAsync({
          org: organization,
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
            org: organization,
            project,
            data: {
              environment: initialValues.environment,
              unsetVariables: [initialValues.key],
            },
          });
        }

        await $updateProjectVariables.mutateAsync({
          org: organization,
          project,
          data: {
            environment: getCurrentEnvironment(isDevelopment, isProduction),
            variables: flatVariable,
          },
        });
      }

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceGetProjectVariablesQueryKey(
          organization,
          project,
        ),
      });

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
    $form.environment = initialValues.environment;
    $form.key = initialValues.key;
    $form.value = initialValues.value;
    isDevelopment = false;
    isProduction = false;
    inputErrors = {};
    isKeyAlreadyExists = false;
    showEnvironmentError = false;
  }

  function checkForExistingKeys() {
    inputErrors = {};
    isKeyAlreadyExists = false;

    const newEnvironment = getCurrentEnvironment(isDevelopment, isProduction);

    if (
      isDuplicateKey(
        newEnvironment,
        $form.key,
        variableNames,
        initialValues.key,
      )
    ) {
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

    initialEnvironment = {
      isDevelopment,
      isProduction,
    };
  }

  function handleDialogOpen() {
    handleReset();
    setInitialCheckboxState();
  }

  function handleEnvironmentChange() {
    showEnvironmentError = true;
  }

  onMount(() => {
    handleDialogOpen();
  });

  $: if (open) {
    handleDialogOpen();
  }

  const debouncedCheckForExistingKeys = debounce(() => {
    checkForExistingKeys();
  }, 500);

  $: if ($form.key) {
    debouncedCheckForExistingKeys();
  }
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
        href="https://docs.rilldata.com/manage/project-management/variables-and-credentials"
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
              onCheckedChange={handleEnvironmentChange}
            />
            <Checkbox
              bind:checked={isProduction}
              id="production"
              label="Production"
              onCheckedChange={handleEnvironmentChange}
            />
          </div>
          {#if hasNoEnvironment}
            <div class="mt-1">
              <p class="text-xs text-red-600 font-normal">
                You must select at least one environment
              </p>
            </div>
          {/if}
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
                  This key already exists for your target environment(s)
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
        onClick={() => {
          open = false;
          handleReset();
        }}>Cancel</Button
      >
      <Button
        type="primary"
        form={$formId}
        disabled={$submitting ||
          !hasNewChanges ||
          hasExistingKeys ||
          hasNoEnvironment}
        submitForm
      >
        Edit
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
