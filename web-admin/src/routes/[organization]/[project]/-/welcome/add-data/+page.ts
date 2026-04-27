export function load({ url: { searchParams } }) {
  return {
    schema: searchParams.get("schema") ?? undefined,
  };
}
