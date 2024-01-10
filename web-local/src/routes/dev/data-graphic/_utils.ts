function smooth(data, accessor, windowSize = 7) {
  return data.map((datum, i) => {
    const window = data.slice(Math.max(0, i - windowSize), i);
    const v = window.reduce((acc, v) => acc + v[accessor], 0);
    return { ...datum, [accessor]: v };
  });
}

export function makeTimeSeries(length = 180, smoothingWindow = 7) {
  let value = 100;
  const data = Array.from({ length }).map((_, i) => {
    value += (Math.random() - 0.5) * 30;
    if (value < 0) value = 1;
    return {
      period: new Date(
        +new Date("2010-01-01 00:01:04") + i * 1000 * 60 * 60 * 24,
      ),
      value,
    };
  });
  return smooth(data, smoothingWindow);
}
