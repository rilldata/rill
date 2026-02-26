<script lang="ts">
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
  import { generateSampleData } from "@rilldata/web-common/features/sample-data/generate-sample-data.ts";
  import { SparklesIcon } from "lucide-svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import IconButton from "../../components/button/IconButton.svelte";
  import SendIcon from "@rilldata/web-common/components/icons/SendIcon.svelte";

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
      void generateSampleData(true, instanceId, values.prompt);
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

<div class="container">
  <div class="header">
    <SparklesIcon size="16px" class="rotate-90" />
    <span>Generate sample data</span>
  </div>

  <div>
    <form
      id={FORM_ID}
      on:submit|preventDefault={submit}
      use:enhance
      class="relative"
    >
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
  </div>
</div>

<style lang="postcss">
  .container {
    @apply flex flex-col p-6 gap-4 w-96;
    @apply border border-primary-200 rounded-lg;
    background: radial-gradient(
      127% 274.12% at 111.07% 104.65%,
      #d2f5ec 0%,
      #ddf3ff 100%
    );
  }

  .container:hover {
    @apply border-accent-primary-action shadow-lg;
  }

  .header {
    @apply flex flex-row items-center;
    @apply text-lg text-fg-primary font-semibold;
  }

  .prompt-input {
    @apply w-full my-4 p-2 min-h-28;
    @apply border border-gray-300 rounded-[2px];
    @apply text-sm leading-relaxed;
  }

  .prompt-input.empty::before {
    content: attr(data-placeholder);
    @apply text-fg-secondary pointer-events-none absolute;
  }

  .error {
    @apply text-xs text-red-600 font-normal pb-2;
  }
</style>
