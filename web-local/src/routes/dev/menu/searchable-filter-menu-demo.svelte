<script lang="ts">
  import SeachableFilterButton from "@rilldata/web-common/components/searchable-filter-menu/SeachableFilterButton.svelte";

  let selectableItems = [
    "sales",
    "sales",
    "sales",
    "2returns",
    "3new customers",
    "4asdfa",
    "5asdfwet",
    "6qwrtqwer",
    "7salesol;",
    "8returns,n;o;",
    "9new customers6l,",
    "10vbafdg",
    "11wetafasdfwet",
    "12qwrbafbatqwer",
    "13sales",
    "14returagfbvans",
    "15new cwebvaaustomers",
    "16aaeasdfa",
    "17asdfwzxcvawseet",
    "18qwasdfasdrtqwer",
    "19sasdfgalesol;",
    "20retuasdfrns,n;o;",
    "21asdfnew customers6l,",
    "22vbafdfasddg",
    "23wetafasdfasdfwet",
    "24qwrbafbasdfatqwer",
  ];

  let selectedItems: boolean[] = selectableItems.map((_, i) => i % 2 == 0);

  $: {
    console.log(
      "selected items",
      selectableItems.filter((_, i) => selectedItems[i])
    );
  }
</script>

<section class="grid grid-flow-row gap-y-4" style="margin-top:40px">
  <h1 class="text-lg">Searchable Filter Menus</h1>
  <SeachableFilterButton
    label="Measures"
    tooltipText="Choose measures to display"
    {selectableItems}
    {selectedItems}
    on:itemClicked={(ev) => {
      console.log("selectedItemsChanged", ev);
      const index = ev.detail;
      selectedItems = selectedItems.map((x, i) => (i === index ? !x : x));
      // selectedItems = ev.detail;
    }}
    on:deselectAll={() => {
      selectedItems = selectedItems.map((_) => false);
    }}
  />
  {#each selectableItems.filter((_, i) => selectedItems[i]) as item}
    {item}&nbsp;
  {/each}
</section>
