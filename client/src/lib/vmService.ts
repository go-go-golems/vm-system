export interface VMProfile {
  id: string;
  name: string;
  engine: string;
  isActive: boolean;
  createdAt: Date;
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
  createdAt: Date;
  closedAt?: Date;
  lastActivityAt: Date;
  lastError?: string;
  name: string;
  vm?: VMProfile;
}

export interface Execution {
  id: string;
  sessionId: string;
  kind: 'repl' | 'run-file' | 'startup';
  input?: string;
  path?: string;
  status: 'running' | 'ok' | 'error' | 'timeout' | 'cancelled';
  startedAt: Date;
  endedAt?: Date;
  result?: unknown;
  error?: string;
  events: ExecutionEvent[];
}

export interface ExecutionEvent {
  seq: number;
  ts: Date;
  type: 'input_echo' | 'console' | 'value' | 'exception' | 'stdout' | 'stderr' | 'system';
  payload: any;
}

interface BuiltinModule {
  id: string;
  name: string;
  kind: string;
  description: string;
  functions: string[];
}

interface BuiltinLibrary {
  id: string;
  name: string;
  version: string;
  description: string;
  source: string;
  type: string;
  config: {
    global: string;
  };
}

export const BUILTIN_MODULES: BuiltinModule[] = [
  {
    id: 'console',
    name: 'console',
    kind: 'builtin',
    description: 'Console logging and debugging',
    functions: ['log', 'warn', 'error', 'info', 'debug'],
  },
  {
    id: 'math',
    name: 'Math',
    kind: 'builtin',
    description: 'Mathematical functions and constants',
    functions: ['abs', 'ceil', 'floor', 'round', 'sqrt', 'pow', 'random'],
  },
  {
    id: 'json',
    name: 'JSON',
    kind: 'builtin',
    description: 'JSON parsing and stringification',
    functions: ['parse', 'stringify'],
  },
  {
    id: 'date',
    name: 'Date',
    kind: 'builtin',
    description: 'Date and time manipulation',
    functions: ['now', 'parse', 'UTC'],
  },
  {
    id: 'array',
    name: 'Array',
    kind: 'builtin',
    description: 'Array manipulation methods',
    functions: ['map', 'filter', 'reduce', 'forEach', 'find', 'some', 'every'],
  },
  {
    id: 'string',
    name: 'String',
    kind: 'builtin',
    description: 'String manipulation methods',
    functions: ['split', 'join', 'slice', 'substring', 'indexOf', 'replace'],
  },
  {
    id: 'object',
    name: 'Object',
    kind: 'builtin',
    description: 'Object manipulation methods',
    functions: ['keys', 'values', 'entries', 'assign', 'freeze'],
  },
  {
    id: 'promise',
    name: 'Promise',
    kind: 'builtin',
    description: 'Asynchronous programming with promises',
    functions: ['resolve', 'reject', 'all', 'race'],
  },
];

export const BUILTIN_LIBRARIES: BuiltinLibrary[] = [
  {
    id: 'lodash',
    name: 'Lodash',
    version: '4.17.21',
    description: 'A modern JavaScript utility library delivering modularity, performance & extras',
    source: 'https://cdn.jsdelivr.net/npm/lodash@4.17.21/lodash.min.js',
    type: 'npm',
    config: { global: '_' },
  },
  {
    id: 'moment',
    name: 'Moment.js',
    version: '2.29.4',
    description: 'Parse, validate, manipulate, and display dates and times in JavaScript',
    source: 'https://cdn.jsdelivr.net/npm/moment@2.29.4/moment.min.js',
    type: 'npm',
    config: { global: 'moment' },
  },
  {
    id: 'axios',
    name: 'Axios',
    version: '1.6.0',
    description: 'Promise based HTTP client for the browser and node.js',
    source: 'https://cdn.jsdelivr.net/npm/axios@1.6.0/dist/axios.min.js',
    type: 'npm',
    config: { global: 'axios' },
  },
  {
    id: 'ramda',
    name: 'Ramda',
    version: '0.29.0',
    description: 'A practical functional library for JavaScript programmers',
    source: 'https://cdn.jsdelivr.net/npm/ramda@0.29.0/dist/ramda.min.js',
    type: 'npm',
    config: { global: 'R' },
  },
  {
    id: 'dayjs',
    name: 'Day.js',
    version: '1.11.10',
    description: 'Fast 2kB alternative to Moment.js with the same modern API',
    source: 'https://cdn.jsdelivr.net/npm/dayjs@1.11.10/dayjs.min.js',
    type: 'npm',
    config: { global: 'dayjs' },
  },
  {
    id: 'zustand',
    name: 'Zustand',
    version: '4.4.7',
    description: 'A small, fast and scalable bearbones state-management solution',
    source: 'https://cdn.jsdelivr.net/npm/zustand@4.4.7/index.js',
    type: 'npm',
    config: { global: 'zustand' },
  },
];

export const PRESET_EXAMPLES = [
  {
    id: 'hello-world',
    name: 'Hello World',
    description: 'Simple console.log example',
    code: `console.log("Hello from vm-system");\n2 + 2;`,
  },
  {
    id: 'math-operations',
    name: 'Math Operations',
    description: 'Basic arithmetic and math functions',
    code: `const values = [2, 4, 6, 8];\nconst sum = values.reduce((acc, n) => acc + n, 0);\nconsole.log("sum", sum);\n({ sum, mean: sum / values.length });`,
  },
  {
    id: 'objects-and-arrays',
    name: 'Objects and Arrays',
    description: 'Map/filter/reduce with object output',
    code: `const users = [{name: "Ana", active: true}, {name: "Bao", active: false}, {name: "Caro", active: true}];\nconst activeNames = users.filter(u => u.active).map(u => u.name);\nconsole.log("active users", activeNames);\n({ count: activeNames.length, activeNames });`,
  },
  {
    id: 'error-demo',
    name: 'Error Demo',
    description: 'Throw and inspect exception output',
    code: `console.log("about to throw");\nthrow new Error("Intentional test error");`,
  },
  {
    id: 'library-check',
    name: 'Library Check',
    description: 'Validate configured global libraries',
    code: `const globals = ["_", "moment", "axios", "R", "dayjs", "zustand"];\nconst available = globals.filter((name) => typeof globalThis[name] !== "undefined");\nconsole.log("available globals", available);\navailable;`,
  },
];

interface APIErrorEnvelope {
  error?: {
    code?: string;
    message?: string;
    details?: unknown;
  };
}

class APIError extends Error {
  status: number;
  code?: string;
  details?: unknown;

  constructor(status: number, message: string, code?: string, details?: unknown) {
    super(message);
    this.name = 'APIError';
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

interface RawTemplate {
  id: string;
  name: string;
  engine: string;
  is_active: boolean;
  exposed_modules: string[];
  libraries: string[];
  created_at: string;
  updated_at: string;
}

interface RawTemplateDetail {
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

interface RawSession {
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

interface RawExecution {
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

interface RawExecutionEvent {
  seq: number;
  ts: string;
  type: 'input_echo' | 'console' | 'value' | 'exception' | 'stdout' | 'stderr' | 'system';
  payload: any;
}

interface TemplateBootstrapSpec {
  name: string;
  engine: string;
  modules?: string[];
  libraries?: string[];
}

const SESSION_NAME_STORAGE_KEY = 'vm-system-ui.session-names';
const CURRENT_SESSION_STORAGE_KEY = 'vm-system-ui.current-session-id';

const API_BASE_URL = (import.meta.env.VITE_VM_SYSTEM_API_BASE_URL || '').trim().replace(/\/$/, '');
const DEFAULT_WORKSPACE_ID = (import.meta.env.VITE_VM_SYSTEM_WORKSPACE_ID || 'ws-web-ui').trim();
const DEFAULT_BASE_COMMIT_OID = (import.meta.env.VITE_VM_SYSTEM_BASE_COMMIT_OID || 'web-ui').trim();
const DEFAULT_WORKTREE_PATH = (import.meta.env.VITE_VM_SYSTEM_WORKTREE_PATH || '/tmp').trim();

const DEFAULT_LIMITS = {
  cpu_ms: 2000,
  wall_ms: 5000,
  mem_mb: 128,
  max_events: 50000,
  max_output_kb: 256,
};

const DEFAULT_RESOLVER = {
  roots: ['.'],
  extensions: ['.js', '.mjs'],
  allow_absolute_repo_imports: true,
};

const DEFAULT_RUNTIME = {
  esm: true,
  strict: true,
  console: true,
};

const DEFAULT_TEMPLATE_SPECS: TemplateBootstrapSpec[] = [
  {
    name: 'Default JavaScript',
    engine: 'goja',
    modules: ['console', 'math', 'json', 'date'],
  },
  {
    name: 'Utility Playground',
    engine: 'goja',
    modules: ['console', 'math', 'json', 'array', 'object'],
    libraries: ['lodash'],
  },
  {
    name: 'Library Sandbox',
    engine: 'goja',
    modules: ['console', 'json'],
    libraries: ['dayjs', 'ramda'],
  },
];

const ABSOLUTE_HTTP_URL_RE = /^https?:\/\//i;

function toDate(value: string | undefined | null): Date | undefined {
  if (!value) {
    return undefined;
  }
  const parsed = new Date(value);
  return Number.isNaN(parsed.getTime()) ? undefined : parsed;
}

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
  if (!Array.isArray(value)) {
    return fallback;
  }
  return value.filter((entry): entry is string => typeof entry === 'string');
}

function asArray<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : [];
}

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
    allow_absolute_repo_imports: parseBool(
      obj.allow_absolute_repo_imports,
      DEFAULT_RESOLVER.allow_absolute_repo_imports
    ),
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

function normalizeExecutionResult(rawResult: unknown): unknown {
  const result = toRecord(rawResult);
  if ('json' in result) {
    return result.json;
  }
  return rawResult;
}

function normalizeExecutionError(rawError: unknown): string {
  if (!rawError) {
    return '';
  }
  if (typeof rawError === 'string') {
    return rawError;
  }
  const errorObj = toRecord(rawError);
  if (typeof errorObj.message === 'string') {
    return errorObj.message;
  }
  return JSON.stringify(rawError);
}

function mapExecutionKind(kind: RawExecution['kind']): Execution['kind'] {
  if (kind === 'run_file') {
    return 'run-file';
  }
  return kind;
}

function encodePathSegment(value: string): string {
  return encodeURIComponent(value);
}

function getURL(path: string, query?: Record<string, string | number | undefined>): string {
  const hasAbsoluteBase = ABSOLUTE_HTTP_URL_RE.test(API_BASE_URL);
  const baseURL = hasAbsoluteBase
    ? new URL(API_BASE_URL.endsWith('/') ? API_BASE_URL : `${API_BASE_URL}/`)
    : new URL(window.location.origin);
  const url = hasAbsoluteBase
    ? new URL(path, baseURL)
    : new URL(`${API_BASE_URL}${path}`, baseURL);
  if (query) {
    Object.entries(query).forEach(([key, value]) => {
      if (value === undefined || value === '') {
        return;
      }
      url.searchParams.set(key, String(value));
    });
  }
  if (hasAbsoluteBase) {
    return url.toString();
  }
  return `${url.pathname}${url.search}`;
}

async function request<T>(
  method: string,
  path: string,
  options?: {
    body?: unknown;
    query?: Record<string, string | number | undefined>;
  }
): Promise<T> {
  const response = await fetch(getURL(path, options?.query), {
    method,
    headers: {
      'Content-Type': 'application/json',
    },
    body: options?.body !== undefined ? JSON.stringify(options.body) : undefined,
  });

  if (!response.ok) {
    let message = `HTTP ${response.status}`;
    let code: string | undefined;
    let details: unknown;

    try {
      const envelope = (await response.json()) as APIErrorEnvelope;
      if (envelope.error?.message) {
        message = envelope.error.message;
      }
      code = envelope.error?.code;
      details = envelope.error?.details;
    } catch {
      // ignore parse errors
    }

    throw new APIError(response.status, message, code, details);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return (await response.json()) as T;
}

class VMService {
  private vms: Map<string, VMProfile> = new Map();
  private sessions: Map<string, VMSession> = new Map();
  private executions: Map<string, Execution> = new Map();
  private executionsBySession: Map<string, string[]> = new Map();
  private sessionActivity: Map<string, Date> = new Map();
  private sessionNames: Record<string, string> = {};
  private currentSessionId: string | null = null;
  private initialized = false;
  private initializing: Promise<void> | null = null;

  constructor() {
    this.loadSessionNames();
    this.loadCurrentSessionId();
  }

  async initialize(): Promise<VMProfile | null> {
    await this.ensureInitialized();
    const currentSession = this.getCurrentSession();
    if (currentSession?.vm) {
      return currentSession.vm;
    }
    return this.getVMs()[0] || null;
  }

  private loadSessionNames() {
    try {
      const raw = window.localStorage.getItem(SESSION_NAME_STORAGE_KEY);
      this.sessionNames = raw ? (JSON.parse(raw) as Record<string, string>) : {};
    } catch {
      this.sessionNames = {};
    }
  }

  private persistSessionNames() {
    try {
      window.localStorage.setItem(SESSION_NAME_STORAGE_KEY, JSON.stringify(this.sessionNames));
    } catch {
      // ignore local storage failures
    }
  }

  private loadCurrentSessionId() {
    try {
      this.currentSessionId = window.localStorage.getItem(CURRENT_SESSION_STORAGE_KEY);
    } catch {
      this.currentSessionId = null;
    }
  }

  private persistCurrentSessionId() {
    try {
      if (this.currentSessionId) {
        window.localStorage.setItem(CURRENT_SESSION_STORAGE_KEY, this.currentSessionId);
      } else {
        window.localStorage.removeItem(CURRENT_SESSION_STORAGE_KEY);
      }
    } catch {
      // ignore local storage failures
    }
  }

  private defaultSessionName(raw: RawSession, vm?: VMProfile): string {
    return vm ? `${vm.name} ${raw.id.slice(0, 8)}` : `Session ${raw.id.slice(0, 8)}`;
  }

  private setSessionAlias(sessionId: string, name: string) {
    const trimmed = name.trim();
    if (!trimmed) {
      delete this.sessionNames[sessionId];
    } else {
      this.sessionNames[sessionId] = trimmed;
    }
    this.persistSessionNames();
  }

  private async mapSession(raw: RawSession): Promise<VMSession> {
    const vm = this.vms.get(raw.vm_id) || (await this.fetchTemplateProfile(raw.vm_id));
    const createdAt = toDate(raw.created_at) || new Date();
    const closedAt = toDate(raw.closed_at);
    const lastActivityAt = this.sessionActivity.get(raw.id) || closedAt || createdAt;
    const name = this.sessionNames[raw.id] || this.defaultSessionName(raw, vm);

    return {
      id: raw.id,
      vmId: raw.vm_id,
      vmProfile: vm?.name || 'Template',
      workspaceId: raw.workspace_id,
      baseCommitOID: raw.base_commit_oid,
      worktreePath: raw.worktree_path,
      status: raw.status,
      createdAt,
      closedAt,
      lastActivityAt,
      lastError: raw.last_error,
      name,
      vm,
    };
  }

  private mapExecution(raw: RawExecution, events: ExecutionEvent[]): Execution {
    return {
      id: raw.id,
      sessionId: raw.session_id,
      kind: mapExecutionKind(raw.kind),
      input: raw.input,
      path: raw.path,
      status: raw.status,
      startedAt: toDate(raw.started_at) || new Date(),
      endedAt: toDate(raw.ended_at),
      result: normalizeExecutionResult(raw.result),
      error: normalizeExecutionError(raw.error),
      events,
    };
  }

  private mapEvents(rawEvents: RawExecutionEvent[]): ExecutionEvent[] {
    return rawEvents.map((event) => ({
      seq: event.seq,
      ts: toDate(event.ts) || new Date(),
      type: event.type,
      payload: event.payload,
    }));
  }

  private mapTemplateDetail(detail: RawTemplateDetail): VMProfile {
    const settings = detail.settings || {};
    return {
      id: detail.template.id,
      name: detail.template.name,
      engine: detail.template.engine,
      isActive: detail.template.is_active,
      createdAt: toDate(detail.template.created_at) || new Date(),
      exposedModules: detail.template.exposed_modules || [],
      libraries: detail.template.libraries || [],
      settings: {
        limits: normalizeLimits(settings.limits),
        resolver: normalizeResolver(settings.resolver),
        runtime: normalizeRuntime(settings.runtime),
      },
      capabilities: asArray(detail.capabilities).map((capability) => ({
        id: capability.id,
        kind: capability.kind,
        name: capability.name,
        enabled: capability.enabled,
        config: toRecord(capability.config),
      })),
      startupFiles: asArray(detail.startup_files).map((file) => ({
        id: file.id,
        path: file.path,
        orderIndex: file.order_index,
        mode: file.mode,
      })),
    };
  }

  private async fetchTemplateProfile(templateID: string): Promise<VMProfile> {
    const detail = await request<RawTemplateDetail>('GET', `/api/v1/templates/${templateID}`);
    const profile = this.mapTemplateDetail(detail);
    this.vms.set(profile.id, profile);
    return profile;
  }

  private async createTemplateFromSpec(spec: TemplateBootstrapSpec): Promise<RawTemplate> {
    const created = await request<RawTemplate>('POST', '/api/v1/templates', {
      body: {
        name: spec.name,
        engine: spec.engine,
      },
    });

    const modules = spec.modules || [];
    const libraries = spec.libraries || [];

    await Promise.all(
      modules.map((moduleName) =>
        request('POST', `/api/v1/templates/${created.id}/modules`, {
          body: { name: moduleName },
        })
      )
    );

    await Promise.all(
      libraries.map((libraryName) =>
        request('POST', `/api/v1/templates/${created.id}/libraries`, {
          body: { name: libraryName },
        })
      )
    );

    return created;
  }

  private async bootstrapDefaultTemplates(): Promise<RawTemplate[]> {
    const created: RawTemplate[] = [];
    for (const spec of DEFAULT_TEMPLATE_SPECS) {
      created.push(await this.createTemplateFromSpec(spec));
    }
    return created;
  }

  private async refreshTemplates(): Promise<VMProfile[]> {
    let templates = asArray(await request<RawTemplate[] | null>('GET', '/api/v1/templates'));
    if (templates.length === 0) {
      templates = await this.bootstrapDefaultTemplates();
    }

    const details = await Promise.all(
      templates.map((template) => this.fetchTemplateProfile(template.id))
    );

    const activeIDs = new Set(details.map((template) => template.id));
    for (const id of Array.from(this.vms.keys())) {
      if (!activeIDs.has(id)) {
        this.vms.delete(id);
      }
    }

    return details;
  }

  private async loadSessions(status?: string): Promise<VMSession[]> {
    const rawSessions = asArray(
      await request<RawSession[] | null>('GET', '/api/v1/sessions', {
        query: { status },
      })
    );

    const mapped = await Promise.all(rawSessions.map((rawSession) => this.mapSession(rawSession)));

    const activeIDs = new Set(mapped.map((session) => session.id));
    for (const existingID of Array.from(this.sessions.keys())) {
      if (!activeIDs.has(existingID)) {
        this.sessions.delete(existingID);
      }
    }

    mapped.forEach((session) => {
      this.sessions.set(session.id, session);
    });

    if (this.currentSessionId && !this.sessions.has(this.currentSessionId)) {
      this.currentSessionId = null;
      this.persistCurrentSessionId();
    }

    if (!this.currentSessionId) {
      const firstReady = mapped.find((session) => session.status === 'ready');
      if (firstReady) {
        this.currentSessionId = firstReady.id;
        this.persistCurrentSessionId();
      }
    }

    return mapped.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());
  }

  private async createSessionInternal(templateID: string, name?: string): Promise<VMSession> {
    const rawSession = await request<RawSession>('POST', '/api/v1/sessions', {
      body: {
        template_id: templateID,
        workspace_id: DEFAULT_WORKSPACE_ID,
        base_commit_oid: DEFAULT_BASE_COMMIT_OID,
        worktree_path: DEFAULT_WORKTREE_PATH,
      },
    });

    if (name?.trim()) {
      this.setSessionAlias(rawSession.id, name);
    }

    const session = await this.mapSession(rawSession);
    this.sessions.set(session.id, session);
    this.currentSessionId = session.id;
    this.persistCurrentSessionId();
    return session;
  }

  private async ensureInitialized() {
    if (!this.initialized) {
      if (!this.initializing) {
        this.initializing = (async () => {
          const templates = await this.refreshTemplates();
          const sessions = await this.loadSessions();
          if (sessions.length === 0 && templates.length > 0) {
            await this.createSessionInternal(templates[0].id, 'Default Session');
            await this.loadSessions();
          }
          this.initialized = true;
        })().finally(() => {
          this.initializing = null;
        });
      }
      await this.initializing;
    }
  }

  private async getExecutionEvents(executionID: string): Promise<ExecutionEvent[]> {
    const rawEvents = asArray(
      await request<RawExecutionEvent[] | null>(
        'GET',
        `/api/v1/executions/${executionID}/events`,
        {
          query: { after_seq: 0 },
        }
      )
    );
    return this.mapEvents(rawEvents);
  }

  private async hydrateExecution(raw: RawExecution): Promise<Execution> {
    const events = await this.getExecutionEvents(raw.id);
    const execution = this.mapExecution(raw, events);
    this.executions.set(execution.id, execution);

    const known = this.executionsBySession.get(execution.sessionId) || [];
    if (!known.includes(execution.id)) {
      known.push(execution.id);
      this.executionsBySession.set(execution.sessionId, known);
    }

    return execution;
  }

  private pickDefaultTemplateID(preferredTemplateID?: string): string {
    if (preferredTemplateID) {
      return preferredTemplateID;
    }
    const existing = this.getCurrentSession();
    if (existing?.vmId) {
      return existing.vmId;
    }
    const first = this.getVMs()[0];
    if (!first) {
      throw new Error('No template available');
    }
    return first.id;
  }

  getVMs(): VMProfile[] {
    return Array.from(this.vms.values()).sort((a, b) => a.name.localeCompare(b.name));
  }

  getVM(id: string): VMProfile | undefined {
    return this.vms.get(id);
  }

  async createSession(templateID?: string, name?: string): Promise<VMSession> {
    await this.ensureInitialized();

    const targetTemplateID = this.pickDefaultTemplateID(templateID);
    return this.createSessionInternal(targetTemplateID, name);
  }

  async listSessions(status?: string): Promise<VMSession[]> {
    await this.ensureInitialized();
    return this.loadSessions(status);
  }

  async getSession(sessionId: string): Promise<VMSession | null> {
    await this.ensureInitialized();

    try {
      const rawSession = await request<RawSession>('GET', `/api/v1/sessions/${sessionId}`);
      const session = await this.mapSession(rawSession);
      this.sessions.set(session.id, session);
      return session;
    } catch (error) {
      if (error instanceof APIError && error.code === 'SESSION_NOT_FOUND') {
        this.sessions.delete(sessionId);
        if (this.currentSessionId === sessionId) {
          this.currentSessionId = null;
          this.persistCurrentSessionId();
        }
        return null;
      }
      throw error;
    }
  }

  getCurrentSession(): VMSession | null {
    if (!this.currentSessionId) {
      return null;
    }
    return this.sessions.get(this.currentSessionId) || null;
  }

  async setCurrentSession(sessionId: string): Promise<void> {
    await this.ensureInitialized();
    let session = this.sessions.get(sessionId) || null;
    if (!session) {
      session = await this.getSession(sessionId);
    }
    if (!session) {
      throw new Error('Session not found');
    }
    if (session.status !== 'ready') {
      throw new Error('Session is not ready');
    }
    this.currentSessionId = sessionId;
    this.persistCurrentSessionId();
  }

  async closeSession(sessionId: string): Promise<void> {
    await this.ensureInitialized();

    const rawSession = await request<RawSession>('POST', `/api/v1/sessions/${sessionId}/close`, {
      body: {},
    });
    const session = await this.mapSession(rawSession);
    this.sessions.set(session.id, session);

    if (this.currentSessionId === sessionId) {
      this.currentSessionId = null;
      this.persistCurrentSessionId();
    }
  }

  async deleteSession(sessionId: string): Promise<void> {
    await this.ensureInitialized();

    const rawSession = await request<RawSession>('DELETE', `/api/v1/sessions/${sessionId}`);
    const session = await this.mapSession(rawSession);
    this.sessions.set(session.id, session);

    if (this.currentSessionId === sessionId) {
      this.currentSessionId = null;
      this.persistCurrentSessionId();
    }
  }

  async executeREPL(code: string, sessionId?: string): Promise<Execution> {
    await this.ensureInitialized();

    const targetSessionID = sessionId || this.currentSessionId;
    if (!targetSessionID) {
      throw new Error('No active session');
    }

    const rawExecution = await request<RawExecution>('POST', '/api/v1/executions/repl', {
      body: {
        session_id: targetSessionID,
        input: code,
      },
    });

    const execution = await this.hydrateExecution(rawExecution);
    this.sessionActivity.set(targetSessionID, new Date());

    const session = this.sessions.get(targetSessionID);
    if (session) {
      session.lastActivityAt = new Date();
      this.sessions.set(targetSessionID, session);
    }

    return execution;
  }

  async getExecution(id: string): Promise<Execution | null> {
    await this.ensureInitialized();

    try {
      const rawExecution = await request<RawExecution>('GET', `/api/v1/executions/${id}`);
      return await this.hydrateExecution(rawExecution);
    } catch (error) {
      if (error instanceof APIError && error.code === 'EXECUTION_NOT_FOUND') {
        this.executions.delete(id);
        return null;
      }
      throw error;
    }
  }

  async getExecutionsBySession(sessionId: string): Promise<Execution[]> {
    await this.ensureInitialized();

    const rawExecutions = asArray(
      await request<RawExecution[] | null>('GET', '/api/v1/executions', {
        query: {
          session_id: sessionId,
          limit: 50,
        },
      })
    );

    const executions = await Promise.all(rawExecutions.map((rawExecution) => this.hydrateExecution(rawExecution)));
    this.executionsBySession.set(
      sessionId,
      executions.map((execution) => execution.id)
    );

    return executions.sort((a, b) => a.startedAt.getTime() - b.startedAt.getTime());
  }

  async getAllExecutions(): Promise<Execution[]> {
    await this.ensureInitialized();
    return Array.from(this.executions.values()).sort((a, b) => a.startedAt.getTime() - b.startedAt.getTime());
  }

  getRecentExecutions(limit = 10): Execution[] {
    return Array.from(this.executions.values())
      .sort((a, b) => b.startedAt.getTime() - a.startedAt.getTime())
      .slice(0, limit);
  }

  async updateTemplateModules(templateID: string, modules: string[]): Promise<VMProfile> {
    await this.ensureInitialized();

    const currentModules = asArray(
      await request<string[] | null>('GET', `/api/v1/templates/${templateID}/modules`)
    );
    const wanted = new Set(modules);
    const existing = new Set(currentModules);

    const toAdd = modules.filter((module) => !existing.has(module));
    const toRemove = currentModules.filter((module) => !wanted.has(module));

    await Promise.all(
      toAdd.map((module) =>
        request('POST', `/api/v1/templates/${templateID}/modules`, {
          body: { name: module },
        })
      )
    );

    await Promise.all(
      toRemove.map((module) =>
        request('DELETE', `/api/v1/templates/${templateID}/modules/${encodePathSegment(module)}`)
      )
    );

    const updated = await this.fetchTemplateProfile(templateID);
    this.refreshSessionsForTemplate(updated);
    return updated;
  }

  async updateTemplateLibraries(templateID: string, libraries: string[]): Promise<VMProfile> {
    await this.ensureInitialized();

    const currentLibraries = asArray(
      await request<string[] | null>('GET', `/api/v1/templates/${templateID}/libraries`)
    );
    const wanted = new Set(libraries);
    const existing = new Set(currentLibraries);

    const toAdd = libraries.filter((library) => !existing.has(library));
    const toRemove = currentLibraries.filter((library) => !wanted.has(library));

    await Promise.all(
      toAdd.map((library) =>
        request('POST', `/api/v1/templates/${templateID}/libraries`, {
          body: { name: library },
        })
      )
    );

    await Promise.all(
      toRemove.map((library) =>
        request('DELETE', `/api/v1/templates/${templateID}/libraries/${encodePathSegment(library)}`)
      )
    );

    const updated = await this.fetchTemplateProfile(templateID);
    this.refreshSessionsForTemplate(updated);
    return updated;
  }

  private refreshSessionsForTemplate(vm: VMProfile) {
    this.sessions.forEach((session) => {
      if (session.vmId === vm.id) {
        this.sessions.set(session.id, {
          ...session,
          vm,
          vmProfile: vm.name,
        });
      }
    });
  }
}

export const vmService = new VMService();
