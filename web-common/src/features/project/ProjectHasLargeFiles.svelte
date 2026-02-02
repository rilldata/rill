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
  import type { V1DirEntry } from "@rilldata/web-common/runtime-client";
  import Rocket from "svelte-radix/Rocket.svelte";
  import { splitFolderAndFileName } from "@rilldata/web-common/features/entity-management/file-path-utils.ts";

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
  <AlertDialogContent noCancel>
    <AlertDialogHeader>
      <AlertDialogTitle>Unable to deploy to Rill Cloud</AlertDialogTitle>
      <AlertDialogDescription>
        <div class="flex flex-col gap-y-2">
          <div>
            Local files over 100 MB can’t be deployed to Rill Cloud. Please
            upload to S3 or another external storage first.
          </div>
          <ul class="flex flex-col list-disc ml-5 gap-y-1">
            {#each files as file (file.path)}
              {@const formattedSize = formatMemorySize(
                file.size ? Number(file.size) : 0,
              )}
              {@const [, fileName] = splitFolderAndFileName(file.path)}
              <li>
                {fileName} ({formattedSize})
              </li>
            {/each}
          </ul>
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button onClick={() => (open = false)} type="primary" large>Ok</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
