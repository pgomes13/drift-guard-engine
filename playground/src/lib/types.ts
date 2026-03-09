export type SchemaType = "openapi" | "graphql" | "grpc";

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

export interface CompareRequest {
  schema_type: SchemaType;
  base_content: string;
  head_content: string;
}
