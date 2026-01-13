<script lang="ts">
  import ResourcesSection from "../status/ResourcesSection.svelte";
  import ParseErrorsSection from "../status/ParseErrorsSection.svelte";
</script>

<main
  class="size-full pt-8 pb-16 lg:pt-12 flex flex-col items-center px-8 sm:px-16 lg:px-32 2xl:px-40 overflow-y-auto bg-white dark:bg-gray-950"
>
  <section class="w-full flex flex-col gap-y-3" style:max-width="1100px">
    <h1 class="text-2xl font-semibold" aria-label="Container title">
      Project status
    </h1>

    <div class="flex flex-col gap-y-4 size-full">
      <ResourcesSection />
      <ParseErrorsSection />
    </div>
  </section>
</main>
