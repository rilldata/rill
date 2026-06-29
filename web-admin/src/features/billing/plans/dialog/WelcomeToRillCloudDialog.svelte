<script lang="ts">
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import Champagne from "@rilldata/web-common/components/icons/Champagne.svelte";
  import { SELF_SERVE_PLANS_BY_NAME, getTranslatedPlanDisplayName } from "@rilldata/web-admin/features/billing/plans/plan-details.ts";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  export let open: boolean;
  export let planName: string;

  $: planDisplayName = getTranslatedPlanDisplayName(planName) ||
    (SELF_SERVE_PLANS_BY_NAME[planName]?.displayName ?? planName);
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </AlertDialogTrigger>
  <AlertDialogContent class="flex flex-row gap-x-2 min-w-[600px]">
    <Champagne size="150px" className="min-w-[150px]" />
    <div class="flex flex-col gap-x-2">
      <AlertDialogHeader>
        <AlertDialogTitle>{m.billing_welcome_to_rill_cloud()}</AlertDialogTitle>
        <AlertDialogDescription>
          {@html m.billing_congrats_plan({
            planName: `<b>${planDisplayName}</b>`,
            docsLink: `<a href="https://docs.rilldata.com/" target="_blank" class="text-primary-600 font-medium">${m.billing_refer_to_docs()}</a>`,
          })}
        </AlertDialogDescription>
      </AlertDialogHeader>
      <div class="grow"></div>
      <AlertDialogFooter class="mt-3">
        <Button type="primary" onClick={() => (open = false)}>{m.billing_got_it()}</Button>
      </AlertDialogFooter>
    </div>
  </AlertDialogContent>
</AlertDialog>
