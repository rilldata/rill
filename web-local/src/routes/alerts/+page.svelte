<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import AlertIcon from "@rilldata/web-common/components/icons/AlertIcon.svelte";
  import Rocket from "svelte-radix/Rocket.svelte";
</script>

<svelte:head>
  <title>Alerts - Rill</title>
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
      Alerts
    </h1>

    <div class="flex flex-col items-center gap-y-4 pt-16">
      <AlertIcon size="32px" />
      <div class="flex flex-col gap-y-2 items-center text-sm">
        <div class="text-fg-secondary font-semibold">
          You don't have any alerts yet
        </div>
        <div class="text-fg-muted">
          Create alerts from any dashboard or via code after deploying
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
