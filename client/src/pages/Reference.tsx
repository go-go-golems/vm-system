import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import { PRESET_EXAMPLES } from '@/lib/types';
import {
  BookOpen,
  ChevronDown,
  ChevronRight,
  Code2,
  FileCode,
  Layers,
  Terminal,
} from 'lucide-react';
import { useState } from 'react';

export default function Reference() {
  return (
    <div className="space-y-6 max-w-3xl">
      <div>
        <h1 className="text-lg font-semibold text-slate-100">Reference</h1>
        <p className="text-sm text-slate-400">Object model, API endpoints, and code examples.</p>
      </div>

      {/* Object model */}
      <section className="space-y-3">
        <h2 className="text-sm font-medium text-slate-200 flex items-center gap-2">
          <Layers className="w-4 h-4 text-blue-500" />
          Object Model
        </h2>
        <Card className="bg-slate-900 border-slate-800">
          <CardContent className="pt-5">
            <pre className="text-xs text-slate-300 font-mono leading-relaxed overflow-x-auto">
{`Template (VM)
│  id, name, engine, exposed_modules[] (native modules), libraries[],
│  settings {limits, resolver, runtime},
│  capabilities[], startup_files[]
│
└─► Session
    │  id, template_id, workspace_id, base_commit_oid,
    │  worktree_path, status, created_at, closed_at
    │
    └─► Execution
        │  id, session_id, kind (repl | run_file | startup),
        │  input, path, status, started_at, ended_at,
        │  result, error
        │
        └─► ExecutionEvent
              execution_id, seq, ts,
              type (console | value | exception | stdout |
                    stderr | system | input_echo),
              payload`}
            </pre>
            <Separator className="bg-slate-800 my-4" />
            <div className="text-xs text-slate-400 space-y-1.5">
              <p><strong className="text-slate-300">Scope rules:</strong></p>
              <ul className="list-disc list-inside space-y-1 text-slate-500">
                <li>Template changes affect <strong className="text-slate-400">new sessions only</strong>. Running sessions keep their snapshot.</li>
                <li>Sessions own their runtime state. Closing the daemon destroys in-memory runtimes.</li>
                <li>Executions belong to exactly one session. Events belong to exactly one execution.</li>
              </ul>
            </div>
          </CardContent>
        </Card>
      </section>

      {/* API endpoints */}
      <section className="space-y-3">
        <h2 className="text-sm font-medium text-slate-200 flex items-center gap-2">
          <Code2 className="w-4 h-4 text-blue-500" />
          API Endpoints
        </h2>
        <Card className="bg-slate-900 border-slate-800">
          <CardContent className="pt-5">
            <div className="overflow-x-auto">
              <table className="w-full text-xs">
                <thead>
                  <tr className="border-b border-slate-800">
                    <th className="text-left py-2 pr-3 text-slate-400 font-medium">Method</th>
                    <th className="text-left py-2 pr-3 text-slate-400 font-medium">Path</th>
                    <th className="text-left py-2 text-slate-400 font-medium">Description</th>
                  </tr>
                </thead>
                <tbody className="text-slate-300">
                  <tr className="border-b border-slate-800/50">
                    <td className="py-1.5 pr-3" colSpan={3}><strong className="text-slate-400">Templates</strong></td>
                  </tr>
                  <ApiRow method="GET" path="/api/v1/templates" desc="List all templates" />
                  <ApiRow method="POST" path="/api/v1/templates" desc="Create a template" />
                  <ApiRow method="GET" path="/api/v1/templates/:id" desc="Get template detail (settings, capabilities, startup files)" />
                  <ApiRow method="POST" path="/api/v1/templates/:id/modules" desc="Add a native module to a template" />
                  <ApiRow method="DELETE" path="/api/v1/templates/:id/modules/:name" desc="Remove a native module" />
                  <ApiRow method="POST" path="/api/v1/templates/:id/libraries" desc="Add a library" />
                  <ApiRow method="DELETE" path="/api/v1/templates/:id/libraries/:name" desc="Remove a library" />

                  <tr className="border-b border-slate-800/50">
                    <td className="py-1.5 pr-3 pt-3" colSpan={3}><strong className="text-slate-400">Sessions</strong></td>
                  </tr>
                  <ApiRow method="GET" path="/api/v1/sessions" desc="List sessions (optional ?status= filter)" />
                  <ApiRow method="POST" path="/api/v1/sessions" desc="Create a session from a template" />
                  <ApiRow method="GET" path="/api/v1/sessions/:id" desc="Get session detail" />
                  <ApiRow method="POST" path="/api/v1/sessions/:id/close" desc="Close a session" />
                  <ApiRow method="DELETE" path="/api/v1/sessions/:id" desc="Delete a session" />

                  <tr className="border-b border-slate-800/50">
                    <td className="py-1.5 pr-3 pt-3" colSpan={3}><strong className="text-slate-400">Executions</strong></td>
                  </tr>
                  <ApiRow method="GET" path="/api/v1/executions" desc="List executions (?session_id=, ?limit=)" />
                  <ApiRow method="POST" path="/api/v1/executions/repl" desc="Execute a REPL snippet" />
                  <ApiRow method="GET" path="/api/v1/executions/:id" desc="Get execution detail" />
                  <ApiRow method="GET" path="/api/v1/executions/:id/events" desc="Get execution events (?after_seq=)" />
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      </section>

      {/* Available globals */}
      <section className="space-y-3">
        <h2 className="text-sm font-medium text-slate-200 flex items-center gap-2">
          <Terminal className="w-4 h-4 text-blue-500" />
          Available Globals
        </h2>
        <Card className="bg-slate-900 border-slate-800">
          <CardContent className="pt-5">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-2">
              {[
                { name: 'console', desc: 'log, warn, error, info, debug' },
                { name: 'Math', desc: 'sqrt, pow, random, floor, ceil…' },
                { name: 'Date', desc: 'now, parse, UTC' },
                { name: 'JSON', desc: 'parse, stringify' },
                { name: 'Array', desc: 'map, filter, reduce, find…' },
                { name: 'Object', desc: 'keys, values, entries, assign' },
                { name: 'String', desc: 'split, replace, indexOf…' },
                { name: 'Promise', desc: 'resolve, reject, all, race' },
              ].map(g => (
                <div key={g.name} className="p-2.5 bg-slate-950 rounded-md border border-slate-800">
                  <code className="text-blue-400 text-xs font-mono">{g.name}</code>
                  <p className="text-[10px] text-slate-500 mt-0.5">{g.desc}</p>
                </div>
              ))}
            </div>
            <Separator className="bg-slate-800 my-4" />
            <div className="text-xs text-slate-500 space-y-1">
              <p>The runtime uses goja (ECMAScript 5.1). JavaScript built-ins are always available.</p>
              <p>Template modules refer to native modules (for example <code className="font-mono">require(\"fs\")</code>) configured per template.</p>
              <p>The last expression in REPL input is automatically returned. All console output is captured as events.</p>
            </div>
          </CardContent>
        </Card>
      </section>

      {/* Code examples */}
      <section className="space-y-3">
        <h2 className="text-sm font-medium text-slate-200 flex items-center gap-2">
          <FileCode className="w-4 h-4 text-blue-500" />
          Code Examples
        </h2>
        <div className="space-y-2">
          {PRESET_EXAMPLES.map(example => (
            <CollapsibleExample key={example.id} name={example.name} description={example.description} code={example.code} />
          ))}
        </div>
      </section>

      {/* Execution model note */}
      <section className="space-y-3">
        <h2 className="text-sm font-medium text-slate-200 flex items-center gap-2">
          <BookOpen className="w-4 h-4 text-blue-500" />
          Execution Model
        </h2>
        <Card className="bg-slate-900 border-slate-800">
          <CardContent className="pt-5 text-xs text-slate-400 leading-relaxed space-y-2">
            <p>
              <strong className="text-slate-300">Startup:</strong> Runs automatically when a session is created. Startup files are
              loaded in order from the template's startup file list. If any startup file fails, the session enters a crashed state.
            </p>
            <p>
              <strong className="text-slate-300">REPL:</strong> Evaluates a code snippet in the session's context. The last expression
              is returned. Session state persists across REPL calls.
            </p>
            <p>
              <strong className="text-slate-300">Run-file:</strong> Executes a workspace file as an entry point. The file path is
              resolved within the session's worktree.
            </p>
            <p>
              <strong className="text-slate-300">Events:</strong> Every execution generates timestamped events — console output,
              return values, exceptions, system messages. Events are stored with monotonically increasing sequence numbers.
            </p>
          </CardContent>
        </Card>
      </section>

      {/* Limitations */}
      <div className="rounded-md bg-slate-900 border border-slate-800 px-4 py-3 text-xs text-slate-400 space-y-1">
        <p className="text-slate-300 font-medium">Limitations</p>
        <ul className="list-disc list-inside space-y-0.5 text-slate-500">
          <li>No DOM or browser APIs (window, document, etc.)</li>
          <li>File execution scoped to the session worktree path</li>
          <li>Resource limits (CPU, wall time, memory, events/output) are template-defined</li>
          <li>Runtime sessions are daemon-owned — daemon restart clears in-memory state</li>
        </ul>
      </div>
    </div>
  );
}

function ApiRow({ method, path, desc }: { method: string; path: string; desc: string }) {
  const methodColor = method === 'GET' ? 'text-blue-400' : method === 'POST' ? 'text-emerald-400' : 'text-red-400';
  return (
    <tr className="border-b border-slate-800/30">
      <td className={`py-1.5 pr-3 font-mono font-medium ${methodColor}`}>{method}</td>
      <td className="py-1.5 pr-3 font-mono text-slate-400 whitespace-nowrap">{path}</td>
      <td className="py-1.5 text-slate-500">{desc}</td>
    </tr>
  );
}

function CollapsibleExample({ name, description, code }: { name: string; description: string; code: string }) {
  const [open, setOpen] = useState(false);
  return (
    <Card className="bg-slate-900 border-slate-800">
      <button
        className="w-full flex items-center gap-2 px-4 py-3 text-left hover:bg-slate-800/30 transition-colors"
        onClick={() => setOpen(!open)}
      >
        {open ? <ChevronDown className="w-3.5 h-3.5 text-slate-500" /> : <ChevronRight className="w-3.5 h-3.5 text-slate-500" />}
        <span className="text-sm text-slate-200 font-medium">{name}</span>
        <span className="text-xs text-slate-500 ml-2">{description}</span>
      </button>
      {open && (
        <CardContent className="pt-0 pb-4 px-4">
          <pre className="bg-slate-950 p-3 rounded-md overflow-x-auto text-xs font-mono text-slate-300">
            {code}
          </pre>
        </CardContent>
      )}
    </Card>
  );
}
