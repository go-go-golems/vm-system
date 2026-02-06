// Mock VM service that simulates the Go backend
import { nanoid } from 'nanoid';

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
  config: Record<string, any>;
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
  status: 'starting' | 'ready' | 'crashed' | 'closed';
  createdAt: Date;
  closedAt?: Date;
  lastActivityAt: Date;
  name: string;
  vm?: VMProfile; // Reference to the VM configuration
}

export interface Execution {
  id: string;
  sessionId: string;
  kind: 'repl' | 'run-file';
  input?: string;
  path?: string;
  status: 'running' | 'ok' | 'error';
  startedAt: Date;
  endedAt?: Date;
  result?: any;
  error?: string;
  events: ExecutionEvent[];
}

export interface ExecutionEvent {
  seq: number;
  ts: Date;
  type: 'input_echo' | 'console' | 'value' | 'exception';
  payload: any;
}

// NOTE: This is a mock frontend implementation for demonstration.
// In the actual Go backend with goja:
// 1. Libraries are downloaded from CDN and cached locally (.vm-cache/libraries/)
// 2. When a session is created, configured libraries are loaded into the goja runtime
// 3. Library code is executed via runtime.RunString() making globals (like _) available
// 4. Test confirmed: Lodash successfully loads and all functions work in goja

// Built-in modules that can be exposed to VMs
export const BUILTIN_MODULES = [
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

// Built-in libraries that can be loaded
export const BUILTIN_LIBRARIES = [
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

// Preset examples
export const PRESET_EXAMPLES = [
  {
    id: 'hello-world',
    name: 'Hello World',
    description: 'Simple console.log example',
    code: `console.log("Hello from VM!");
const greeting = "Welcome to the VM System";
console.log(greeting);
greeting;`,
  },
  {
    id: 'math-operations',
    name: 'Math Operations',
    description: 'Basic arithmetic and math functions',
    code: `const a = 10;
const b = 20;
const sum = a + b;
const product = a * b;

console.log("Sum:", sum);
console.log("Product:", product);
console.log("Square root of 16:", Math.sqrt(16));

{ sum, product, sqrt: Math.sqrt(16) };`,
  },
  {
    id: 'array-operations',
    name: 'Array Operations',
    description: 'Working with arrays',
    code: `const numbers = [1, 2, 3, 4, 5];
const doubled = numbers.map(n => n * 2);
const sum = numbers.reduce((a, b) => a + b, 0);
const filtered = numbers.filter(n => n > 2);

console.log("Original:", numbers);
console.log("Doubled:", doubled);
console.log("Sum:", sum);
console.log("Filtered (>2):", filtered);

{ doubled, sum, filtered };`,
  },
  {
    id: 'object-manipulation',
    name: 'Object Manipulation',
    description: 'Creating and manipulating objects',
    code: `const person = {
  name: "Alice",
  age: 30,
  city: "San Francisco"
};

const updatedPerson = {
  ...person,
  age: 31,
  occupation: "Engineer"
};

console.log("Original:", person);
console.log("Updated:", updatedPerson);

updatedPerson;`,
  },
  {
    id: 'async-simulation',
    name: 'Async Simulation',
    description: 'Simulating asynchronous operations',
    code: `// Note: Real async/await not supported in this VM
// This demonstrates synchronous code structure

function fetchData() {
  console.log("Fetching data...");
  return { id: 1, name: "Sample Data" };
}

const data = fetchData();
console.log("Data received:", data);

data;`,
  },
  {
    id: 'error-handling',
    name: 'Error Handling',
    description: 'Try-catch error handling',
    code: `try {
  console.log("Attempting operation...");
  const result = 10 / 2;
  console.log("Result:", result);
  
  // Uncomment to trigger error:
  // throw new Error("Something went wrong!");
  
  result;
} catch (error) {
  console.log("Error caught:", error.message);
  null;
}`,
  },
  {
    id: 'string-operations',
    name: 'String Operations',
    description: 'String manipulation methods',
    code: `const text = "JavaScript VM System";
const upper = text.toUpperCase();
const lower = text.toLowerCase();
const words = text.split(" ");
const reversed = text.split("").reverse().join("");

console.log("Original:", text);
console.log("Uppercase:", upper);
console.log("Lowercase:", lower);
console.log("Words:", words);
console.log("Reversed:", reversed);

{ upper, lower, words, reversed };`,
  },
  {
    id: 'function-demo',
    name: 'Function Demo',
    description: 'Defining and using functions',
    code: `function fibonacci(n) {
  if (n <= 1) return n;
  return fibonacci(n - 1) + fibonacci(n - 2);
}

const fib10 = fibonacci(10);
console.log("Fibonacci(10):", fib10);

const factorial = (n) => {
  if (n <= 1) return 1;
  return n * factorial(n - 1);
};

const fact5 = factorial(5);
console.log("Factorial(5):", fact5);

{ fib10, fact5 };`,
  },
  {
    id: 'lodash-demo',
    name: 'Lodash Utilities',
    description: 'Using Lodash library functions (requires lodash library)',
    code: `// This example REQUIRES the 'lodash' library to be enabled in VM Config
// It will fail if lodash is not configured

if (typeof _ === 'undefined') {
  throw new Error('Lodash library not loaded! Enable it in VM Config > Libraries.');
}

const users = [
  { name: 'Alice', age: 30, active: true },
  { name: 'Bob', age: 25, active: false },
  { name: 'Charlie', age: 35, active: true },
  { name: 'David', age: 28, active: true }
];

// Use actual Lodash functions
const activeUsers = _.filter(users, { active: true });
console.log("Active users (_.filter):", activeUsers);

// Group by age range using Lodash
const grouped = _.groupBy(users, user => user.age < 30 ? '20s' : '30s');
console.log("Grouped by age (_.groupBy):", grouped);

// Get names using Lodash map
const names = _.map(users, 'name');
console.log("Names (_.map):", names);

// Find user using Lodash
const charlie = _.find(users, { name: 'Charlie' });
console.log("Found Charlie (_.find):", charlie);

{ activeUsers, grouped, names, charlie };`,
  },
  {
    id: 'date-manipulation',
    name: 'Date Manipulation',
    description: 'Working with dates using built-in Date module',
    code: `// Using built-in Date module
const now = new Date();
console.log("Current date:", now.toISOString());

// Create specific dates
const birthday = new Date('2024-01-15');
console.log("Birthday:", birthday.toDateString());

// Date arithmetic
const tomorrow = new Date(now);
tomorrow.setDate(tomorrow.getDate() + 1);
console.log("Tomorrow:", tomorrow.toDateString());

// Calculate difference
const diff = tomorrow - now;
const hours = Math.floor(diff / (1000 * 60 * 60));
console.log("Hours until tomorrow:", hours);

{ now: now.toISOString(), tomorrow: tomorrow.toISOString(), hoursUntil: hours };`,
  },
  {
    id: 'json-processing',
    name: 'JSON Processing',
    description: 'Parse and stringify JSON data',
    code: `// JSON module demo
const data = {
  user: "Alice",
  preferences: {
    theme: "dark",
    notifications: true
  },
  tags: ["admin", "developer"]
};

// Stringify
const jsonString = JSON.stringify(data, null, 2);
console.log("JSON string:");
console.log(jsonString);

// Parse
const parsed = JSON.parse(jsonString);
console.log("Parsed back:", parsed);

// Verify
const isEqual = JSON.stringify(data) === JSON.stringify(parsed);
console.log("Data preserved:", isEqual);

{ original: data, parsed, isEqual };`,
  },
  {
    id: 'advanced-array',
    name: 'Advanced Array Operations',
    description: 'Complex array transformations and reductions',
    code: `// Advanced array operations
const products = [
  { name: 'Laptop', price: 1200, category: 'Electronics' },
  { name: 'Mouse', price: 25, category: 'Electronics' },
  { name: 'Desk', price: 300, category: 'Furniture' },
  { name: 'Chair', price: 200, category: 'Furniture' },
  { name: 'Monitor', price: 400, category: 'Electronics' }
];

// Calculate total by category
const totalByCategory = products.reduce((acc, product) => {
  if (!acc[product.category]) {
    acc[product.category] = 0;
  }
  acc[product.category] += product.price;
  return acc;
}, {});
console.log("Total by category:", totalByCategory);

// Find expensive items (>$300)
const expensive = products.filter(p => p.price > 300);
console.log("Expensive items:", expensive);

// Apply discount
const discounted = products.map(p => ({
  ...p,
  salePrice: Math.round(p.price * 0.9)
}));
console.log("With 10% discount:", discounted);

{ totalByCategory, expensive, discounted };`,
  },
  {
    id: 'vm-module-check',
    name: 'VM Module Check',
    description: 'Check which modules are available in the current VM',
    code: `// Check available modules
console.log("Checking VM capabilities...");

// Test console module
try {
  console.log("✓ Console module available");
} catch (e) {
  console.log("✗ Console module not available");
}

// Test Math module
try {
  const result = Math.sqrt(16);
  console.log("✓ Math module available, sqrt(16) =", result);
} catch (e) {
  console.log("✗ Math module not available");
}

// Test JSON module
try {
  const obj = { test: true };
  const str = JSON.stringify(obj);
  console.log("✓ JSON module available");
} catch (e) {
  console.log("✗ JSON module not available");
}

// Test Array methods
try {
  const arr = [1, 2, 3].map(x => x * 2);
  console.log("✓ Array methods available:", arr);
} catch (e) {
  console.log("✗ Array methods not available");
}

"Module check complete";`,
  },
  {
    id: 'vm-library-check',
    name: 'VM Library Check',
    description: 'Check which external libraries are loaded',
    code: `// Check loaded libraries
console.log("Checking loaded libraries...");

const libraries = [];

// Check Lodash
if (typeof _ !== 'undefined') {
  console.log("✓ Lodash available");
  libraries.push('lodash');
} else {
  console.log("✗ Lodash not loaded (enable in VM Config)");
}

// Check Moment.js
if (typeof moment !== 'undefined') {
  console.log("✓ Moment.js available");
  libraries.push('moment');
} else {
  console.log("✗ Moment.js not loaded");
}

// Check Ramda
if (typeof R !== 'undefined') {
  console.log("✓ Ramda available");
  libraries.push('ramda');
} else {
  console.log("✗ Ramda not loaded");
}

// Check Day.js
if (typeof dayjs !== 'undefined') {
  console.log("✓ Day.js available");
  libraries.push('dayjs');
} else {
  console.log("✗ Day.js not loaded");
}

// Check Zustand
if (typeof zustand !== 'undefined') {
  console.log("✓ Zustand available");
  libraries.push('zustand');
} else {
  console.log("✗ Zustand not loaded");
}

console.log("\nLoaded libraries:", libraries.length);

{ loadedLibraries: libraries, count: libraries.length };`,
  },
  {
    id: 'zustand-state',
    name: 'Zustand State Management',
    description: 'Using Zustand for state management (requires zustand library)',
    code: `// This example REQUIRES the 'zustand' library to be enabled in VM Config
// It will fail if zustand is not configured

if (typeof zustand === 'undefined') {
  throw new Error('Zustand library not loaded! Enable it in VM Config > Libraries.');
}

// Note: In a real browser/node environment, zustand would be used differently
// This demonstrates the library availability check
console.log('Zustand library loaded:', typeof zustand);

// Simulating Zustand state management pattern for demo

// Create a simple state store
const createStore = (initialState) => {
  let state = initialState;
  const listeners = [];

  return {
    getState: () => state,
    setState: (partial) => {
      state = { ...state, ...partial };
      listeners.forEach(listener => listener(state));
    },
    subscribe: (listener) => {
      listeners.push(listener);
      return () => {
        const index = listeners.indexOf(listener);
        if (index > -1) listeners.splice(index, 1);
      };
    }
  };
};

// Create a counter store
const counterStore = createStore({
  count: 0,
  increment: function() {
    this.setState({ count: this.getState().count + 1 });
  },
  decrement: function() {
    this.setState({ count: this.getState().count - 1 });
  }
});

console.log("Initial state:", counterStore.getState());

// Subscribe to changes
counterStore.subscribe((state) => {
  console.log("State changed:", state);
});

// Update state
counterStore.setState({ count: counterStore.getState().count + 1 });
counterStore.setState({ count: counterStore.getState().count + 5 });
counterStore.setState({ count: counterStore.getState().count - 2 });

const finalState = counterStore.getState();
console.log("Final state:", finalState);

finalState;`,
  },
  {
    id: 'functional-ramda',
    name: 'Functional Programming with Ramda',
    description: 'Using Ramda for functional programming (requires ramda library)',
    code: `// This example REQUIRES the 'ramda' library to be enabled in VM Config
// It will fail if ramda is not configured

if (typeof R === 'undefined') {
  throw new Error('Ramda library not loaded! Enable it in VM Config > Libraries.');
}

const users = [
  { id: 1, name: 'Alice', age: 30, role: 'admin' },
  { id: 2, name: 'Bob', age: 25, role: 'user' },
  { id: 3, name: 'Charlie', age: 35, role: 'admin' },
  { id: 4, name: 'David', age: 28, role: 'user' }
];

// Use actual Ramda functions
const isAdmin = R.propEq('role', 'admin');
const isOver30 = R.propSatisfies(age => age > 30, 'age');
const getName = R.prop('name');

// Filter admins using Ramda
const admins = R.filter(isAdmin, users);
console.log("Admins (R.filter):", R.map(getName, admins));

// Filter users over 30
const over30 = R.filter(isOver30, users);
console.log("Over 30 (R.filter):", R.map(getName, over30));

// Compose: admins over 30 using Ramda composition
const adminAndOver30 = R.both(isAdmin, isOver30);
const adminOver30 = R.filter(adminAndOver30, users);
console.log("Admin over 30 (R.both + R.filter):", R.map(getName, adminOver30));

// Transform data using Ramda
const createSummary = user => ({
  name: user.name,
  summary: \`\${user.name} (\${user.age}) - \${user.role}\`
});
const userSummaries = R.map(createSummary, users);

console.log("Summaries (R.map):", userSummaries);

// Demonstrate pipe
const getAdminNames = R.pipe(
  R.filter(isAdmin),
  R.map(getName),
  R.join(', ')
);
console.log("Admin names (R.pipe):", getAdminNames(users));

{ admins: admins.length, over30: over30.length, adminOver30, userSummaries };`,
  },
  {
    id: 'vm-capability-demo',
    name: 'VM Capability Demo',
    description: 'Demonstrate different VM capabilities working together',
    code: `// Comprehensive VM capability demonstration
console.log("=== VM Capability Demo ===");

// 1. Console module
console.log("\n1. Console logging:");
console.log("Standard log");
console.log("Multiple", "arguments", 123);

// 2. Math module
console.log("\n2. Math operations:");
const calculations = {
  sqrt: Math.sqrt(144),
  pow: Math.pow(2, 8),
  random: Math.floor(Math.random() * 100),
  pi: Math.PI
};
console.log("Calculations:", calculations);

// 3. Array methods
console.log("\n3. Array operations:");
const numbers = [1, 2, 3, 4, 5];
const processed = {
  doubled: numbers.map(n => n * 2),
  filtered: numbers.filter(n => n > 2),
  sum: numbers.reduce((a, b) => a + b, 0)
};
console.log("Array processing:", processed);

// 4. Object methods
console.log("\n4. Object operations:");
const obj = { a: 1, b: 2, c: 3 };
const objOps = {
  keys: Object.keys(obj),
  values: Object.values(obj),
  entries: Object.entries(obj)
};
console.log("Object operations:", objOps);

// 5. JSON module
console.log("\n5. JSON operations:");
const data = { name: "Test", values: [1, 2, 3] };
const jsonStr = JSON.stringify(data);
const parsed = JSON.parse(jsonStr);
console.log("JSON round-trip successful:", JSON.stringify(data) === JSON.stringify(parsed));

// 6. Date module
console.log("\n6. Date operations:");
const now = new Date();
console.log("Current time:", now.toISOString());

console.log("\n=== Demo Complete ===");

{ calculations, processed, objOps, timestamp: now.toISOString() };`,
  },
];

// Mock VM service
class VMService {
  private vms: Map<string, VMProfile> = new Map();
  private sessions: Map<string, VMSession> = new Map();
  private executions: Map<string, Execution> = new Map();
  private executionsBySession: Map<string, string[]> = new Map();
  private currentSessionId: string | null = null;
  private gcInterval: number | null = null;
  private readonly IDLE_TIMEOUT_MS = 5 * 60 * 1000; // 5 minutes
  private readonly GC_CHECK_INTERVAL_MS = 60 * 1000; // 1 minute

  constructor() {
    // Create default VM profile
    this.createDefaultVM();
    this.startGarbageCollection();
  }

  private startGarbageCollection() {
    if (this.gcInterval) return;

    this.gcInterval = window.setInterval(() => {
      const now = Date.now();
      const sessionsToClose: string[] = [];

      this.sessions.forEach((session, sessionId) => {
        if (
          session.status === 'ready' &&
          now - session.lastActivityAt.getTime() > this.IDLE_TIMEOUT_MS
        ) {
          sessionsToClose.push(sessionId);
        }
      });

      sessionsToClose.forEach((sessionId) => {
        console.log(`[GC] Closing idle session: ${sessionId}`);
        this.closeSession(sessionId);
      });
    }, this.GC_CHECK_INTERVAL_MS);
  }

  private stopGarbageCollection() {
    if (this.gcInterval) {
      clearInterval(this.gcInterval);
      this.gcInterval = null;
    }
  }

  private createDefaultVM() {
    const vm: VMProfile = {
      id: nanoid(),
      name: 'Default VM',
      engine: 'goja',
      isActive: true,
      createdAt: new Date(),
      exposedModules: ['console', 'math', 'json', 'array', 'object'],
      libraries: [],
      settings: {
        limits: {
          cpu_ms: 2000,
          wall_ms: 5000,
          mem_mb: 128,
          max_events: 50000,
          max_output_kb: 256,
        },
        resolver: {
          roots: ['.'],
          extensions: ['.js', '.mjs'],
          allow_absolute_repo_imports: true,
        },
        runtime: {
          esm: true,
          strict: true,
          console: true,
        },
      },
      capabilities: [
        {
          id: nanoid(),
          kind: 'module',
          name: 'console',
          enabled: true,
          config: {},
        },
      ],
      startupFiles: [],
    };

    this.vms.set(vm.id, vm);

    // Create a default session
    const session = this.createSessionSync(vm.id, 'default-workspace', 'Default Session');
    this.currentSessionId = session.id;
  }

  private createSessionSync(vmId: string, workspaceId: string, name: string): VMSession {
    const vm = this.vms.get(vmId);
    const session: VMSession = {
      id: nanoid(),
      vmId,
      vmProfile: vm?.name || 'Unknown VM',
      workspaceId,
      status: 'ready',
      createdAt: new Date(),
      lastActivityAt: new Date(),
      name,
      vm, // Include full VM configuration
    };

    this.sessions.set(session.id, session);
    this.executionsBySession.set(session.id, []);
    return session;
  }

  getVMs(): VMProfile[] {
    return Array.from(this.vms.values());
  }

  getVM(id: string): VMProfile | undefined {
    return this.vms.get(id);
  }

  async createSession(name?: string): Promise<VMSession> {
    const defaultVM = Array.from(this.vms.values())[0];
    if (!defaultVM) {
      throw new Error('No VM profile available');
    }

    const sessionName = name || `Session ${this.sessions.size + 1}`;
    const session = this.createSessionSync(
      defaultVM.id,
      `workspace-${nanoid()}`,
      sessionName
    );

    return session;
  }

  async listSessions(): Promise<VMSession[]> {
    return Array.from(this.sessions.values()).sort(
      (a, b) => b.lastActivityAt.getTime() - a.lastActivityAt.getTime()
    );
  }

  async getSession(sessionId: string): Promise<VMSession | null> {
    return this.sessions.get(sessionId) || null;
  }

  getCurrentSession(): VMSession | null {
    if (!this.currentSessionId) return null;
    return this.sessions.get(this.currentSessionId) || null;
  }

  async setCurrentSession(sessionId: string): Promise<void> {
    const session = this.sessions.get(sessionId);
    if (!session) {
      throw new Error('Session not found');
    }
    if (session.status !== 'ready') {
      throw new Error('Session is not ready');
    }
    this.currentSessionId = sessionId;
    this.touchSession(sessionId);
  }

  async closeSession(sessionId: string): Promise<void> {
    const session = this.sessions.get(sessionId);
    if (!session) return;

    session.status = 'closed';
    session.closedAt = new Date();
    this.sessions.set(sessionId, session);

    // If this was the current session, clear it
    if (this.currentSessionId === sessionId) {
      this.currentSessionId = null;
    }
  }

  async deleteSession(sessionId: string): Promise<void> {
    this.sessions.delete(sessionId);
    this.executionsBySession.delete(sessionId);

    if (this.currentSessionId === sessionId) {
      this.currentSessionId = null;
    }
  }

  private touchSession(sessionId: string) {
    const session = this.sessions.get(sessionId);
    if (session) {
      session.lastActivityAt = new Date();
      this.sessions.set(sessionId, session);
    }
  }

  async executeREPL(code: string, sessionId?: string): Promise<Execution> {
    const targetSessionId = sessionId || this.currentSessionId;
    if (!targetSessionId) {
      throw new Error('No active session');
    }

    const session = this.sessions.get(targetSessionId);
    if (!session) {
      throw new Error('Session not found');
    }

    if (session.status !== 'ready') {
      throw new Error('Session is not ready');
    }

    this.touchSession(targetSessionId);

    const execution: Execution = {
      id: nanoid(),
      sessionId: targetSessionId,
      kind: 'repl',
      input: code,
      status: 'running',
      startedAt: new Date(),
      events: [],
    };

    this.executions.set(execution.id, execution);

    // Track execution by session
    const sessionExecutions = this.executionsBySession.get(targetSessionId) || [];
    sessionExecutions.push(execution.id);
    this.executionsBySession.set(targetSessionId, sessionExecutions);

    // Simulate execution with setTimeout
    await new Promise((resolve) => setTimeout(resolve, 100));

    // Add input echo event
    execution.events.push({
      seq: 1,
      ts: new Date(),
      type: 'input_echo',
      payload: { text: code },
    });

    try {
      // Execute code using eval (in a real implementation, this would use goja)
      const result = this.safeEval(code, execution);

      execution.status = 'ok';
      execution.endedAt = new Date();
      execution.result = result;

      // Add value event
      execution.events.push({
        seq: execution.events.length + 1,
        ts: new Date(),
        type: 'value',
        payload: {
          type: typeof result,
          preview: String(result),
          json: result,
        },
      });
    } catch (error: any) {
      execution.status = 'error';
      execution.endedAt = new Date();
      execution.error = error.message;

      // Add exception event
      execution.events.push({
        seq: execution.events.length + 1,
        ts: new Date(),
        type: 'exception',
        payload: {
          message: error.message,
          stack: error.stack,
        },
      });
    }

    return execution;
  }

  private safeEval(code: string, execution: Execution): any {
    // Create a custom console that captures output
    const consoleOutput: string[] = [];
    const customConsole = {
      log: (...args: any[]) => {
        const message = args.map((arg) => String(arg)).join(' ');
        consoleOutput.push(message);

        // Add console event
        execution.events.push({
          seq: execution.events.length + 1,
          ts: new Date(),
          type: 'console',
          payload: {
            level: 'log',
            text: message,
          },
        });
      },
    };

    // Create a safe execution context
    const context = {
      console: customConsole,
      Math,
      Date,
      Array,
      Object,
      String,
      Number,
      Boolean,
      JSON,
    };

    // Wrap code in a function to create a scope
    const wrappedCode = `
      with (context) {
        return (function() {
          ${code}
        })();
      }
    `;

    try {
      // Execute the code
      const func = new Function('context', wrappedCode);
      return func(context);
    } catch (error) {
      throw error;
    }
  }

  async getExecution(id: string): Promise<Execution | null> {
    return this.executions.get(id) || null;
  }

  async getExecutionsBySession(sessionId: string): Promise<Execution[]> {
    const executionIds = this.executionsBySession.get(sessionId) || [];
    return executionIds
      .map((id) => this.executions.get(id))
      .filter((exec): exec is Execution => exec !== undefined)
      .sort((a, b) => a.startedAt.getTime() - b.startedAt.getTime());
  }

  async getAllExecutions(): Promise<Execution[]> {
    return Array.from(this.executions.values()).sort(
      (a, b) => a.startedAt.getTime() - b.startedAt.getTime()
    );
  }

  getRecentExecutions(limit: number = 10): Execution[] {
    return Array.from(this.executions.values())
      .sort((a, b) => b.startedAt.getTime() - a.startedAt.getTime())
      .slice(0, limit);
  }

  destroy() {
    this.stopGarbageCollection();
  }
}

export const vmService = new VMService();
