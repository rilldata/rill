<script lang="ts">
  import CTAHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CTAMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import CTANeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import MoonCircleOutline from "@rilldata/web-common/components/icons/MoonCircleOutline.svelte";
</script>

<div class="flex flex-col justify-center items-center gap-y-6 my-20">
  <div class="flex flex-col gap-y-2">
    <MoonCircleOutline size="104px" className="text-slate-300" />
    <CTAHeader variant="bold">This org’s projects are hibernating</CTAHeader>
    <CTAMessage>
      Please reach out to your administrator to regain access.
    </CTAMessage>
    <CTANeedHelp />
  </div>
</div>
