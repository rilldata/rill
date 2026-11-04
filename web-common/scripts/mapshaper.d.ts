/** The subset of mapshaper's API used by generate-map-boundaries.ts; mapshaper does not ship types. */
declare module "mapshaper" {
  const mapshaper: {
    /** Runs mapshaper commands against in-memory input files and resolves with the output files by name. */
    applyCommands(
      commands: string,
      input: Record<string, unknown>,
    ): Promise<Record<string, string | Uint8Array>>;
  };
  export default mapshaper;
}
