export type Param = string;

export type PathOption = {
  label: string;
  depth?: number;
  href?: string;
  section?: string;
  pill?: string;
};

export type PathOptions = Map<Param, PathOption>;
