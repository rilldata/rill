<script>
  /**
   * @name {test} this is a test!
   */
</script>

<div class="p-5" style="font-size:14px">
  <pre class="bg-slate-800  text-white p-12 rounded-lg">
<span class="text-red-100">CREATE AND GENERATE AGGREGATING INDEX</span>
	<span class="ui-copy-strong text-yellow-500">idx_store_sales</span> 
<span class="text-red-100">ON</span> (
	<span class="text-slate-500">
	-- <span class="text-blue-400 underline">@name</span> Date Sold
	-- <span class="text-blue-400 underline">@description</span
      > here is the description
	</span>
	<span class="text-yellow-500">ss_sold_date_sk</span>, 

	<span class="text-slate-500">
	-- <span class="text-blue-400 underline">@name</span> Item SKU
	-- <span class="text-blue-400 underline">@description</span
      > the sku of the item, blah
	-- blah blah
	</span>
	<span class="text-yellow-500">ss_item_sk</span>, 

	<span class="text-slate-500"
      >-- a user comment to note that the metrics are below</span
    >

	<span class="text-slate-500">
	-- <span class="text-blue-400 underline">@name</span> Sales Price (USD)
	-- <span class="text-blue-400 underline">@description</span
      > The sales price in USD
	-- blah blah
	</span>
	sum(<span class="text-yellow-500">ss_ext_sales_price</span>), 

	<span class="text-slate-500">
	-- <span class="text-blue-400 underline">@name</span> Discount amount
	-- <span class="text-blue-400 underline">@description</span
      > The discount amount w/ other 
	-- <span class="text-blue-400 underline">@owner</span> Michael Driscoll
	</span>
	<span class="dimension">
	sum(<span class="text-yellow-500">ss_ext_discount_amt</span>)
	</span>
);
</pre>
</div>

<style>
  /* .dimension:after {
	content: '-- Dimension ';
	position: absolute;
	transform: translateX(1.5em);
} */
</style>
