<script lang="ts">
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog/index.js";
  import { Button } from "@rilldata/web-common/components/button/index.js";
  import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size.ts";
  import { AlertCircleIcon } from "lucide-svelte";
  import type { V1DirEntry } from "@rilldata/web-common/runtime-client";
  import Rocket from "svelte-radix/Rocket.svelte";

  export let open = false;
  export let files: V1DirEntry[];
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild let:builder>
    <Button type="primary" builders={[builder]}>
      <Rocket size="16px" />

      Deploy
    </Button>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>
        <div class="flex flex-row items-center gap-x-2">
          <AlertCircleIcon size={16} class="text-red-600" />
          <span>Large files detected</span>
        </div>
      </AlertDialogTitle>
      <AlertDialogDescription>
        <div class="flex flex-col gap-y-2">
          <div>This project has files too large to upload (100MB).</div>
          <div class="flex flex-col gap-y-1">
            {#each files as file (file.path)}
              <div>
                {file.path} ({formatMemorySize(
                  file.size ? Number(file.size) : 0,
                )})
              </div>
            {/each}
            <div>/path/another/file (200MB)</div>
          </div>
          <div>Please upload the file to s3 or similar providers.</div>
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter class="mt-5">
      <Button onClick={() => (open = false)} type="primary">Ok</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
