// https://github.com/rilldata/rill/pull/5929/files#diff-7e33f00ad59643313709bc6c54ef7d24f0b93fc63a48a7f17d86b2795237e93eR8
export enum EnvironmentType {
  UNDEFINED = "",
  DEVELOPMENT = "dev",
  PRODUCTION = "prod",
}

export type EnvironmentTypes = "dev" | "prod" | "";

export type EnvironmentVariable = {
  key: string;
  value: string;
};
