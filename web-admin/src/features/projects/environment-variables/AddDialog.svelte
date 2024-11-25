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
  import {
    createAdminServiceUpdateProjectVariables,
    getAdminServiceGetProjectVariablesQueryKey,
  } from "@rilldata/web-admin/client";
  import { type AdminServiceUpdateProjectVariablesBodyVariables } from "@rilldata/web-admin/client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string, array } from "yup";
  import { type VariableNames } from "./types";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import { Trash2Icon, UploadIcon } from "lucide-svelte";
  import {
    getCurrentEnvironment,
    getEnvironmentLabel,
    isDuplicateKey,
  } from "./utils";

  export let open = false;
  export let variableNames: VariableNames = [];

  let inputErrors: { [key: number]: boolean } = {};
  let isKeyAlreadyExists = false;
  let isDevelopment = true;
  let isProduction = true;
  let fileInput: HTMLInputElement;

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: isEnvironmentSelected = isDevelopment || isProduction;
  $: hasExistingKeys = Object.values(inputErrors).some((error) => error);
  $: hasNewChanges = $form.variables.some(
    (variable) => variable.key !== "" || variable.value !== "",
  );

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
          // FIXME: after https://github.com/rilldata/rill/pull/6121
          key: string().optional(),
          value: string().optional(),
        }),
      ),
    }),
  );

  const { form, enhance, submit, submitting, errors, allErrors, reset } =
    superForm(defaults(initialValues, schema), {
      SPA: true,
      validators: schema,
      // See: https://superforms.rocks/concepts/nested-data
      dataType: "json",
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        // Omit draft variables that do not have a key
        const filteredVariables = values.variables.filter(
          ({ key }) => key !== "",
        );

        // Flatten the variables to match the schema
        const flatVariables = Object.fromEntries(
          filteredVariables.map(({ key, value }) => [key, value]),
        );

        try {
          await handleUpdateProjectVariables(flatVariables);
          open = false;
        } catch (error) {
          console.error(error);
        }
      },
    });

  async function handleUpdateProjectVariables(
    flatVariables: AdminServiceUpdateProjectVariablesBodyVariables,
  ) {
    checkForExistingKeys();

    try {
      await $updateProjectVariables.mutateAsync({
        organization,
        project,
        data: {
          environment: getCurrentEnvironment(isDevelopment, isProduction),
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

  function handleAdd() {
    $form.variables = [...$form.variables, { key: "", value: "" }];
  }

  function handleKeyChange(index: number, event: Event) {
    const target = event.target as HTMLInputElement;
    $form.variables[index].key = target.value;
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
    reset();
    isDevelopment = true;
    isProduction = true;
    inputErrors = {};
    isKeyAlreadyExists = false;
  }

  function checkForExistingKeys() {
    inputErrors = {};
    isKeyAlreadyExists = false;

    const existingKeys = $form.variables.map((variable) => {
      return {
        environment: getEnvironmentLabel(
          getCurrentEnvironment(isDevelopment, isProduction),
        ),
        name: variable.key,
      };
    });

    existingKeys.forEach((key, idx) => {
      const variableEnvironment = key.environment;
      const variableKey = key.name;

      if (isDuplicateKey(variableEnvironment, variableKey, variableNames)) {
        inputErrors[idx] = true;
        isKeyAlreadyExists = true;
      }
    });
  }

  function handleFileUpload(event) {
    const file = event.target.files[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => {
        const contents = e.target.result;
        parseFile(contents);
        checkForExistingKeys();
      };
      reader.readAsText(file);
    }
  }

  function parseFile(contents) {
    const lines = contents.split("\n");

    lines.forEach((line) => {
      // Trim the line and check if it starts with '#'
      const trimmedLine = line.trim();
      if (trimmedLine.startsWith("#")) {
        return; // Skip comment lines
      }

      const [key, value] = trimmedLine.split("=");
      if (key && value) {
        if (key.trim() && value.trim()) {
          const filteredVariables = $form.variables.filter(
            (variable) =>
              variable.key.trim() !== "" || variable.value.trim() !== "",
          );

          $form.variables = [
            ...filteredVariables,
            { key: key.trim(), value: value.trim() },
          ];
        }
      }
    });
  }

  function getKeyFromError(error: { path: string; messages: string[] }) {
    return error.path.split("[")[1].split("]")[0];
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
      <DialogTitle>Add environment variables</DialogTitle>
    </DialogHeader>
    <DialogDescription>
      For help, see <a
        href="https://docs.rilldata.com/tutorials/administration/project/credentials-env-variable-management"
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
          on:click={() => fileInput.click()}
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
                  textClass={inputErrors[index] ||
                  ($errors.variables &&
                    $errors.variables[index] &&
                    $errors.variables[index].key)
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
          <Button type="dashed" class="w-full mt-4" on:click={handleAdd}>
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
                  These keys already exist for this project.
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
        form={formId}
        disabled={$submitting ||
          hasExistingKeys ||
          !hasNewChanges ||
          !isEnvironmentSelected}
        submitForm>Create</Button
      >
    </DialogFooter>
  </DialogContent>
</Dialog>
