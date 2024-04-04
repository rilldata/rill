<script lang="ts">
  export let rowCount: string;
  export let columnCount: string;
</script>

<div class="size-full px-4">
  <table class="size-full text-gray-800">
    <tbody>
      <tr>
        <td>
          <slot name="row-header" />
        </td>
        <td class="text-right">
          <b>{rowCount}</b>
        </td>
      </tr>
      <tr>
        <td>
          <slot name="column-header" />
        </td>
        <td class="text-right">
          <b>{columnCount}</b>
        </td>
      </tr>
    </tbody>
  </table>
</div>
