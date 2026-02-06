import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import {
  BookOpen,
  Code2,
  Cpu,
  FileCode,
  Home,
  Layers,
  Play,
  Settings,
  Terminal,
  Zap,
} from 'lucide-react';
import { Link } from 'wouter';

export default function Docs() {
  return (
    <div className="min-h-screen bg-slate-950">
      {/* Header */}
      <header className="border-b border-slate-800 bg-slate-900/50 backdrop-blur sticky top-0 z-10">
        <div className="container py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center">
                <Terminal className="w-6 h-6 text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold text-slate-100">VM System Documentation</h1>
                <p className="text-sm text-slate-400">Complete guide and API reference</p>
              </div>
            </div>

            <Link href="/">
              <Button variant="outline" className="bg-slate-900 border-slate-700 text-slate-300">
                <Home className="w-4 h-4 mr-2" />
                Back to Editor
              </Button>
            </Link>
          </div>
        </div>
      </header>

      {/* Main content */}
      <main className="container py-8">
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
          {/* Sidebar navigation */}
          <aside className="lg:col-span-1">
            <div className="sticky top-24">
              <nav className="space-y-1">
                <a
                  href="#overview"
                  className="block px-3 py-2 text-sm text-slate-300 hover:bg-slate-800 rounded-md transition-colors"
                >
                  Overview
                </a>
                <a
                  href="#getting-started"
                  className="block px-3 py-2 text-sm text-slate-300 hover:bg-slate-800 rounded-md transition-colors"
                >
                  Getting Started
                </a>
                <a
                  href="#features"
                  className="block px-3 py-2 text-sm text-slate-300 hover:bg-slate-800 rounded-md transition-colors"
                >
                  Features
                </a>
                <a
                  href="#examples"
                  className="block px-3 py-2 text-sm text-slate-300 hover:bg-slate-800 rounded-md transition-colors"
                >
                  Code Examples
                </a>
                <a
                  href="#vm-architecture"
                  className="block px-3 py-2 text-sm text-slate-300 hover:bg-slate-800 rounded-md transition-colors"
                >
                  VM Architecture
                </a>
                <a
                  href="#api-reference"
                  className="block px-3 py-2 text-sm text-slate-300 hover:bg-slate-800 rounded-md transition-colors"
                >
                  API Reference
                </a>
                <a
                  href="#limitations"
                  className="block px-3 py-2 text-sm text-slate-300 hover:bg-slate-800 rounded-md transition-colors"
                >
                  Limitations
                </a>
              </nav>
            </div>
          </aside>

          {/* Documentation content */}
          <div className="lg:col-span-3 space-y-8">
            {/* Overview */}
            <section id="overview">
              <h2 className="text-3xl font-bold text-slate-100 mb-4 flex items-center gap-2">
                <BookOpen className="w-8 h-8 text-blue-500" />
                Overview
              </h2>
              <Card className="bg-slate-900 border-slate-800">
                <CardContent className="pt-6 space-y-4 text-slate-300">
                  <p className="leading-relaxed">
                    The VM System is a JavaScript execution environment built on the{' '}
                    <code className="px-2 py-1 bg-slate-800 rounded text-blue-400 font-mono text-sm">
                      goja
                    </code>{' '}
                    runtime. It provides a sandboxed environment for running JavaScript code with
                    configurable resource limits, capability management, and detailed execution
                    tracking.
                  </p>
                  <p className="leading-relaxed">
                    This web interface allows you to interactively test JavaScript code, explore
                    preset examples, and monitor execution results in real-time. The system captures
                    console output, return values, and exceptions, providing a comprehensive view of
                    your code's behavior.
                  </p>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-6">
                    <div className="flex items-start gap-3 p-4 bg-slate-800/50 rounded-lg">
                      <Zap className="w-5 h-5 text-blue-500 mt-1" />
                      <div>
                        <h3 className="font-semibold text-slate-200 mb-1">Fast Execution</h3>
                        <p className="text-sm text-slate-400">
                          Powered by goja for near-native JavaScript performance
                        </p>
                      </div>
                    </div>
                    <div className="flex items-start gap-3 p-4 bg-slate-800/50 rounded-lg">
                      <Settings className="w-5 h-5 text-blue-500 mt-1" />
                      <div>
                        <h3 className="font-semibold text-slate-200 mb-1">Configurable</h3>
                        <p className="text-sm text-slate-400">
                          Fine-grained control over capabilities and resource limits
                        </p>
                      </div>
                    </div>
                    <div className="flex items-start gap-3 p-4 bg-slate-800/50 rounded-lg">
                      <Terminal className="w-5 h-5 text-blue-500 mt-1" />
                      <div>
                        <h3 className="font-semibold text-slate-200 mb-1">Interactive</h3>
                        <p className="text-sm text-slate-400">
                          Real-time REPL with instant feedback and event streaming
                        </p>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </section>

            <Separator className="bg-slate-800" />

            {/* Getting Started */}
            <section id="getting-started">
              <h2 className="text-3xl font-bold text-slate-100 mb-4 flex items-center gap-2">
                <Play className="w-8 h-8 text-blue-500" />
                Getting Started
              </h2>
              <Card className="bg-slate-900 border-slate-800">
                <CardContent className="pt-6 space-y-6 text-slate-300">
                  <div>
                    <h3 className="text-xl font-semibold text-slate-200 mb-3">Quick Start</h3>
                    <ol className="space-y-3 list-decimal list-inside">
                      <li className="leading-relaxed">
                        <strong className="text-slate-200">Write your code</strong> in the editor on
                        the left side of the screen
                      </li>
                      <li className="leading-relaxed">
                        <strong className="text-slate-200">Click "Run Code"</strong> or press{' '}
                        <kbd className="px-2 py-1 bg-slate-800 rounded font-mono text-sm">
                          ⌘/Ctrl + Enter
                        </kbd>{' '}
                        to execute
                      </li>
                      <li className="leading-relaxed">
                        <strong className="text-slate-200">View results</strong> in the console
                        output panel on the right
                      </li>
                      <li className="leading-relaxed">
                        <strong className="text-slate-200">Try examples</strong> from the dropdown
                        menu to explore different features
                      </li>
                    </ol>
                  </div>

                  <div className="bg-slate-800/50 p-4 rounded-lg">
                    <h4 className="text-sm font-semibold text-slate-200 mb-2 flex items-center gap-2">
                      <Code2 className="w-4 h-4 text-blue-500" />
                      Your First Program
                    </h4>
                    <pre className="bg-slate-950 p-4 rounded-md overflow-x-auto">
                      <code className="text-sm font-mono text-slate-300">
                        {`// Simple hello world
console.log("Hello, VM System!");

// Variables and operations
const x = 10;
const y = 20;
const result = x + y;

console.log("Result:", result);

// Return a value
result;`}
                      </code>
                    </pre>
                  </div>

                  <div>
                    <h3 className="text-xl font-semibold text-slate-200 mb-3">
                      Keyboard Shortcuts
                    </h3>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                      <div className="flex items-center justify-between p-3 bg-slate-800/50 rounded-lg">
                        <span className="text-sm">Execute code</span>
                        <kbd className="px-2 py-1 bg-slate-950 rounded font-mono text-sm">
                          ⌘/Ctrl + Enter
                        </kbd>
                      </div>
                      <div className="flex items-center justify-between p-3 bg-slate-800/50 rounded-lg">
                        <span className="text-sm">Clear console</span>
                        <span className="text-sm text-slate-500">Click "Clear" button</span>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </section>

            <Separator className="bg-slate-800" />

            {/* Features */}
            <section id="features">
              <h2 className="text-3xl font-bold text-slate-100 mb-4 flex items-center gap-2">
                <Layers className="w-8 h-8 text-blue-500" />
                Features
              </h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Card className="bg-slate-900 border-slate-800">
                  <CardHeader>
                    <CardTitle className="text-slate-200">Console API</CardTitle>
                    <CardDescription className="text-slate-400">
                      Full console.log support with output capture
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="text-slate-300">
                    <pre className="bg-slate-950 p-3 rounded-md overflow-x-auto">
                      <code className="text-sm font-mono">
                        {`console.log("Hello");
console.log("Multiple", "args");
console.log({ key: "value" });`}
                      </code>
                    </pre>
                  </CardContent>
                </Card>

                <Card className="bg-slate-900 border-slate-800">
                  <CardHeader>
                    <CardTitle className="text-slate-200">Return Values</CardTitle>
                    <CardDescription className="text-slate-400">
                      Last expression is automatically returned
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="text-slate-300">
                    <pre className="bg-slate-950 p-3 rounded-md overflow-x-auto">
                      <code className="text-sm font-mono">
                        {`const result = 1 + 2;
// Returns: 3
result;`}
                      </code>
                    </pre>
                  </CardContent>
                </Card>

                <Card className="bg-slate-900 border-slate-800">
                  <CardHeader>
                    <CardTitle className="text-slate-200">Error Handling</CardTitle>
                    <CardDescription className="text-slate-400">
                      Try-catch blocks and exception tracking
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="text-slate-300">
                    <pre className="bg-slate-950 p-3 rounded-md overflow-x-auto">
                      <code className="text-sm font-mono">
                        {`try {
  throw new Error("Oops!");
} catch (e) {
  console.log(e.message);
}`}
                      </code>
                    </pre>
                  </CardContent>
                </Card>

                <Card className="bg-slate-900 border-slate-800">
                  <CardHeader>
                    <CardTitle className="text-slate-200">Built-in Objects</CardTitle>
                    <CardDescription className="text-slate-400">
                      Access to standard JavaScript globals
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="text-slate-300">
                    <pre className="bg-slate-950 p-3 rounded-md overflow-x-auto">
                      <code className="text-sm font-mono">
                        {`Math.sqrt(16);
Date.now();
JSON.stringify({ a: 1 });`}
                      </code>
                    </pre>
                  </CardContent>
                </Card>
              </div>
            </section>

            <Separator className="bg-slate-800" />

            {/* Code Examples */}
            <section id="examples">
              <h2 className="text-3xl font-bold text-slate-100 mb-4 flex items-center gap-2">
                <FileCode className="w-8 h-8 text-blue-500" />
                Code Examples
              </h2>
              <div className="space-y-4">
                <Card className="bg-slate-900 border-slate-800">
                  <CardHeader>
                    <CardTitle className="text-slate-200">Array Operations</CardTitle>
                    <CardDescription className="text-slate-400">
                      Working with arrays using map, filter, and reduce
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <pre className="bg-slate-950 p-4 rounded-md overflow-x-auto">
                      <code className="text-sm font-mono text-slate-300">
                        {`const numbers = [1, 2, 3, 4, 5];

// Double each number
const doubled = numbers.map(n => n * 2);
console.log("Doubled:", doubled);

// Sum all numbers
const sum = numbers.reduce((a, b) => a + b, 0);
console.log("Sum:", sum);

// Filter numbers greater than 2
const filtered = numbers.filter(n => n > 2);
console.log("Filtered:", filtered);

{ doubled, sum, filtered };`}
                      </code>
                    </pre>
                  </CardContent>
                </Card>

                <Card className="bg-slate-900 border-slate-800">
                  <CardHeader>
                    <CardTitle className="text-slate-200">Object Manipulation</CardTitle>
                    <CardDescription className="text-slate-400">
                      Creating and transforming objects
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <pre className="bg-slate-950 p-4 rounded-md overflow-x-auto">
                      <code className="text-sm font-mono text-slate-300">
                        {`const person = {
  name: "Alice",
  age: 30,
  city: "San Francisco"
};

// Spread operator for updates
const updated = {
  ...person,
  age: 31,
  occupation: "Engineer"
};

console.log("Original:", person);
console.log("Updated:", updated);

updated;`}
                      </code>
                    </pre>
                  </CardContent>
                </Card>

                <Card className="bg-slate-900 border-slate-800">
                  <CardHeader>
                    <CardTitle className="text-slate-200">Functions</CardTitle>
                    <CardDescription className="text-slate-400">
                      Defining and using functions
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <pre className="bg-slate-950 p-4 rounded-md overflow-x-auto">
                      <code className="text-sm font-mono text-slate-300">
                        {`// Regular function
function fibonacci(n) {
  if (n <= 1) return n;
  return fibonacci(n - 1) + fibonacci(n - 2);
}

// Arrow function
const factorial = (n) => {
  if (n <= 1) return 1;
  return n * factorial(n - 1);
};

const fib10 = fibonacci(10);
const fact5 = factorial(5);

console.log("Fibonacci(10):", fib10);
console.log("Factorial(5):", fact5);

{ fib10, fact5 };`}
                      </code>
                    </pre>
                  </CardContent>
                </Card>
              </div>
            </section>

            <Separator className="bg-slate-800" />

            {/* VM Architecture */}
            <section id="vm-architecture">
              <h2 className="text-3xl font-bold text-slate-100 mb-4 flex items-center gap-2">
                <Cpu className="w-8 h-8 text-blue-500" />
                VM Architecture
              </h2>
              <Card className="bg-slate-900 border-slate-800">
                <CardContent className="pt-6 space-y-6 text-slate-300">
                  <div>
                    <h3 className="text-xl font-semibold text-slate-200 mb-3">Components</h3>
                    <div className="space-y-4">
                      <div className="p-4 bg-slate-800/50 rounded-lg">
                        <h4 className="font-semibold text-slate-200 mb-2">VM Profile</h4>
                        <p className="text-sm leading-relaxed">
                          Defines the configuration template for a VM instance, including engine
                          type, resource limits, module resolution rules, and runtime settings. Each
                          profile can have multiple capabilities and startup files.
                        </p>
                      </div>
                      <div className="p-4 bg-slate-800/50 rounded-lg">
                        <h4 className="font-semibold text-slate-200 mb-2">VM Session</h4>
                        <p className="text-sm leading-relaxed">
                          Represents an active runtime instance created from a VM profile. Sessions
                          maintain their own execution context and can run multiple code snippets
                          sequentially while preserving state.
                        </p>
                      </div>
                      <div className="p-4 bg-slate-800/50 rounded-lg">
                        <h4 className="font-semibold text-slate-200 mb-2">Execution</h4>
                        <p className="text-sm leading-relaxed">
                          Each code execution within a session is tracked as a separate execution
                          record. Executions capture input, output, timing, and all events that
                          occur during runtime.
                        </p>
                      </div>
                      <div className="p-4 bg-slate-800/50 rounded-lg">
                        <h4 className="font-semibold text-slate-200 mb-2">Events</h4>
                        <p className="text-sm leading-relaxed">
                          Events are timestamped records of everything that happens during execution:
                          console output, return values, exceptions, and more. Events are stored in
                          sequence for replay and debugging.
                        </p>
                      </div>
                    </div>
                  </div>

                  <div>
                    <h3 className="text-xl font-semibold text-slate-200 mb-3">Resource Limits</h3>
                    <div className="grid grid-cols-2 gap-3">
                      <div className="p-3 bg-slate-800/50 rounded-lg">
                        <div className="text-sm font-semibold text-slate-200 mb-1">CPU Time</div>
                        <div className="text-xs text-slate-400">2000ms per execution</div>
                      </div>
                      <div className="p-3 bg-slate-800/50 rounded-lg">
                        <div className="text-sm font-semibold text-slate-200 mb-1">Wall Time</div>
                        <div className="text-xs text-slate-400">5000ms per execution</div>
                      </div>
                      <div className="p-3 bg-slate-800/50 rounded-lg">
                        <div className="text-sm font-semibold text-slate-200 mb-1">Memory</div>
                        <div className="text-xs text-slate-400">128MB per session</div>
                      </div>
                      <div className="p-3 bg-slate-800/50 rounded-lg">
                        <div className="text-sm font-semibold text-slate-200 mb-1">Output</div>
                        <div className="text-xs text-slate-400">256KB per execution</div>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </section>

            <Separator className="bg-slate-800" />

            {/* API Reference */}
            <section id="api-reference">
              <h2 className="text-3xl font-bold text-slate-100 mb-4 flex items-center gap-2">
                <Code2 className="w-8 h-8 text-blue-500" />
                API Reference
              </h2>
              <Card className="bg-slate-900 border-slate-800">
                <CardContent className="pt-6 space-y-6 text-slate-300">
                  <div>
                    <h3 className="text-xl font-semibold text-slate-200 mb-3">
                      Available Globals
                    </h3>
                    <div className="space-y-3">
                      <div className="p-4 bg-slate-800/50 rounded-lg">
                        <code className="text-blue-400 font-mono">console.log(...args)</code>
                        <p className="text-sm mt-2">
                          Outputs one or more values to the console. Supports strings, numbers,
                          objects, and arrays.
                        </p>
                      </div>
                      <div className="p-4 bg-slate-800/50 rounded-lg">
                        <code className="text-blue-400 font-mono">Math</code>
                        <p className="text-sm mt-2">
                          Standard JavaScript Math object with methods like sqrt, pow, random, etc.
                        </p>
                      </div>
                      <div className="p-4 bg-slate-800/50 rounded-lg">
                        <code className="text-blue-400 font-mono">Date</code>
                        <p className="text-sm mt-2">
                          Date constructor and methods for working with dates and times.
                        </p>
                      </div>
                      <div className="p-4 bg-slate-800/50 rounded-lg">
                        <code className="text-blue-400 font-mono">JSON</code>
                        <p className="text-sm mt-2">
                          JSON.parse() and JSON.stringify() for working with JSON data.
                        </p>
                      </div>
                      <div className="p-4 bg-slate-800/50 rounded-lg">
                        <code className="text-blue-400 font-mono">Array, Object, String</code>
                        <p className="text-sm mt-2">
                          Standard JavaScript constructors with all prototype methods.
                        </p>
                      </div>
                    </div>
                  </div>

                  <div>
                    <h3 className="text-xl font-semibold text-slate-200 mb-3">
                      Execution Model
                    </h3>
                    <div className="p-4 bg-slate-800/50 rounded-lg space-y-3">
                      <p className="text-sm leading-relaxed">
                        Code is executed in a sandboxed environment with the following behavior:
                      </p>
                      <ul className="space-y-2 text-sm list-disc list-inside">
                        <li>The last expression in your code is automatically returned</li>
                        <li>All console.log calls are captured and displayed in the output panel</li>
                        <li>Exceptions are caught and displayed with stack traces</li>
                        <li>Each execution is independent but shares the same session context</li>
                        <li>Variables from previous executions are not preserved</li>
                      </ul>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </section>

            <Separator className="bg-slate-800" />

            {/* Limitations */}
            <section id="limitations">
              <h2 className="text-3xl font-bold text-slate-100 mb-4">Limitations</h2>
              <Card className="bg-slate-900 border-slate-800">
                <CardContent className="pt-6 space-y-4 text-slate-300">
                  <p className="leading-relaxed">
                    This VM system has some intentional limitations for security and performance:
                  </p>
                  <ul className="space-y-2 list-disc list-inside">
                    <li>No async/await or Promise support (synchronous execution only)</li>
                    <li>No DOM or browser APIs (window, document, etc.)</li>
                    <li>No file system access</li>
                    <li>No network requests (fetch, XMLHttpRequest)</li>
                    <li>No module imports (import/require)</li>
                    <li>No setTimeout/setInterval</li>
                    <li>Limited to standard JavaScript built-ins</li>
                  </ul>
                  <div className="mt-6 p-4 bg-blue-950/30 border border-blue-800/50 rounded-lg">
                    <p className="text-sm text-blue-300">
                      <strong>Note:</strong> These limitations ensure a safe, predictable execution
                      environment. For more advanced features, consider upgrading to a full-stack VM
                      deployment with custom capabilities.
                    </p>
                  </div>
                </CardContent>
              </Card>
            </section>

            {/* Footer CTA */}
            <div className="mt-12 p-6 bg-gradient-to-r from-blue-950/50 to-slate-900 border border-blue-800/50 rounded-lg">
              <h3 className="text-xl font-semibold text-slate-100 mb-2">Ready to start coding?</h3>
              <p className="text-slate-300 mb-4">
                Head back to the editor and try out the examples or write your own JavaScript code.
              </p>
              <Link href="/">
                <Button className="bg-blue-600 hover:bg-blue-700 text-white">
                  <Terminal className="w-4 h-4 mr-2" />
                  Go to Editor
                </Button>
              </Link>
            </div>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t border-slate-800 bg-slate-900/50 backdrop-blur mt-12">
        <div className="container py-6">
          <div className="flex items-center justify-between text-sm text-slate-500">
            <div>VM System Documentation v1.0</div>
            <div>Built with goja and React</div>
          </div>
        </div>
      </footer>
    </div>
  );
}
