export function stripDefaultUrlParams(
  searchParams: URLSearchParams,
  defaultUrlParams: URLSearchParams,
) {
  const strippedUrlParams = new URLSearchParams();
  searchParams.forEach((value, key) => {
    const defaultValue = defaultUrlParams.get(key);
    if (defaultValue !== null && value === defaultValue) {
      return;
    }
    strippedUrlParams.set(key, value);
  });
  return strippedUrlParams;
}

export function mergeDefaultUrlParams(
  searchParams: URLSearchParams,
  defaultUrlParams: URLSearchParams,
) {
  const finalUrlParams = new URLSearchParams(searchParams);
  defaultUrlParams.forEach((value, key) => {
    if (searchParams.has(key)) return;
    finalUrlParams.set(key, value);
  });
  return finalUrlParams;
}
