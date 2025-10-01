<script lang="ts">
  import SubmissionError from "@rilldata/web-common/components/forms/SubmissionError.svelte";
  import CopyIcon from "@rilldata/web-common/components/icons/CopyIcon.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import NeedHelpText from "./NeedHelpText.svelte";
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";

  export let connector: V1ConnectorDriver;
  export let isSourceForm: boolean;
  export let dsnError: string | null;
  export let paramsError: string | null;
  export let clickhouseError: string | null;
  export let dsnErrorDetails: string | undefined;
  export let paramsErrorDetails: string | undefined;
  export let clickhouseErrorDetails: string | undefined;
  export let hasOnlyDsn: boolean;
  export let connectionTab: string;
  export let yamlPreview: string;
  export let copied: boolean;
  export let copyYamlPreview: () => void;
</script>

<div
  class="add-data-side-panel flex flex-col gap-6 p-6 bg-[#FAFAFA] w-full max-w-full border-l-0 border-t mt-6 pl-0 pt-6 md:w-96 md:min-w-[320px] md:max-w-[400px] md:border-l md:border-t-0 md:mt-0 md:pl-6"
>
  {#if dsnError || paramsError || clickhouseError}
    <SubmissionError
      message={clickhouseError ??
        (hasOnlyDsn || connectionTab === "dsn" ? dsnError : paramsError) ??
        ""}
      details={clickhouseErrorDetails ??
        (hasOnlyDsn || connectionTab === "dsn"
          ? dsnErrorDetails
          : paramsErrorDetails) ??
        ""}
    />
  {/if}

  <div>
    <div class="text-sm leading-none font-medium mb-4">
      {isSourceForm ? "Model preview" : "Connector preview"}
    </div>
    <div class="relative">
      <button
        class="absolute top-2 right-2 p-1 rounded"
        type="button"
        aria-label="Copy YAML"
        on:click={copyYamlPreview}
      >
        {#if copied}
          <Check size="16px" />
        {:else}
          <CopyIcon size="16px" />
        {/if}
      </button>
      <pre
        class="bg-muted p-3 rounded text-xs border border-gray-200 overflow-x-auto">{yamlPreview}</pre>
    </div>
  </div>

  <NeedHelpText {connector} />
</div>
