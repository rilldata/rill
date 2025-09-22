<script lang="ts">
  import { page } from "$app/stores";
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
  import DeployIcon from "@rilldata/web-common/components/icons/DeployIcon.svelte";
  import { getDeployRoute } from "@rilldata/web-common/features/project/deploy/route-utils.ts";

  export let open: boolean;

  $: deployCTAUrl = getDeployRoute($page);
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent class="min-w-[600px]">
    <div class="flex flex-row">
      <div class="w-[150px]">
        <DeployIcon size="150px" />
      </div>
      <div class="flex flex-col">
        <AlertDialogHeader>
          <AlertDialogTitle>Deploy this project for free</AlertDialogTitle>
          <AlertDialogDescription>
            You’re about to start a
            <a
              href="https://www.rilldata.com/pricing"
              target="_blank"
              class="text-primary-600"
            >
              30-day FREE trial
            </a>
            of Rill Cloud, where you can set alerts, share dashboards, and more.
            The trial grants you 1 project up to 10GB.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter class="mt-5">
          <Button onClick={() => (open = false)} type="secondary">Back</Button>
          <Button
            onClick={() => (open = false)}
            type="primary"
            href={$deployCTAUrl}
            target="_blank"
          >
            Continue
          </Button>
        </AlertDialogFooter>
      </div>
    </div>
  </AlertDialogContent>
</AlertDialog>
