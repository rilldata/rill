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

  export let open = false;

  let isDevelopment = false;
  let isProduction = false;

  // TODO: wire up `UpsertProjectVariable`
</script>

<Dialog bind:open>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent class="translate-y-[-200px]">
    <DialogHeader>
      <DialogTitle>Add environment variables</DialogTitle>
      <DialogDescription class="pt-2">
        For help, see <a
          href="https://docs.rilldata.com/tutorials/administration/project/credential-envvariable-mangement"
          target="_blank">documentation</a
        >
      </DialogDescription>
      <div class="flex flex-col gap-y-5">
        <!-- TODO: Import .env file -->
        <!-- File upload component -->
        <div class="flex flex-col items-start pt-5 gap-1">
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
        <div class="flex flex-col items-start pt-5 gap-1">
          <div class="text-sm font-medium text-gray-800">Variables</div>
          <!-- TODO: KV fields -->
          <!-- TODO: Add variable button -->
          <!-- TODO: error message -->
        </div>
      </div>
    </DialogHeader>
    <DialogFooter>
      <Button type="plain">Cancel</Button>
      <Button type="primary">Create</Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
