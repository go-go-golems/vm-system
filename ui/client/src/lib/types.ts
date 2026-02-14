// ---------------------------------------------------------------------------
// Domain types (camelCase, used throughout the UI)
// ---------------------------------------------------------------------------

export interface VMProfile {
  id: string;
  name: string;
  engine: string;
  isActive: boolean;
  createdAt: string; // ISO string — avoids Date serialization issues in Redux
  exposedModules: string[];
  libraries: string[];
  settings: {
    limits: {
      cpu_ms: number;
      wall_ms: number;
      mem_mb: number;
      max_events: number;
      max_output_kb: number;
    };
    resolver: {
      roots: string[];
      extensions: string[];
      allow_absolute_repo_imports: boolean;
    };
    runtime: {
      esm: boolean;
      strict: boolean;
      console: boolean;
    };
  };
  capabilities: VMCapability[];
  startupFiles: VMStartupFile[];
}

export interface VMCapability {
  id: string;
  kind: string;
  name: string;
  enabled: boolean;
  config: Record<string, unknown>;
}

export interface VMStartupFile {
  id: string;
  path: string;
  orderIndex: number;
  mode: 'eval' | 'import';
}

export interface VMSession {
  id: string;
  vmId: string;
  vmProfile: string;
  workspaceId: string;
  baseCommitOID: string;
  worktreePath: string;
  status: 'starting' | 'ready' | 'crashed' | 'closed';
  createdAt: string;
  closedAt?: string;
  lastError?: string;
  name: string; // computed client-side from alias or default
}

export interface Execution {
  id: string;
  sessionId: string;
  kind: 'repl' | 'run-file' | 'startup';
  input?: string;
  path?: string;
  status: 'running' | 'ok' | 'error' | 'timeout' | 'cancelled';
  startedAt: string;
  endedAt?: string;
  result?: unknown;
  error?: string;
  events: ExecutionEvent[];
}

export interface ExecutionEvent {
  seq: number;
  ts: string;
  type: 'input_echo' | 'console' | 'value' | 'exception' | 'stdout' | 'stderr' | 'system';
  payload: any;
}

// ---------------------------------------------------------------------------
// Raw API response types (snake_case, as returned by the backend)
// ---------------------------------------------------------------------------

export interface RawTemplate {
  id: string;
  name: string;
  engine: string;
  is_active: boolean;
  exposed_modules: string[];
  libraries: string[];
  created_at: string;
  updated_at: string;
}

export interface RawTemplateDetail {
  template: RawTemplate;
  settings?: {
    limits?: unknown;
    resolver?: unknown;
    runtime?: unknown;
  };
  capabilities: Array<{
    id: string;
    kind: string;
    name: string;
    enabled: boolean;
    config?: unknown;
  }> | null;
  startup_files: Array<{
    id: string;
    path: string;
    order_index: number;
    mode: 'eval' | 'import';
  }> | null;
}

export interface RawSession {
  id: string;
  vm_id: string;
  workspace_id: string;
  base_commit_oid: string;
  worktree_path: string;
  status: 'starting' | 'ready' | 'crashed' | 'closed';
  created_at: string;
  closed_at?: string;
  last_error?: string;
}

export interface RawExecution {
  id: string;
  session_id: string;
  kind: 'repl' | 'run_file' | 'startup';
  input?: string;
  path?: string;
  status: 'running' | 'ok' | 'error' | 'timeout' | 'cancelled';
  started_at: string;
  ended_at?: string;
  result?: unknown;
  error?: unknown;
}

export interface RawExecutionEvent {
  seq: number;
  ts: string;
  type: 'input_echo' | 'console' | 'value' | 'exception' | 'stdout' | 'stderr' | 'system';
  payload: any;
}

// ---------------------------------------------------------------------------
// Mutation arg types
// ---------------------------------------------------------------------------

export interface CreateSessionArgs {
  template_id: string;
  workspace_id: string;
  base_commit_oid: string;
  worktree_path: string;
}

// ---------------------------------------------------------------------------
// Static data (modules, libraries, presets)
// ---------------------------------------------------------------------------

export interface BuiltinModule {
  id: string;
  name: string;
  kind: string;
  description: string;
  functions: string[];
}

export interface BuiltinLibrary {
  id: string;
  name: string;
  version: string;
  description: string;
  source: string;
  type: string;
  config: { global: string };
}

export const BUILTIN_MODULES: BuiltinModule[] = [
  { id: 'database', name: 'database', kind: 'native', description: 'SQLite access via go-go-goja native module', functions: ['configure', 'query', 'exec', 'close'] },
  { id: 'exec', name: 'exec', kind: 'native', description: 'Run external commands through native module bridge', functions: ['run'] },
  { id: 'fs', name: 'fs', kind: 'native', description: 'Read/write files through native module bridge', functions: ['readFileSync', 'writeFileSync'] },
];

export const BUILTIN_LIBRARIES: BuiltinLibrary[] = [
  { id: 'lodash', name: 'Lodash', version: '4.17.21', description: 'A modern JavaScript utility library delivering modularity, performance & extras', source: 'https://cdn.jsdelivr.net/npm/lodash@4.17.21/lodash.min.js', type: 'npm', config: { global: '_' } },
  { id: 'moment', name: 'Moment.js', version: '2.29.4', description: 'Parse, validate, manipulate, and display dates and times in JavaScript', source: 'https://cdn.jsdelivr.net/npm/moment@2.29.4/moment.min.js', type: 'npm', config: { global: 'moment' } },
  { id: 'axios', name: 'Axios', version: '1.6.0', description: 'Promise based HTTP client for the browser and node.js', source: 'https://cdn.jsdelivr.net/npm/axios@1.6.0/dist/axios.min.js', type: 'npm', config: { global: 'axios' } },
  { id: 'ramda', name: 'Ramda', version: '0.29.0', description: 'A practical functional library for JavaScript programmers', source: 'https://cdn.jsdelivr.net/npm/ramda@0.29.0/dist/ramda.min.js', type: 'npm', config: { global: 'R' } },
  { id: 'dayjs', name: 'Day.js', version: '1.11.10', description: 'Fast 2kB alternative to Moment.js with the same modern API', source: 'https://cdn.jsdelivr.net/npm/dayjs@1.11.10/dayjs.min.js', type: 'npm', config: { global: 'dayjs' } },
  { id: 'zustand', name: 'Zustand', version: '4.4.7', description: 'A small, fast and scalable bearbones state-management solution', source: 'https://cdn.jsdelivr.net/npm/zustand@4.4.7/index.js', type: 'npm', config: { global: 'zustand' } },
];

export const PRESET_EXAMPLES = [
  { id: 'hello-world', name: 'Hello World', description: 'Simple console.log example', code: `console.log("Hello from vm-system");\n2 + 2;` },
  { id: 'math-operations', name: 'Math Operations', description: 'Basic arithmetic and math functions', code: `const values = [2, 4, 6, 8];\nconst sum = values.reduce((acc, n) => acc + n, 0);\nconsole.log("sum", sum);\n({ sum, mean: sum / values.length });` },
  { id: 'objects-and-arrays', name: 'Objects and Arrays', description: 'Map/filter/reduce with object output', code: `const users = [{name: "Ana", active: true}, {name: "Bao", active: false}, {name: "Caro", active: true}];\nconst activeNames = users.filter(u => u.active).map(u => u.name);\nconsole.log("active users", activeNames);\n({ count: activeNames.length, activeNames });` },
  { id: 'error-demo', name: 'Error Demo', description: 'Throw and inspect exception output', code: `console.log("about to throw");\nthrow new Error("Intentional test error");` },
  { id: 'library-check', name: 'Library Check', description: 'Validate configured global libraries', code: `const globals = ["_", "moment", "axios", "R", "dayjs", "zustand"];\nconst available = globals.filter((name) => typeof globalThis[name] !== "undefined");\nconsole.log("available globals", available);\navailable;` },
];

// ---------------------------------------------------------------------------
// Environment defaults
// ---------------------------------------------------------------------------

export const API_BASE_URL = (import.meta.env.VITE_VM_SYSTEM_API_BASE_URL || '').trim().replace(/\/$/, '');
export const DEFAULT_WORKSPACE_ID = (import.meta.env.VITE_VM_SYSTEM_WORKSPACE_ID || 'ws-web-ui').trim();
export const DEFAULT_BASE_COMMIT_OID = (import.meta.env.VITE_VM_SYSTEM_BASE_COMMIT_OID || 'web-ui').trim();
export const DEFAULT_WORKTREE_PATH = (import.meta.env.VITE_VM_SYSTEM_WORKTREE_PATH || '/tmp').trim();

export const DEFAULT_TEMPLATE_SPECS = [
  { name: 'Default JavaScript', engine: 'goja', modules: [] },
  { name: 'Utility Playground', engine: 'goja', modules: [], libraries: ['lodash'] },
  { name: 'Library Sandbox', engine: 'goja', modules: [], libraries: ['dayjs', 'ramda'] },
] as const;
