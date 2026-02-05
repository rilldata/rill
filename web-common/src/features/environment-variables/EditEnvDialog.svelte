<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import {
    Dialog,
    DialogContent,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { createEventDispatcher } from "svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import { isDuplicateKey } from "./utils";
  import type { EnvVariable } from "./types";

  export let open = false;
  export let keyName: string;
  export let value: string;
  export let existingVariables: EnvVariable[] = [];

  const dispatch = createEventDispatcher<{
    save: { oldKey: string; key: string; value: string };
  }>();

  let isKeyAlreadyExists = false;
  let initialKey = keyName;
  let initialValue = value;

  const schema = yup(
    object({
      key: string()
        .required("Key is required")
        .matches(
          /^[a-zA-Z_][a-zA-Z0-9_.]*$/,
          "Key must start with a letter or underscore and can only contain letters, digits, underscores, and dots",
        ),
      value: string().required().default(""),
    }),
  );

  const { form, enhance, submit, submitting, errors } = superForm(
    defaults({ key: "", value: "" }, schema),
    {
      SPA: true,
      validators: schema,
      dataType: "json",
      async onUpdate({ form: formResult }) {
        if (!formResult.valid) return;

        checkForExistingKey();
        if (isKeyAlreadyExists) return;

        dispatch("save", {
          oldKey: initialKey,
          key: formResult.data.key,
          value: formResult.data.value,
        });

        open = false;
      },
    },
  );

  $: hasNewChanges = $form.key !== initialKey || $form.value !== initialValue;

  function handleKeyChange(event: Event) {
    const target = event.target as HTMLInputElement;
    $form.key = target.value;
    checkForExistingKey();
  }

  function handleValueChange(event: Event) {
    const target = event.target as HTMLInputElement;
    $form.value = target.value;
  }

  function checkForExistingKey() {
    const existingKeys = existingVariables.map((v) => v.key);
    isKeyAlreadyExists = isDuplicateKey($form.key, existingKeys, initialKey);
  }

  $: if (open) {
    initialKey = keyName;
    initialValue = value;
    $form.key = keyName;
    $form.value = value;
    isKeyAlreadyExists = false;
  }

  $: isSubmitDisabled = $submitting || !hasNewChanges || isKeyAlreadyExists;
</script>

<Dialog bind:open>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent class="translate-y-[-200px]">
    <DialogHeader>
      <DialogTitle>Edit environment variable</DialogTitle>
    </DialogHeader>
    <form class="w-full" on:submit|preventDefault={submit} use:enhance>
      <div class="flex flex-col gap-y-5">
        <div class="flex flex-col items-start gap-1">
          <div class="text-sm font-medium text-fg-primary">Variable</div>
          <div class="flex flex-col w-full gap-2">
            <div class="flex flex-row items-center gap-2">
              <Input
                bind:value={$form.key}
                label=""
                id="edit-key"
                textClass={isKeyAlreadyExists || $errors.key
                  ? "error-input-wrapper"
                  : ""}
                placeholder="Key"
                on:input={handleKeyChange}
              />
              <Input
                bind:value={$form.value}
                label=""
                id="edit-value"
                placeholder="Value"
                on:input={handleValueChange}
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
                  This key already exists
                </p>
              </div>
            {/if}
          </div>
        </div>
      </div>
    </form>

    <DialogFooter>
      <Button
        type="tertiary"
        onClick={() => {
          open = false;
        }}>Cancel</Button
      >
      <Button type="primary" disabled={isSubmitDisabled} onClick={submit}>
        Save
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
