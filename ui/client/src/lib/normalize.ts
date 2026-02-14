/**
 * Pure normalization functions extracted from the old VMService class.
 * Used in RTK Query transformResponse callbacks.
 */

import type {
  VMProfile,
  VMSession,
  Execution,
  ExecutionEvent,
  RawTemplateDetail,
  RawSession,
  RawExecution,
  RawExecutionEvent,
} from './types';

// ---------------------------------------------------------------------------
// Low-level helpers
// ---------------------------------------------------------------------------

function toRecord(value: unknown): Record<string, unknown> {
  if (value && typeof value === 'object' && !Array.isArray(value)) {
    return value as Record<string, unknown>;
  }
  return {};
}

function parseNumber(value: unknown, fallback: number): number {
  return typeof value === 'number' && Number.isFinite(value) ? value : fallback;
}

function parseBool(value: unknown, fallback: boolean): boolean {
  return typeof value === 'boolean' ? value : fallback;
}

function parseStringArray(value: unknown, fallback: string[]): string[] {
  if (!Array.isArray(value)) return fallback;
  return value.filter((entry): entry is string => typeof entry === 'string');
}

export function asArray<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : [];
}

// ---------------------------------------------------------------------------
// Settings normalization
// ---------------------------------------------------------------------------

const DEFAULT_LIMITS = { cpu_ms: 2000, wall_ms: 5000, mem_mb: 128, max_events: 50000, max_output_kb: 256 };
const DEFAULT_RESOLVER = { roots: ['.'], extensions: ['.js', '.mjs'], allow_absolute_repo_imports: true };
const DEFAULT_RUNTIME = { esm: true, strict: true, console: true };

function normalizeLimits(raw: unknown): VMProfile['settings']['limits'] {
  const obj = toRecord(raw);
  return {
    cpu_ms: parseNumber(obj.cpu_ms, DEFAULT_LIMITS.cpu_ms),
    wall_ms: parseNumber(obj.wall_ms, DEFAULT_LIMITS.wall_ms),
    mem_mb: parseNumber(obj.mem_mb, DEFAULT_LIMITS.mem_mb),
    max_events: parseNumber(obj.max_events, DEFAULT_LIMITS.max_events),
    max_output_kb: parseNumber(obj.max_output_kb, DEFAULT_LIMITS.max_output_kb),
  };
}

function normalizeResolver(raw: unknown): VMProfile['settings']['resolver'] {
  const obj = toRecord(raw);
  return {
    roots: parseStringArray(obj.roots, DEFAULT_RESOLVER.roots),
    extensions: parseStringArray(obj.extensions, DEFAULT_RESOLVER.extensions),
    allow_absolute_repo_imports: parseBool(obj.allow_absolute_repo_imports, DEFAULT_RESOLVER.allow_absolute_repo_imports),
  };
}

function normalizeRuntime(raw: unknown): VMProfile['settings']['runtime'] {
  const obj = toRecord(raw);
  return {
    esm: parseBool(obj.esm, DEFAULT_RUNTIME.esm),
    strict: parseBool(obj.strict, DEFAULT_RUNTIME.strict),
    console: parseBool(obj.console, DEFAULT_RUNTIME.console),
  };
}

// ---------------------------------------------------------------------------
// Template normalization
// ---------------------------------------------------------------------------

export function mapTemplateDetail(detail: RawTemplateDetail): VMProfile {
  const settings = detail.settings || {};
  return {
    id: detail.template.id,
    name: detail.template.name,
    engine: detail.template.engine,
    isActive: detail.template.is_active,
    createdAt: detail.template.created_at,
    exposedModules: detail.template.exposed_modules || [],
    libraries: detail.template.libraries || [],
    settings: {
      limits: normalizeLimits(settings.limits),
      resolver: normalizeResolver(settings.resolver),
      runtime: normalizeRuntime(settings.runtime),
    },
    capabilities: asArray(detail.capabilities).map(c => ({
      id: c.id,
      kind: c.kind,
      name: c.name,
      enabled: c.enabled,
      config: toRecord(c.config),
    })),
    startupFiles: asArray(detail.startup_files).map(f => ({
      id: f.id,
      path: f.path,
      orderIndex: f.order_index,
      mode: f.mode,
    })),
  };
}

// ---------------------------------------------------------------------------
// Session normalization
// ---------------------------------------------------------------------------

export function mapSession(
  raw: RawSession,
  sessionNames: Record<string, string>,
): VMSession {
  const name = sessionNames[raw.id] || `Session ${raw.id.slice(0, 8)}`;
  return {
    id: raw.id,
    vmId: raw.vm_id,
    vmProfile: '', // filled in by the component using template data
    workspaceId: raw.workspace_id,
    baseCommitOID: raw.base_commit_oid,
    worktreePath: raw.worktree_path,
    status: raw.status,
    createdAt: raw.created_at,
    closedAt: raw.closed_at,
    lastError: raw.last_error,
    name,
  };
}

// ---------------------------------------------------------------------------
// Execution normalization
// ---------------------------------------------------------------------------

function normalizeExecutionResult(rawResult: unknown): unknown {
  const result = toRecord(rawResult);
  if ('json' in result) return result.json;
  return rawResult;
}

function normalizeExecutionError(rawError: unknown): string {
  if (!rawError) return '';
  if (typeof rawError === 'string') return rawError;
  const errorObj = toRecord(rawError);
  if (typeof errorObj.message === 'string') return errorObj.message;
  return JSON.stringify(rawError);
}

function mapExecutionKind(kind: RawExecution['kind']): Execution['kind'] {
  if (kind === 'run_file') return 'run-file';
  return kind;
}

export function mapExecution(raw: RawExecution, events: ExecutionEvent[]): Execution {
  return {
    id: raw.id,
    sessionId: raw.session_id,
    kind: mapExecutionKind(raw.kind),
    input: raw.input,
    path: raw.path,
    status: raw.status,
    startedAt: raw.started_at,
    endedAt: raw.ended_at,
    result: normalizeExecutionResult(raw.result),
    error: normalizeExecutionError(raw.error),
    events,
  };
}

export function mapEvents(rawEvents: RawExecutionEvent[]): ExecutionEvent[] {
  return rawEvents.map(e => ({
    seq: e.seq,
    ts: e.ts,
    type: e.type,
    payload: e.payload,
  }));
}
