<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import {
    Dialog,
    DialogContent,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { Plus, Trash2Icon, UploadIcon } from "lucide-svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { array, object, string } from "yup";
  import { createEventDispatcher } from "svelte";
  import { isDuplicateKey } from "./utils";
  import type { EnvVariable } from "./types";

  export let open = false;
  export let existingVariables: EnvVariable[] = [];

  const dispatch = createEventDispatcher<{
    add: { variables: EnvVariable[] };
  }>();

  let inputErrors: { [key: number]: { type: string } } = {};
  let isKeyAlreadyExists = false;
  let fileInput: HTMLInputElement;

  $: hasExistingKeys = Object.keys(inputErrors).length > 0;
  $: hasNewChanges = $form.variables.some(
    (variable) => variable.key !== "" || variable.value !== "",
  );

  const formId = "add-environment-variables-form";

  const initialValues = {
    variables: [{ key: "", value: "" }],
  };

  const schema = yup(
    object({
      variables: array(
        object({
          key: string()
            .required()
            .default("")
            .matches(
              /^[a-zA-Z_][a-zA-Z0-9_.]*$/,
              "Key must start with a letter or underscore and can only contain letters, digits, underscores, and dots",
            ),
          value: string().required().default(""),
        }),
      ).required(),
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

        const duplicates = checkForExistingKeys();
        if (duplicates > 0) {
          return;
        }

        const filteredVariables = values.variables.filter(
          ({ key }) => key !== "",
        );

        dispatch("add", { variables: filteredVariables });
        open = false;
        handleReset();
      },
    },
  );

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
    inputErrors = {};
    isKeyAlreadyExists = false;
  }

  function checkForExistingKeys() {
    inputErrors = {};
    isKeyAlreadyExists = false;
    let isDuplicateWithinForm = false;
    let isDuplicateWithExisting = false;

    const formKeys = $form.variables
      .filter((variable) => variable.key.trim() !== "")
      .map((variable) => variable.key);

    const formDuplicates = new Set();
    if (new Set(formKeys).size !== formKeys.length) {
      formKeys.forEach((key, index) => {
        if (formKeys.indexOf(key) !== index) {
          formDuplicates.add(formKeys.indexOf(key));
          formDuplicates.add(index);
          isDuplicateWithinForm = true;
        }
      });
    }

    const existingDuplicates = new Set();
    const existingKeys = existingVariables.map((v) => v.key);

    $form.variables.forEach((variable, index) => {
      if (variable.key.trim() !== "") {
        if (isDuplicateKey(variable.key, existingKeys)) {
          existingDuplicates.add(index);
          isDuplicateWithExisting = true;
        }
      }
    });

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
    const lines = contents.split("\n");
    const variables: EnvVariable[] = [];

    for (const line of lines) {
      const trimmed = line.trim();
      if (trimmed && !trimmed.startsWith("#")) {
        const [key, ...valueParts] = trimmed.split("=");
        if (key) {
          variables.push({
            key: key.trim(),
            value: valueParts
              .join("=")
              .trim()
              .replace(/^["']|["']$/g, ""),
          });
        }
      }
    }

    if (variables.length > 0) {
      const filteredVariables = $form.variables.filter(
        (variable) =>
          variable.key.trim() !== "" || variable.value.trim() !== "",
      );

      $form.variables = [...filteredVariables, ...variables];
    }
  }

  function getKeyFromError(error: { path: string; messages: string[] }) {
    return error.path.split("[")[1].split("]")[0];
  }

  $: isSubmitDisabled =
    $submitting ||
    hasExistingKeys ||
    !hasNewChanges ||
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
    <form
      id={formId}
      class="w-full"
      on:submit|preventDefault={submit}
      use:enhance
    >
      <div class="flex flex-col gap-y-5">
        <div class="flex flex-col gap-y-1">
          <p class="text-xs text-fg-muted">
            Press ⌘⇧. to show .env in file picker
          </p>
          <Button
            type="secondary"
            small
            class="w-fit flex flex-row items-center gap-x-2"
            onClick={() => fileInput.click()}
          >
            <UploadIcon size="14px" />
            <span>Import .env</span>
          </Button>
        </div>
        <input
          type="file"
          bind:this={fileInput}
          on:change={handleFileUpload}
          class="hidden"
        />
        <div class="flex flex-col items-start gap-1">
          <div class="text-sm font-medium text-fg-primary">Variables</div>
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
                    if ($form.variables.length === 1) {
                      handleReset();
                    } else {
                      handleRemove(index);
                    }
                  }}
                >
                  <Trash2Icon size="16px" class="text-fg-muted" />
                </IconButton>
              </div>
            {/each}
          </div>
          <Button type="dashed" class="w-full mt-4" onClick={handleAdd}>
            <Plus size="14px" />
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
                      ? "These keys already exist"
                      : "This key already exists"}
                  {:else}
                    Some keys are duplicated or already exist
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
