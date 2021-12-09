<script>
export let data;
$: header = Object.keys(data[0]);
$: intermediate = data.reduce((columns, row) => {
    header.forEach((key) => {
        if (!(key in columns)) { columns[key] = []; }
        columns[key].push(row[key]);
    })
    return columns;
}, {})
</script>

<div class='table-container'>
    <table>
        {#each header as key}
            <tr>
                <th>{key}</th>
                {#each intermediate[key] as value}
                    <td>{`${value}`.slice(0,20)}</td>
                {/each}
            </tr>
        {/each}
    </table>
</div>

<style>
th {
    position: sticky;
    left: 0px;
    background-color: white;
}
</style>