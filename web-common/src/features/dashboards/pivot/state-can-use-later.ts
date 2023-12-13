function preparePivotData(
  data,
  dimensions: string[],
  otherDimensionsAxisValues,
  expanded,
  i = 1
) {
  if (
    dimensions.slice(i).length > 0 &&
    otherDimensionsAxisValues[i - 1]?.length
  ) {
    data.forEach((row) => {
      row.subRows = otherDimensionsAxisValues[i - 1].map((valueObj) => ({
        [dimensions[0]]: valueObj[dimensions[i]],
      }));

      preparePivotData(
        row.subRows,
        dimensions,
        otherDimensionsAxisValues,
        expanded,
        i + 1
      );
    });
  }
}

function addExpandedDataToPivot(data, dimensions, expandedRowMeasureValues) {
  const pivotData = data;

  expandedRowMeasureValues.forEach((expandedRowData) => {
    const indices = expandedRowData.expandIndex
      .split(".")
      .map((index) => parseInt(index, 10));

    let parent = pivotData; // Keep a reference to the parent array
    let lastIdx = 0; // Keep track of the last index

    // Traverse the data array to the right position
    for (let i = 0; i < indices.length; i++) {
      if (!parent[indices[i]]) break;
      if (i < indices.length - 1) {
        parent = parent[indices[i]].subRows;
      }
      lastIdx = indices[i];
    }

    // Update the specific array at the position
    if (parent[lastIdx] && parent[lastIdx].subRows) {
      if (!expandedRowData?.data?.length) {
        parent[lastIdx].subRows = [{ [dimensions[0]]: "" }];
      } else {
        parent[lastIdx].subRows = expandedRowData?.data.map((row) => ({
          ...row,
          [dimensions[0]]: row[dimensions[indices.length]],
        }));
      }
    }
  });
  return pivotData;
}
