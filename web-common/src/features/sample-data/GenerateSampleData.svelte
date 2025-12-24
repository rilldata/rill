<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
  import { generateSampleData } from "@rilldata/web-common/features/sample-data/generate-sample-data.ts";
  import { SparklesIcon } from "lucide-svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";

  export let initializeProject: boolean;
  export let open = false;

  $: ({ instanceId } = $runtime);

  const FORM_ID = "generate-sample-data-form";

  const schema = yup(
    object({
      prompt: string()
        .required("Please describe your data")
        .min(10, "Please provide more detail (at least 10 characters)"),
    }),
  );
  const initialValues = { prompt: "" };
  const superFormInstance = superForm(defaults(initialValues, schema), {
    SPA: true,
    validators: schema,
    dataType: "json",
    async onUpdate({ form }) {
      if (!form.valid) return;
      const values = form.data;
      void generateSampleData(initializeProject, instanceId, values.prompt);
      open = false;
    },
    invalidateAll: false,
  });
  $: ({ form, errors, enhance, submit } = superFormInstance);

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === "Enter" && !event.shiftKey) {
      event.preventDefault();
      submit();
    }
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Trigger asChild let:builder>
    {#if initializeProject}
      <Button builders={[builder]} large>
        <SparklesIcon size="14px" />
        <span>Generate sample data</span>
      </Button>
    {:else}
      <div class="hidden"></div>
    {/if}
  </Dialog.Trigger>
  <Dialog.Content noClose>
    <form id={FORM_ID} on:submit|preventDefault={submit} use:enhance>
      <Dialog.Header>
        <Dialog.Title>Generate sample data</Dialog.Title>
        <Dialog.Description>
          <div>What is the business context or domain of your data?</div>
        </Dialog.Description>
      </Dialog.Header>
      <textarea
        class="prompt-input"
        bind:value={$form.prompt}
        class:empty={$form.prompt.length === 0}
        placeholder="e.g. Sales transaction of an e-commerce store"
        on:keydown={handleKeydown}
      />
      {#if $errors.prompt}
        <div class="error">{$errors.prompt?.[0]}</div>
      {/if}

      <Dialog.Footer>
        <Button type="secondary" large onClick={() => (open = false)}>
          Cancel
        </Button>
        <Button type="primary" large form={FORM_ID} onClick={submit}>
          Generate
        </Button>
      </Dialog.Footer>
    </form>
  </Dialog.Content>
</Dialog.Root>

<style lang="postcss">
  .prompt-input {
    @apply w-full my-4 p-2 min-h-[2.5rem];
    @apply border border-gray-300 rounded-[2px];
    @apply text-sm leading-relaxed;
  }

  .prompt-input.empty::before {
    content: attr(data-placeholder);
    @apply text-gray-400 pointer-events-none absolute;
  }

  .error {
    @apply text-xs text-red-600 font-normal pb-2;
  }
</style>
