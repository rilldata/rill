<script lang="ts">
  import CreateShareableURLForm from "@rilldata/web-admin/features/shareable-urls/CreateShareableURLForm.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Link from "@rilldata/web-common/components/icons/Link.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import {
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
  } from "@rilldata/web-common/components/tabs";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
</script>

<Popover>
  <PopoverTrigger asChild let:builder>
    <Button type="secondary" builders={[builder]}>Share</Button>
  </PopoverTrigger>
  <PopoverContent align="end" class="w-[402px] p-0">
    <Tabs>
      <TabsList>
        <TabsTrigger value="tab1">Copy link</TabsTrigger>
        <TabsTrigger value="tab2">Create public link</TabsTrigger>
      </TabsList>
      <TabsContent value="tab1" class="mt-0 p-4">
        <div class="flex flex-col gap-y-2">
          <h3>Share your current view with a fellow member of the project.</h3>
          <Button
            type="secondary"
            on:click={() => {
              copyToClipboard(window.location.href, "Link copied to clipboard");
            }}
          >
            <Link size="16px" className="text-primary-500" />
            Copy link
          </Button>
        </div>
      </TabsContent>
      <TabsContent value="tab2" class="mt-0 p-4">
        <CreateShareableURLForm />
      </TabsContent>
    </Tabs>
  </PopoverContent>
</Popover>
