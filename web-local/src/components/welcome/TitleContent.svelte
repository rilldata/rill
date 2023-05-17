<script lang="ts">
  import CtaButtonShort from "@rilldata/web-common/components/calls-to-action/CTAButtonShort.svelte";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import RillLogoSquareNegative from "@rilldata/web-common/components/icons/RillLogoSquareNegative.svelte";
  import RadixH1 from "@rilldata/web-common/components/typography/RadixH1.svelte";
  import Subheading from "@rilldata/web-common/components/typography/Subheading.svelte";
  import AddSourceModal from "@rilldata/web-common/features/sources/add-source/AddSourceModal.svelte";

  let showAddSourceModal = false;

  const openShowAddSourceModal = () => {
    showAddSourceModal = true;
  };
</script>

<section class="flex flex-col gap-y-6 items-center text-center">
  <RillLogoSquareNegative size="7em" />
  <RadixH1>Welcome to Rill</RadixH1>
  <div class="flex flex-col gap-y-2">
    <Subheading>
      A Rill project lets you build an <span class="text-base font-semibold"
        >un-dashboard</span
      > that people will actually use.
    </Subheading>
    <Subheading>Let’s get started.</Subheading>
  </div>
  <CtaButtonShort on:click={openShowAddSourceModal} variant="primary">
    <Add className="text-white" slot="icon" />
    Add data
  </CtaButtonShort>
  {#if showAddSourceModal}
    <AddSourceModal
      on:close={() => {
        showAddSourceModal = false;
      }}
    />
  {/if}
</section>
