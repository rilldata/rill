const svgWidth = 40;
const svgHeight = 13;

export function createSparkline(
  dataArr: Array<unknown>,
  accessor: (v: unknown) => number,
) {
  const noDataSpark = `
  <svg width="${svgWidth}" height="${svgHeight}" xmlns="http://www.w3.org/2000/svg">
      <path d="M0,${svgHeight / 2} L${svgWidth},${
        svgHeight / 2
      }" fill="none" stroke="#9CA3AF" />
  </svg>
`;
  // Check if dataArr is present and has data
  if (!dataArr || dataArr.length === 0) {
    // Return SVG with a flat line in the middle of svgHeight
    return noDataSpark;
  }
  const data = accessor ? dataArr?.map(accessor) : (dataArr as number[]);
  const maxY = Math.max(...data);
  const minY = Math.min(...data);

  const normalizedData = data.map(
    (y) => svgHeight - ((y - minY) / (maxY - minY)) * svgHeight,
  );

  // Normalized data may have NaNs when data is only nulls and 0s
  const hasNaN = normalizedData.every((y) => isNaN(y));
  if (hasNaN) return noDataSpark;

  let d = "";
  normalizedData.forEach((y, i) => {
    const x = (i / (data.length - 1)) * svgWidth;
    if (i === 0) {
      d += `M${x},${y} `;
    } else {
      d += `L${x},${y} `;
    }
  });

  return `
    <svg width="${svgWidth}" height="${svgHeight}" xmlns="http://www.w3.org/2000/svg">
        <path d="${d}" fill="none" stroke="#9CA3AF" />
    </svg>
    `;
}
