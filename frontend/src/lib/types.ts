export type User = {
  id: string;
  name: string;
  email: string;
  created_at: string;
};

export type Workspace = {
  id: string;
  user_id: string;
  name: string;
  db_type: string;
  has_database: boolean;
  created_at: string;
  updated_at: string;
};

export type DatabaseSchema = {
  tables: Array<{
    name: string;
    row_count: number;
    columns: Array<{ name: string; type: string; primary_key: boolean; nullable: boolean }>;
    foreign_keys: Array<{ column: string; ref_table: string; ref_column: string }>;
  }>;
};

export type QueryResult = {
  columns: string[];
  rows: unknown[][];
  row_count: number;
  execution_ms: number;
};

export type HistoryItem = {
  id: string;
  question?: string;
  generated_sql?: string;
  executed_sql?: string;
  explanation?: string;
  status: string;
  error_message?: string;
  execution_ms?: number;
  created_at: string;
};

export type AIHealth = {
  status: string;
  provider: string;
  model: string;
  base_url: string;
};
