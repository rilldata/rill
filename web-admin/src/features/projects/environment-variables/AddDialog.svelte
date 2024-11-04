<script lang="ts">
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
  import ErrorMessage from "./ErrorMessage.svelte";
  import KeyValueItem from "./KeyValueItem.svelte";

  export let open = false;

  let errorMessage = "";
  let isDevelopment = false;
  let isProduction = false;
  let variables: { key: string; value: string }[] = [];

  $: console.log(variables);

  function handleDelete(index: number) {
    variables = variables.filter((_, i) => i !== index);
  }

  // TODO: wire up `createAdminServiceUpdateProjectVariables`
</script>

<Dialog bind:open>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent class="translate-y-[-200px]">
    <DialogHeader>
      <DialogTitle>Add environment variables</DialogTitle>
    </DialogHeader>
    <DialogDescription>
      For help, see <a
        href="https://docs.rilldata.com/tutorials/administration/project/credential-envvariable-mangement"
        target="_blank">documentation</a
      >
    </DialogDescription>
    <div class="flex flex-col gap-y-5">
      <Button type="secondary" small class="w-fit">
        <!-- TODO: onclick to trigger file upload, parse the content -->
        <span>Import .env file</span>
      </Button>
      <div class="flex flex-col items-start gap-1">
        <div class="text-sm font-medium text-gray-800">Environment</div>
        <div class="flex flex-row gap-4 mt-1">
          <!-- TODO: check the usage before changing the label color to text-gray-800 -->
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
          {#each variables as variable, index}
            <KeyValueItem
              {variable}
              {index}
              on:delete={() => handleDelete(index)}
            />
          {/each}
        </div>
        <Button
          type="dashed"
          class="w-full mt-4"
          on:click={() => {
            variables = [...variables, { key: "", value: "" }];
          }}
        >
          <Plus size="16px" />
          <span>Add variable</span>
        </Button>
        {#if errorMessage}
          <div class="mt-1">
            <ErrorMessage />
          </div>
        {/if}
      </div>
    </div>

    <DialogFooter>
      <Button type="plain">Cancel</Button>
      <Button type="primary">Create</Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
