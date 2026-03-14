export type Severity = "breaking" | "non-breaking" | "info";

export interface Change {
  type: string;
  severity: Severity;
  path: string;
  method: string;
  location: string;
  description: string;
  before?: string;
  after?: string;
}

export interface Summary {
  total: number;
  breaking: number;
  non_breaking: number;
  info: number;
}

export interface DiffResult {
  base_file: string;
  head_file: string;
  changes: Change[];
  summary: Summary;
}

export interface Hit {
  file: string;
  line_num: number;
  line: string;
  change_type: string;
  change_path: string;
}

export interface ImpactOptions {
  format?: "text" | "json" | "markdown" | "github";
}

/** Diff two OpenAPI 3.x schemas (YAML or JSON). */
export function compareOpenAPI(base: string, head: string): DiffResult;

/** Diff two GraphQL SDL schemas (.graphql or .gql). */
export function compareGraphQL(base: string, head: string): DiffResult;

/** Diff two Protobuf schemas (.proto). */
export function compareGRPC(base: string, head: string): DiffResult;

/**
 * Scan a source directory for references to breaking changes.
 * Returns Hit[] when format is "json" (default), otherwise a formatted string.
 */
export function impact(diffResult: DiffResult, scanDir?: string, options?: ImpactOptions & { format: "json" }): Hit[];
export function impact(diffResult: DiffResult, scanDir?: string, options?: ImpactOptions & { format: "text" | "markdown" | "github" }): string;
export function impact(diffResult: DiffResult, scanDir?: string, options?: ImpactOptions): Hit[] | string;
