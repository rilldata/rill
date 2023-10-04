const svgWidth = 40;
const svgHeight = 13;

export function createSparkline(dataArr, accessor) {
  const data = accessor ? dataArr?.map(accessor) : dataArr;
  const maxY = Math.max(...data);
  const minY = Math.min(...data);

  const normalizedData = data.map(
    (y) => svgHeight - ((y - minY) / (maxY - minY)) * svgHeight
  );

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
