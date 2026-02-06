// Mock VM service that simulates the Go backend
import { nanoid } from 'nanoid';

export interface VMProfile {
  id: string;
  name: string;
  engine: string;
  isActive: boolean;
  createdAt: Date;
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
  workspaceId: string;
  status: 'starting' | 'ready' | 'crashed' | 'closed';
  createdAt: Date;
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
];

// Mock VM service
class VMService {
  private vms: Map<string, VMProfile> = new Map();
  private sessions: Map<string, VMSession> = new Map();
  private executions: Map<string, Execution> = new Map();
  private currentSession: VMSession | null = null;

  constructor() {
    // Create default VM profile
    this.createDefaultVM();
  }

  private createDefaultVM() {
    const vm: VMProfile = {
      id: nanoid(),
      name: 'Default VM',
      engine: 'goja',
      isActive: true,
      createdAt: new Date(),
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

    // Create a session for this VM
    const session: VMSession = {
      id: nanoid(),
      vmId: vm.id,
      workspaceId: 'default-workspace',
      status: 'ready',
      createdAt: new Date(),
    };

    this.sessions.set(session.id, session);
    this.currentSession = session;
  }

  getVMs(): VMProfile[] {
    return Array.from(this.vms.values());
  }

  getVM(id: string): VMProfile | undefined {
    return this.vms.get(id);
  }

  getCurrentSession(): VMSession | null {
    return this.currentSession;
  }

  async executeREPL(code: string): Promise<Execution> {
    if (!this.currentSession) {
      throw new Error('No active session');
    }

    const execution: Execution = {
      id: nanoid(),
      sessionId: this.currentSession.id,
      kind: 'repl',
      input: code,
      status: 'running',
      startedAt: new Date(),
      events: [],
    };

    this.executions.set(execution.id, execution);

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

  getExecution(id: string): Execution | undefined {
    return this.executions.get(id);
  }

  getRecentExecutions(limit: number = 10): Execution[] {
    return Array.from(this.executions.values())
      .sort((a, b) => b.startedAt.getTime() - a.startedAt.getTime())
      .slice(0, limit);
  }
}

export const vmService = new VMService();
