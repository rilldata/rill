<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceUpdateProjectVariables,
    getAdminServiceGetProjectVariablesQueryKey,
    type AdminServiceUpdateProjectVariablesBodyVariables,
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
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { parse as parseDotenv } from "dotenv";
  import { Plus, Trash2Icon, UploadIcon } from "lucide-svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { array, object, string } from "yup";
  import { type VariableNames } from "./types";
  import { getCurrentEnvironment, isDuplicateKey } from "./utils";

  export let open = false;
  export let variableNames: VariableNames = [];

  let inputErrors: { [key: number]: { type: string } } = {};
  let isKeyAlreadyExists = false;
  let isDevelopment = true;
  let isProduction = true;
  let fileInput: HTMLInputElement;
  let showEnvironmentError = false;

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: hasExistingKeys = Object.keys(inputErrors).length > 0;
  $: hasNewChanges = $form.variables.some(
    (variable) => variable.key !== "" || variable.value !== "",
  );
  $: hasNoEnvironment = showEnvironmentError && !isDevelopment && !isProduction;

  const queryClient = useQueryClient();
  const updateProjectVariables = createAdminServiceUpdateProjectVariables();

  const formId = "add-environment-variables-form";

  const initialValues = {
    variables: [{ key: "", value: "" }],
  };

  const schema = yup(
    object({
      variables: array(
        object({
          key: string()
            .optional()
            .matches(
              /^[a-zA-Z_][a-zA-Z0-9_.]*$/,
              // See: https://github.com/rilldata/rill/pull/6121/files#diff-04140a6ac071a4bac716371f8b66a56c89c9d52cfbf2b05ea1e14ee8d4e301e7R12
              "Key must start with a letter or underscore and can only contain letters, digits, underscores, and dots",
            ),
          value: string().optional(),
        }),
      ),
    }),
  );

  const { form, enhance, submit, submitting, allErrors } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      dataType: "json",
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        // Check for duplicates before proceeding
        const duplicates = checkForExistingKeys();
        if (duplicates > 0) {
          return;
        }

        const filteredVariables = values.variables.filter(
          ({ key }) => key !== "",
        );

        const flatVariables = Object.fromEntries(
          filteredVariables.map(({ key, value }) => [key, value]),
        );

        try {
          await handleUpdateProjectVariables(flatVariables);
          open = false;
          handleReset();
        } catch (error) {
          console.error(error);
        }
      },
    },
  );

  async function handleUpdateProjectVariables(
    flatVariables: AdminServiceUpdateProjectVariablesBodyVariables,
  ) {
    try {
      await $updateProjectVariables.mutateAsync({
        org: organization,
        project,
        data: {
          environment: getCurrentEnvironment(isDevelopment, isProduction),
          variables: flatVariables,
        },
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceGetProjectVariablesQueryKey(
          organization,
          project,
        ),
      });

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

  function handleAdd() {
    $form.variables = [...$form.variables, { key: "", value: "" }];
  }

  function handleKeyChange(index: number, event: Event) {
    const target = event.target as HTMLInputElement;
    $form.variables[index].key = target.value;
    delete inputErrors[index];
    isKeyAlreadyExists = false;
  }

  function handleValueChange(index: number, event: Event) {
    const target = event.target as HTMLInputElement;
    $form.variables[index].value = target.value;
  }

  function handleRemove(index: number) {
    $form.variables = $form.variables.filter((_, i) => i !== index);
    checkForExistingKeys();
  }

  function handleReset() {
    $form = initialValues;
    isDevelopment = true;
    isProduction = true;
    inputErrors = {};
    isKeyAlreadyExists = false;
    showEnvironmentError = false;
  }

  function checkForExistingKeys() {
    inputErrors = {};
    isKeyAlreadyExists = false;
    let isDuplicateWithinForm = false;
    let isDuplicateWithExisting = false;

    // First check for duplicates within the form
    const formKeys = $form.variables
      .filter((variable) => variable.key.trim() !== "")
      .map((variable) => variable.key);

    const formDuplicates = new Set();
    // Check for duplicates using Set
    if (new Set(formKeys).size !== formKeys.length) {
      // Find indices of duplicate keys
      formKeys.forEach((key, index) => {
        if (formKeys.indexOf(key) !== index) {
          // Mark both the original and duplicate entries as errors
          formDuplicates.add(formKeys.indexOf(key));
          formDuplicates.add(index);
          isDuplicateWithinForm = true;
        }
      });
    }

    // Then check against existing variables
    const existingDuplicates = new Set();
    const existingKeys = $form.variables
      .filter((variable) => variable.key.trim() !== "")
      .map((variable) => {
        return {
          environment: getCurrentEnvironment(isDevelopment, isProduction),
          name: variable.key,
        };
      });

    existingKeys.forEach((key, _) => {
      const variableEnvironment = key.environment;
      const variableKey = key.name;

      if (isDuplicateKey(variableEnvironment, variableKey, variableNames)) {
        const originalIndex = $form.variables.findIndex(
          (v) => v.key === variableKey,
        );
        existingDuplicates.add(originalIndex);
        isDuplicateWithExisting = true;
      }
    });

    // Combine the errors
    formDuplicates.forEach((index: number) => {
      inputErrors[index] = { type: "draft" };
    });
    existingDuplicates.forEach((index: number) => {
      inputErrors[index] = { type: "existing" };
    });

    isKeyAlreadyExists = isDuplicateWithinForm || isDuplicateWithExisting;

    return formDuplicates.size + existingDuplicates.size;
  }

  function handleFileUpload(event: Event) {
    const file = (event.target as HTMLInputElement).files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (e: ProgressEvent<FileReader>) => {
        const contents = e.target?.result;
        if (typeof contents === "string") {
          parseFile(contents);
          checkForExistingKeys();
        }
      };
      reader.readAsText(file);
    }
  }

  function parseFile(contents: string) {
    const parsedVariables = parseDotenv(contents);

    for (const [key, value] of Object.entries(parsedVariables)) {
      const filteredVariables = $form.variables.filter(
        (variable) =>
          variable.key.trim() !== "" || variable.value.trim() !== "",
      );

      $form.variables = [...filteredVariables, { key, value }];
    }
  }

  function getKeyFromError(error: { path: string; messages: string[] }) {
    return error.path.split("[")[1].split("]")[0];
  }

  function handleEnvironmentChange() {
    showEnvironmentError = true;
    checkForExistingKeys();
  }

  $: isSubmitDisabled =
    $submitting ||
    hasExistingKeys ||
    !hasNewChanges ||
    hasNoEnvironment ||
    Object.values($form.variables).every((v) => !v.key.trim());
</script>

<Dialog
  bind:open
  onOpenChange={(isOpen) => {
    if (!isOpen) {
      handleReset();
    }
  }}
  onOutsideClick={() => {
    open = false;
    handleReset();
  }}
>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent class="translate-y-[-200px]">
    <DialogHeader>
      <DialogTitle>Add environment variables</DialogTitle>
    </DialogHeader>
    <DialogDescription>
      For help, see <a
        href="https://docs.rilldata.com/manage/project-management/variables-and-credentials"
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
        <Button
          type="secondary"
          small
          class="w-fit flex flex-row items-center gap-x-2"
          onClick={() => fileInput.click()}
        >
          <UploadIcon size="14px" />
          <span>Import .env</span>
        </Button>
        <input
          type="file"
          bind:this={fileInput}
          on:change={handleFileUpload}
          class="hidden"
        />
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
          <div class="text-sm font-medium text-gray-800">Variables</div>
          <!-- 64 (gap 16px * 4) + 160 (item height 32 * 5) = 224 -->
          <div
            class="flex flex-col gap-y-4 w-full overflow-y-auto max-h-[224px]"
          >
            {#each $form.variables as variable, index}
              <div
                class="flex flex-row items-center gap-2"
                id={`variable-${index}`}
              >
                <Input
                  bind:value={variable.key}
                  id={`key-${index}`}
                  label=""
                  textClass={inputErrors[index] &&
                  inputErrors[index].type === "draft"
                    ? "error-input-wrapper"
                    : ""}
                  placeholder="Key"
                  on:input={(e) => handleKeyChange(index, e)}
                  onBlur={() => {
                    checkForExistingKeys();
                  }}
                />
                <Input
                  bind:value={variable.value}
                  id={`value-${index}`}
                  label=""
                  placeholder="Value"
                  on:input={(e) => handleValueChange(index, e)}
                />
                <IconButton
                  on:click={() => {
                    // Reset if there is only one variable
                    if ($form.variables.length === 1) {
                      handleReset();
                    } else {
                      handleRemove(index);
                    }
                  }}
                >
                  <Trash2Icon size="16px" class="text-gray-500" />
                </IconButton>
              </div>
            {/each}
          </div>
          <Button type="dashed" class="w-full mt-4" onClick={handleAdd}>
            <Plus size="16px" />
            <span>Add variable</span>
          </Button>
          <div class="mt-1">
            {#if $allErrors.length}
              <ul class="flex flex-col gap-y-1">
                {#each $allErrors as error}
                  <li>
                    <b>{$form.variables[getKeyFromError(error)].key}</b>
                    <span class="text-xs text-red-600 font-normal">
                      {error.messages}
                    </span>
                  </li>
                {/each}
              </ul>
            {/if}
            {#if isKeyAlreadyExists}
              <div class="mt-1">
                <p class="text-xs text-red-600 font-normal">
                  {#if Object.values(inputErrors).every((err) => err.type === "draft")}
                    {Object.keys(inputErrors).length > 1
                      ? "Duplicate keys are not allowed"
                      : "This key is duplicated"}
                  {:else if Object.values(inputErrors).every((err) => err.type === "existing")}
                    {Object.keys(inputErrors).length > 1
                      ? "These keys already exist for your target environment(s)"
                      : "This key already exists for your target environment(s)"}
                  {:else}
                    Some keys are duplicated or already exist in target
                    environment(s)
                  {/if}
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
        }}
      >
        Cancel
      </Button>
      <Button
        type="primary"
        form={formId}
        disabled={isSubmitDisabled}
        submitForm
      >
        Create
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
