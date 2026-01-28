<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
  import { generateSampleData } from "@rilldata/web-common/features/sample-data/generate-sample-data.ts";
  import { SparklesIcon } from "lucide-svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import IconButton from "../../components/button/IconButton.svelte";
  import SendIcon from "@rilldata/web-common/components/icons/SendIcon.svelte";
  import { builderActions, getAttrs } from "bits-ui";

  export let type: "init" | "home" | "modal";
  export let open = false;

  const initializeProject = type === "init";

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
    {#if type === "init"}
      <Button builders={[builder]} large>
        <SparklesIcon size="14px" class="text-blue-500 rotate-90" />
        <span>Generate sample data</span>
      </Button>
    {:else if type === "home"}
      <Button class="button-home" builders={[builder]} large>
        <SparklesIcon size="14px" class="text-blue-500 rotate-90" />
        <span>Generate sample data</span>
      </Button>
    {:else}
      <div class="hidden"></div>
    {/if}
  </Dialog.Trigger>
  <Dialog.Content>
    <form
      id={FORM_ID}
      on:submit|preventDefault={submit}
      use:enhance
      class="relative"
    >
      <Dialog.Header>
        <Dialog.Title class="flex flex-row items-center gap-x-1 text-blue-500">
          <SparklesIcon size="16px" class="rotate-90" />
          <span>Generate sample data</span>
        </Dialog.Title>
        <Dialog.Description>
          <div>What is the business context or domain of your data?</div>
        </Dialog.Description>
      </Dialog.Header>
      <textarea
        class="prompt-input"
        bind:value={$form.prompt}
        class:empty={$form.prompt.length === 0}
        placeholder={`E.g. "e-commerce transactions"`}
        on:keydown={handleKeydown}
      />
      <div class="absolute right-3 bottom-8">
        <IconButton ariaLabel="Send message" on:click={submit}>
          <SendIcon size="1.3em" />
        </IconButton>
      </div>
      {#if $errors.prompt}
        <div class="error">{$errors.prompt?.[0]}</div>
      {/if}
    </form>
  </Dialog.Content>
</Dialog.Root>

<style lang="postcss">
  .prompt-input {
    @apply w-full my-4 p-2 min-h-28;
    @apply border border-gray-300 rounded-[2px];
    @apply text-sm leading-relaxed;
  }

  .prompt-input.empty::before {
    content: attr(data-placeholder);
    @apply text-fg-secondary pointer-events-none absolute;
  }

  .button-home {
    @apply flex flex-row text-center items-center justify-center gap-x-2 h-12;
    @apply text-sm border rounded-md shadow-sm;
  }

  .error {
    @apply text-xs text-red-600 font-normal pb-2;
  }
</style>
