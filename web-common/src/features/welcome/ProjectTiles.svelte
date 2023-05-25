<script lang="ts">
  import Subheading from "@rilldata/web-common/components/typography/Subheading.svelte";
  import Card from "../../components/card/Card.svelte";
  import CardDescription from "../../components/card/CardDescription.svelte";
  import CardHeader from "../../components/card/CardHeader.svelte";
  import EmptyProject from "./EmptyProject.svelte";

  const examples = [1, 2, 3];
</script>

<section class="flex flex-col items-center gap-y-5 mt-4 mb-6">
  <Subheading>Or jump right into a project.</Subheading>
  <div class="flex flex-row gap-4 flex-wrap">
    <EmptyProject />
    {#each examples as example}
      <Card>
        <div
          class="flex items-center justify-center w-full h-full m-auto bg-gray-200"
        >
          Image placeholder
        </div>
        <CardHeader>Example {example}</CardHeader>
        <CardDescription>Description</CardDescription>
      </Card>
    {/each}
  </div>
</section>
