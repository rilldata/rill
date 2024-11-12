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
  import { EnvironmentType } from "./types";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import { Trash2Icon } from "lucide-svelte";

  export let open = false;
  export let variableNames: string[] = [];
  export let inputErrors: { [key: number]: boolean } = {};

  let isKeyAlreadyExists = false;
  let isDevelopment = false;
  let isProduction = false;
  let fileInput: HTMLInputElement;

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: hasExistingKeys = Object.values(inputErrors).some((error) => error);
  $: hasNewChanges = $form.newVariables.some(
    (variable) => variable.key !== "" || variable.value !== "",
  );

  const queryClient = useQueryClient();
  const updateProjectVariables = createAdminServiceUpdateProjectVariables();

  const formId = "add-environment-variables-form";

  const initialValues = {
    newVariables: [{ key: "", value: "" }],
  };

  const schema = yup(
    object({
      newVariables: array(
        object({
          key: string().optional(),
          value: string().optional(),
        }),
      ),
    }),
  );

  const { form, enhance, submit, submitting } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      // See: https://superforms.rocks/concepts/nested-data
      dataType: "json",
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        // Omit draft variables that do not have a key
        const filteredVariables = values.newVariables.filter(
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

  async function handleUpdateProjectVariables(
    flatVariables: AdminServiceUpdateProjectVariablesBodyVariables,
  ) {
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

  function handleAdd() {
    $form.newVariables = [...$form.newVariables, { key: "", value: "" }];
  }

  function handleKeyChange(index: number, event: Event) {
    const target = event.target as HTMLInputElement;
    $form.newVariables[index].key = target.value;
  }

  function handleValueChange(index: number, event: Event) {
    const target = event.target as HTMLInputElement;
    $form.newVariables[index].value = target.value;
  }

  function handleRemove(index: number) {
    $form.newVariables = $form.newVariables.filter((_, i) => i !== index);
  }

  function handleReset() {
    $form.newVariables = [{ key: "", value: "" }];
    inputErrors = {};
    isKeyAlreadyExists = false;
  }

  function checkForExistingKeys() {
    const existingKeys = $form.newVariables.map((variable) => variable.key);
    inputErrors = {};
    isKeyAlreadyExists = false;

    existingKeys.forEach((key, index) => {
      // Case sensitive
      if (variableNames.some((existingKey) => existingKey === key)) {
        inputErrors[index] = true;
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
      const [key, value] = line.split("=");
      if (key && value) {
        if (key.trim() && value.trim()) {
          $form.newVariables = [
            ...$form.newVariables,
            { key: key.trim(), value: value.trim() },
          ];
        }
      }
    });
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
          class="w-fit"
          on:click={() => fileInput.click()}
        >
          <span>Import .env file</span>
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
          <div class="flex flex-col gap-y-4 w-full">
            {#each $form.newVariables as variable, index}
              <div
                class="flex flex-row items-center gap-2"
                id={`variable-${index}`}
              >
                <Input
                  bind:value={variable.key}
                  id={`key-${index}`}
                  label=""
                  textClass={inputErrors[index] ? "error-input-wrapper" : ""}
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
                    // Do not allow to remove the last variable
                    if (index !== $form.newVariables.length - 1) {
                      handleRemove(index);
                    } else {
                      handleReset();
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
          {#if isKeyAlreadyExists}
            <div class="mt-1">
              <p class="text-xs text-red-600 font-normal">
                These keys already exist for this environment.
              </p>
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
          handleReset();
        }}>Cancel</Button
      >
      <Button
        type="primary"
        form={formId}
        disabled={$submitting || hasExistingKeys || !hasNewChanges}
        submitForm>Create</Button
      >
    </DialogFooter>
  </DialogContent>
</Dialog>
