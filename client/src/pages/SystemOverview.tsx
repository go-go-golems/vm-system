import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import {
  ArrowRight,
  BookOpen,
  Box,
  Database,
  FileCode,
  GitBranch,
  Home,
  Layers,
  Network,
  Server,
  Settings,
  Terminal,
  Zap,
} from 'lucide-react';
import { Link } from 'wouter';

export default function SystemOverview() {
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
                <h1 className="text-xl font-bold text-slate-100">System Overview</h1>
                <p className="text-sm text-slate-400">Architecture and implementation details</p>
              </div>
            </div>

            <div className="flex items-center gap-2">
              <Link href="/docs">
                <Button
                  variant="outline"
                  className="bg-slate-900 border-slate-700 text-slate-300"
                >
                  <BookOpen className="w-4 h-4 mr-2" />
                  User Guide
                </Button>
              </Link>
              <Link href="/">
                <Button variant="outline" className="bg-slate-900 border-slate-700 text-slate-300">
                  <Home className="w-4 h-4 mr-2" />
                  Back to Editor
                </Button>
              </Link>
            </div>
          </div>
        </div>
      </header>

      {/* Main content */}
      <main className="container py-8 max-w-5xl">
        <div className="space-y-12">
          {/* Introduction */}
          <section>
            <h2 className="text-3xl font-bold text-slate-100 mb-6">
              What is the VM System?
            </h2>
            <Card className="bg-slate-900 border-slate-800">
              <CardContent className="pt-6 space-y-4 text-slate-300 leading-relaxed">
                <p>
                  The VM System is a comprehensive JavaScript execution environment built on the{' '}
                  <strong className="text-blue-400">goja</strong> runtime (a pure Go implementation
                  of ECMAScript 5.1). It provides a sandboxed, configurable environment for running
                  JavaScript code with fine-grained control over capabilities, resource limits, and
                  execution tracking.
                </p>
                <p>
                  This system is designed to integrate with a{' '}
                  <strong className="text-blue-400">dual-storage filesystem</strong> (Git + SQLite)
                  that provides version control, workspace management, and session isolation. Together,
                  these subsystems enable a complete development environment where code can be written,
                  executed, versioned, and published.
                </p>
                <p>
                  The architecture follows a clear separation of concerns: the VM subsystem handles
                  code execution and runtime management, while the filesystem subsystem manages code
                  storage, versioning, and workspace isolation. This web interface provides an
                  interactive playground for testing the VM capabilities.
                </p>
              </CardContent>
            </Card>
          </section>

          <Separator className="bg-slate-800" />

          {/* Architecture Overview */}
          <section>
            <h2 className="text-3xl font-bold text-slate-100 mb-6 flex items-center gap-2">
              <Layers className="w-8 h-8 text-blue-500" />
              System Architecture
            </h2>

            <div className="space-y-6">
              {/* High-level diagram */}
              <Card className="bg-slate-900 border-slate-800">
                <CardContent className="pt-6">
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div className="p-6 bg-slate-800/50 rounded-lg border-2 border-blue-500/30">
                      <div className="flex items-center gap-2 mb-3">
                        <Terminal className="w-5 h-5 text-blue-500" />
                        <h3 className="font-semibold text-slate-200">VM Subsystem</h3>
                      </div>
                      <p className="text-sm text-slate-400">
                        Manages JavaScript execution, profiles, sessions, and event logging
                      </p>
                    </div>

                    <div className="flex items-center justify-center">
                      <ArrowRight className="w-6 h-6 text-slate-600" />
                    </div>

                    <div className="p-6 bg-slate-800/50 rounded-lg border-2 border-emerald-500/30">
                      <div className="flex items-center gap-2 mb-3">
                        <Database className="w-5 h-5 text-emerald-500" />
                        <h3 className="font-semibold text-slate-200">Dual-Storage FS</h3>
                      </div>
                      <p className="text-sm text-slate-400">
                        Git + SQLite for version control, workspaces, and session isolation
                      </p>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Component breakdown */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Card className="bg-slate-900 border-slate-800">
                  <CardContent className="pt-6 space-y-3">
                    <div className="flex items-center gap-2 mb-2">
                      <Box className="w-5 h-5 text-blue-500" />
                      <h3 className="font-semibold text-slate-200">VM Config Store</h3>
                    </div>
                    <p className="text-sm text-slate-400 leading-relaxed">
                      SQLite database storing VM profiles, capabilities, startup files, and resource
                      limits. Each profile defines a template for creating VM sessions with specific
                      configurations.
                    </p>
                  </CardContent>
                </Card>

                <Card className="bg-slate-900 border-slate-800">
                  <CardContent className="pt-6 space-y-3">
                    <div className="flex items-center gap-2 mb-2">
                      <Server className="w-5 h-5 text-blue-500" />
                      <h3 className="font-semibold text-slate-200">Session Manager</h3>
                    </div>
                    <p className="text-sm text-slate-400 leading-relaxed">
                      Creates and manages VM sessions, tracks runtime state, and routes execution
                      requests. Each session is tied to a workspace and maintains its own goja
                      runtime instance.
                    </p>
                  </CardContent>
                </Card>

                <Card className="bg-slate-900 border-slate-800">
                  <CardContent className="pt-6 space-y-3">
                    <div className="flex items-center gap-2 mb-2">
                      <Zap className="w-5 h-5 text-blue-500" />
                      <h3 className="font-semibold text-slate-200">Execution Runner</h3>
                    </div>
                    <p className="text-sm text-slate-400 leading-relaxed">
                      Handles startup scripts, REPL evaluation, and file execution. Captures all
                      events (console output, return values, exceptions) and enforces resource
                      limits.
                    </p>
                  </CardContent>
                </Card>

                <Card className="bg-slate-900 border-slate-800">
                  <CardContent className="pt-6 space-y-3">
                    <div className="flex items-center gap-2 mb-2">
                      <Network className="w-5 h-5 text-blue-500" />
                      <h3 className="font-semibold text-slate-200">Module Resolver</h3>
                    </div>
                    <p className="text-sm text-slate-400 leading-relaxed">
                      Implements import resolution for workspace files and host-provided modules.
                      Enforces capability allowlists and ensures secure module exposure.
                    </p>
                  </CardContent>
                </Card>
              </div>
            </div>
          </section>

          <Separator className="bg-slate-800" />

          {/* VM Profiles */}
          <section>
            <h2 className="text-3xl font-bold text-slate-100 mb-6 flex items-center gap-2">
              <Settings className="w-8 h-8 text-blue-500" />
              VM Profiles
            </h2>

            <Card className="bg-slate-900 border-slate-800">
              <CardContent className="pt-6 space-y-6 text-slate-300">
                <p className="leading-relaxed">
                  A <strong className="text-blue-400">VM Profile</strong> is a configuration template
                  that defines how a VM session should behave. Think of it as a blueprint that
                  specifies the engine type, available capabilities, resource constraints, and
                  initialization scripts.
                </p>

                <div className="space-y-4">
                  <div className="p-4 bg-slate-800/50 rounded-lg">
                    <h4 className="font-semibold text-slate-200 mb-2">Engine Configuration</h4>
                    <p className="text-sm text-slate-400 leading-relaxed">
                      Specifies the JavaScript runtime engine (goja, QuickJS, Node.js-like, or
                      custom). Each engine has different performance characteristics and feature
                      support. This implementation uses <code className="text-blue-400">goja</code>,
                      a pure Go implementation of ECMAScript 5.1.
                    </p>
                  </div>

                  <div className="p-4 bg-slate-800/50 rounded-lg">
                    <h4 className="font-semibold text-slate-200 mb-2">Resource Limits</h4>
                    <p className="text-sm text-slate-400 leading-relaxed mb-3">
                      Each profile defines strict resource constraints to prevent runaway code:
                    </p>
                    <ul className="text-sm text-slate-400 space-y-1 list-disc list-inside">
                      <li>
                        <strong>CPU time</strong>: Maximum CPU milliseconds per execution (default:
                        2000ms)
                      </li>
                      <li>
                        <strong>Wall time</strong>: Maximum real-world time per execution (default:
                        5000ms)
                      </li>
                      <li>
                        <strong>Memory</strong>: Maximum memory usage per session (default: 128MB)
                      </li>
                      <li>
                        <strong>Output size</strong>: Maximum console output per execution (default:
                        256KB)
                      </li>
                      <li>
                        <strong>Event count</strong>: Maximum events per execution (default: 50,000)
                      </li>
                    </ul>
                  </div>

                  <div className="p-4 bg-slate-800/50 rounded-lg">
                    <h4 className="font-semibold text-slate-200 mb-2">Module Resolution</h4>
                    <p className="text-sm text-slate-400 leading-relaxed">
                      Defines how imports are resolved, including root directories, file extensions,
                      and whether absolute repository imports are allowed. This ensures code can only
                      access files within its designated workspace.
                    </p>
                  </div>

                  <div className="p-4 bg-slate-800/50 rounded-lg">
                    <h4 className="font-semibold text-slate-200 mb-2">Runtime Settings</h4>
                    <p className="text-sm text-slate-400 leading-relaxed">
                      Controls runtime behavior such as ESM support, strict mode enforcement, and
                      console API availability. These settings affect how JavaScript code is parsed
                      and executed.
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </section>

          <Separator className="bg-slate-800" />

          {/* Capabilities */}
          <section>
            <h2 className="text-3xl font-bold text-slate-100 mb-6 flex items-center gap-2">
              <Network className="w-8 h-8 text-blue-500" />
              Capabilities & Module Exposure
            </h2>

            <Card className="bg-slate-900 border-slate-800">
              <CardContent className="pt-6 space-y-6 text-slate-300">
                <p className="leading-relaxed">
                  The VM system follows a <strong className="text-blue-400">deny-by-default</strong>{' '}
                  security model. No modules or globals are available unless explicitly enabled
                  through capabilities. This ensures that code can only access the APIs and resources
                  it's been granted permission to use.
                </p>

                <div className="space-y-4">
                  <div className="p-4 bg-slate-800/50 rounded-lg">
                    <h4 className="font-semibold text-slate-200 mb-2">Module Capabilities</h4>
                    <p className="text-sm text-slate-400 leading-relaxed mb-3">
                      Control which host-provided modules can be imported. Examples include:
                    </p>
                    <ul className="text-sm text-slate-400 space-y-1 list-disc list-inside">
                      <li>
                        <code className="text-blue-400">fetch</code> - HTTP client for making
                        requests
                      </li>
                      <li>
                        <code className="text-blue-400">@host/kv</code> - Key-value storage API
                      </li>
                      <li>
                        <code className="text-blue-400">@host/db</code> - Database access
                      </li>
                      <li>
                        <code className="text-blue-400">fs</code> - Filesystem operations (with path
                        restrictions)
                      </li>
                    </ul>
                    <p className="text-sm text-slate-400 leading-relaxed mt-3">
                      Each capability can have its own configuration (e.g., allowed hosts for fetch,
                      path restrictions for fs).
                    </p>
                  </div>

                  <div className="p-4 bg-slate-800/50 rounded-lg">
                    <h4 className="font-semibold text-slate-200 mb-2">Global Capabilities</h4>
                    <p className="text-sm text-slate-400 leading-relaxed">
                      Control which global objects and functions are injected into the runtime. This
                      includes <code className="text-blue-400">console</code>,{' '}
                      <code className="text-blue-400">setTimeout</code>,{' '}
                      <code className="text-blue-400">crypto</code>, and other globals. Each must be
                      explicitly enabled.
                    </p>
                  </div>

                  <div className="p-4 bg-slate-800/50 rounded-lg">
                    <h4 className="font-semibold text-slate-200 mb-2">Import Resolution Flow</h4>
                    <div className="text-sm text-slate-400 space-y-2 mt-3">
                      <p className="leading-relaxed">
                        When code attempts to import a module, the resolver follows this logic:
                      </p>
                      <ol className="list-decimal list-inside space-y-1 ml-2">
                        <li>Check if the specifier is a relative path (./ or ../)</li>
                        <li>If relative, resolve within the workspace directory</li>
                        <li>If bare specifier, look up in capability allowlist</li>
                        <li>If not found or disabled, throw "module not allowed" error</li>
                        <li>If found and enabled, load the host-provided module with its config</li>
                      </ol>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </section>

          <Separator className="bg-slate-800" />

          {/* Sessions & Executions */}
          <section>
            <h2 className="text-3xl font-bold text-slate-100 mb-6 flex items-center gap-2">
              <Terminal className="w-8 h-8 text-blue-500" />
              Sessions & Executions
            </h2>

            <Card className="bg-slate-900 border-slate-800">
              <CardContent className="pt-6 space-y-6 text-slate-300">
                <div>
                  <h3 className="text-xl font-semibold text-slate-200 mb-3">VM Sessions</h3>
                  <p className="leading-relaxed mb-4">
                    A <strong className="text-blue-400">VM Session</strong> is a runtime instance
                    created from a VM profile. Each session has its own goja runtime, execution
                    context, and is tied to a specific workspace. Sessions maintain state across
                    multiple executions, allowing you to build up context over time.
                  </p>
                  <div className="p-4 bg-slate-800/50 rounded-lg">
                    <h4 className="font-semibold text-slate-200 mb-2 text-sm">Session Lifecycle</h4>
                    <div className="flex items-center gap-2 text-sm">
                      <span className="px-2 py-1 bg-blue-950 border border-blue-800 text-blue-400 rounded font-mono">
                        starting
                      </span>
                      <ArrowRight className="w-4 h-4 text-slate-600" />
                      <span className="px-2 py-1 bg-emerald-950 border border-emerald-800 text-emerald-400 rounded font-mono">
                        ready
                      </span>
                      <ArrowRight className="w-4 h-4 text-slate-600" />
                      <span className="px-2 py-1 bg-slate-800 border border-slate-700 text-slate-400 rounded font-mono">
                        crashed/closed
                      </span>
                    </div>
                  </div>
                </div>

                <div>
                  <h3 className="text-xl font-semibold text-slate-200 mb-3">Execution Types</h3>
                  <div className="space-y-3">
                    <div className="p-4 bg-slate-800/50 rounded-lg">
                      <h4 className="font-semibold text-slate-200 mb-2 flex items-center gap-2">
                        <FileCode className="w-4 h-4 text-blue-500" />
                        Startup Execution
                      </h4>
                      <p className="text-sm text-slate-400 leading-relaxed">
                        Runs automatically when a session is created. Startup files are loaded in
                        order from the VM profile's startup file list. If any startup file fails, the
                        session enters a crashed state and cannot be used.
                      </p>
                    </div>

                    <div className="p-4 bg-slate-800/50 rounded-lg">
                      <h4 className="font-semibold text-slate-200 mb-2 flex items-center gap-2">
                        <Terminal className="w-4 h-4 text-blue-500" />
                        REPL Execution
                      </h4>
                      <p className="text-sm text-slate-400 leading-relaxed">
                        Evaluates a code snippet in the session's context. This is what you're using
                        in the web interface. The last expression is automatically returned, and all
                        console output is captured as events.
                      </p>
                    </div>

                    <div className="p-4 bg-slate-800/50 rounded-lg">
                      <h4 className="font-semibold text-slate-200 mb-2 flex items-center gap-2">
                        <FileCode className="w-4 h-4 text-blue-500" />
                        Run-File Execution
                      </h4>
                      <p className="text-sm text-slate-400 leading-relaxed">
                        Executes a file from the workspace as an entry point. The file path is
                        resolved within the workspace, and optional arguments and environment
                        variables can be passed.
                      </p>
                    </div>
                  </div>
                </div>

                <div>
                  <h3 className="text-xl font-semibold text-slate-200 mb-3">Event Capture</h3>
                  <p className="leading-relaxed mb-4">
                    Every execution generates a stream of timestamped events that capture everything
                    that happens during runtime. Events are stored with monotonically increasing
                    sequence numbers, making them replayable and debuggable.
                  </p>
                  <div className="grid grid-cols-2 gap-3">
                    <div className="p-3 bg-slate-800/50 rounded-lg">
                      <code className="text-blue-400 text-sm">console</code>
                      <p className="text-xs text-slate-500 mt-1">Console output events</p>
                    </div>
                    <div className="p-3 bg-slate-800/50 rounded-lg">
                      <code className="text-blue-400 text-sm">value</code>
                      <p className="text-xs text-slate-500 mt-1">Return value events</p>
                    </div>
                    <div className="p-3 bg-slate-800/50 rounded-lg">
                      <code className="text-blue-400 text-sm">exception</code>
                      <p className="text-xs text-slate-500 mt-1">Error and exception events</p>
                    </div>
                    <div className="p-3 bg-slate-800/50 rounded-lg">
                      <code className="text-blue-400 text-sm">system</code>
                      <p className="text-xs text-slate-500 mt-1">Lifecycle and debug events</p>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </section>

          <Separator className="bg-slate-800" />

          {/* Dual-Storage Filesystem */}
          <section>
            <h2 className="text-3xl font-bold text-slate-100 mb-6 flex items-center gap-2">
              <GitBranch className="w-8 h-8 text-emerald-500" />
              Dual-Storage Filesystem
            </h2>

            <Card className="bg-slate-900 border-slate-800">
              <CardContent className="pt-6 space-y-6 text-slate-300">
                <p className="leading-relaxed">
                  The VM system integrates with a{' '}
                  <strong className="text-emerald-400">dual-storage filesystem</strong> that combines
                  Git and SQLite to provide version control, workspace management, and session
                  isolation. This architecture ensures that code is both version-controlled and easily
                  queryable.
                </p>

                <div className="space-y-4">
                  <div className="p-4 bg-slate-800/50 rounded-lg">
                    <h4 className="font-semibold text-slate-200 mb-2">Git Storage</h4>
                    <p className="text-sm text-slate-400 leading-relaxed">
                      A standard Git repository stores the authoritative file tree with one linear
                      history (no branches). Sessions use Git worktrees (detached HEAD) to provide
                      isolated working directories. This enables interoperability with standard Git
                      tools and provides a familiar version control model.
                    </p>
                  </div>

                  <div className="p-4 bg-slate-800/50 rounded-lg">
                    <h4 className="font-semibold text-slate-200 mb-2">SQLite Storage</h4>
                    <p className="text-sm text-slate-400 leading-relaxed">
                      A SQLite database mirrors the Git object graph (blobs, trees, commits) using
                      Git SHA-1 commit IDs as canonical identifiers. It also stores workspace overlays
                      (per-session file edits) and provides fast querying for provenance, introspection,
                      and offline packaging.
                    </p>
                  </div>

                  <div className="p-4 bg-slate-800/50 rounded-lg">
                    <h4 className="font-semibold text-slate-200 mb-2">Workspace Isolation</h4>
                    <p className="text-sm text-slate-400 leading-relaxed">
                      Each session has its own workspace tied to a base commit. File edits are tracked
                      as overlays in SQLite without modifying the Git repository. When ready, changes
                      can be published to create a new commit on the main line (fast-forward only, no
                      merges).
                    </p>
                  </div>

                  <div className="p-4 bg-slate-800/50 rounded-lg">
                    <h4 className="font-semibold text-slate-200 mb-2">Consistency Guarantees</h4>
                    <p className="text-sm text-slate-400 leading-relaxed">
                      Every state-changing operation writes to both Git and SQLite. If one write
                      succeeds and the other fails, the system marks the repository as "needs
                      reconciliation" and fails the request. This ensures both stores remain
                      bitwise-consistent.
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </section>

          <Separator className="bg-slate-800" />

          {/* Implementation Notes */}
          <section>
            <h2 className="text-3xl font-bold text-slate-100 mb-6">Implementation Notes</h2>

            <Card className="bg-slate-900 border-slate-800">
              <CardContent className="pt-6 space-y-6 text-slate-300">
                <div>
                  <h3 className="text-xl font-semibold text-slate-200 mb-3">This Web Interface</h3>
                  <p className="leading-relaxed mb-4">
                    This web UI is a <strong className="text-blue-400">demonstration interface</strong>{' '}
                    that simulates the VM system's behavior in the browser. It implements a mock
                    backend using browser-based JavaScript execution to showcase the user experience
                    and API design.
                  </p>
                  <div className="p-4 bg-blue-950/30 border border-blue-800/50 rounded-lg">
                    <p className="text-sm text-blue-300 leading-relaxed">
                      <strong>Note:</strong> The actual VM system is implemented in Go with goja and
                      integrates with a real dual-storage filesystem. This web interface provides an
                      interactive way to explore the concepts and test JavaScript code execution.
                    </p>
                  </div>
                </div>

                <div>
                  <h3 className="text-xl font-semibold text-slate-200 mb-3">Go Implementation</h3>
                  <p className="leading-relaxed mb-4">
                    The production VM system is built in Go and uses the following key libraries:
                  </p>
                  <ul className="space-y-2 text-sm text-slate-400">
                    <li className="flex items-start gap-2">
                      <span className="text-blue-500 mt-1">•</span>
                      <div>
                        <strong className="text-slate-300">goja</strong> - Pure Go implementation of
                        ECMAScript 5.1 for JavaScript execution
                      </div>
                    </li>
                    <li className="flex items-start gap-2">
                      <span className="text-blue-500 mt-1">•</span>
                      <div>
                        <strong className="text-slate-300">libgit2/git2go</strong> - Git operations
                        and repository management
                      </div>
                    </li>
                    <li className="flex items-start gap-2">
                      <span className="text-blue-500 mt-1">•</span>
                      <div>
                        <strong className="text-slate-300">mattn/go-sqlite3</strong> - SQLite
                        database driver with CGO
                      </div>
                    </li>
                    <li className="flex items-start gap-2">
                      <span className="text-blue-500 mt-1">•</span>
                      <div>
                        <strong className="text-slate-300">spf13/cobra</strong> - CLI framework for
                        command-line interface
                      </div>
                    </li>
                  </ul>
                </div>
              </CardContent>
            </Card>
          </section>

          {/* Footer CTA */}
          <div className="mt-12 p-6 bg-gradient-to-r from-blue-950/50 to-slate-900 border border-blue-800/50 rounded-lg">
            <h3 className="text-xl font-semibold text-slate-100 mb-2">Ready to explore?</h3>
            <p className="text-slate-300 mb-4">
              Try out the interactive editor or read the user guide to learn more about using the VM
              system.
            </p>
            <div className="flex gap-3">
              <Link href="/">
                <Button className="bg-blue-600 hover:bg-blue-700 text-white">
                  <Terminal className="w-4 h-4 mr-2" />
                  Go to Editor
                </Button>
              </Link>
              <Link href="/docs">
                <Button variant="outline" className="bg-slate-900 border-slate-700 text-slate-300">
                  <BookOpen className="w-4 h-4 mr-2" />
                  User Guide
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
            <div>VM System Architecture Documentation v1.0</div>
            <div>Built with goja, Git, and SQLite</div>
          </div>
        </div>
      </footer>
    </div>
  );
}
