export interface GithubEventFields {
  is_fresh_connection?: boolean;
  is_overwrite?: boolean;
  has_subpath?: boolean;
  has_non_default_branch?: boolean;
  failure_error?: string;
}
