<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
  import Rocket from "svelte-radix/Rocket.svelte";
</script>

<svelte:head>
  <title>Reports - Rill</title>
</svelte:head>

<main
  class="bg-surface-base size-full pt-8 pb-16 lg:pt-12 flex flex-col items-center px-8 sm:px-16 lg:px-32 2xl:px-40 overflow-y-auto"
  style="scrollbar-gutter: stable;"
>
  <section class="w-full flex flex-col gap-y-3" style:max-width="800px">
    <h1
      class="text-2xl text-fg-primary font-semibold"
      aria-label="Container title"
    >
      Reports
    </h1>

    <div class="flex flex-col items-center gap-y-4 pt-16">
      <ReportIcon size="32px" />
      <div class="flex flex-col gap-y-2 items-center text-sm">
        <div class="text-fg-secondary font-semibold">
          You don't have any reports yet
        </div>
        <div class="text-fg-muted">
          Schedule reports from any dashboard after deploying
        </div>
      </div>
      <div class="mt-4">
        <Button type="primary" href="/deploy">
          <Rocket size="16px" />
          Deploy to Rill Cloud
        </Button>
      </div>
    </div>
  </section>
</main>
