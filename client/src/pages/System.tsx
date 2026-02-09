import { Card, CardContent } from '@/components/ui/card';
import { useListTemplatesQuery, useListSessionsQuery } from '@/lib/api';
import type { VMSession } from '@/lib/types';
import {
  Activity,
  Box,
  Database,
  Layers,
  Loader2,
  Terminal,
  Zap,
} from 'lucide-react';

export default function System() {
  const { data: templates = [], isLoading: loadingTemplates } = useListTemplatesQuery();
  const { data: sessions = [], isLoading: loadingSessions } = useListSessionsQuery();

  if (loadingTemplates || loadingSessions) {
    return (
      <div className="flex items-center justify-center h-64 text-slate-500">
        <Loader2 className="w-5 h-5 animate-spin mr-2" />
        Loading…
      </div>
    );
  }

  const readyCount = sessions.filter((s: VMSession) => s.status === 'ready').length;
  const closedCount = sessions.filter((s: VMSession) => s.status === 'closed').length;
  const crashedCount = sessions.filter((s: VMSession) => s.status === 'crashed').length;

  return (
    <div className="space-y-4 max-w-3xl">
      <div>
        <h1 className="text-lg font-semibold text-slate-100">System</h1>
        <p className="text-sm text-slate-400">Runtime health and daemon status.</p>
      </div>

      {/* Status cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        <Card className="bg-slate-900 border-slate-800">
          <CardContent className="pt-4 pb-4">
            <div className="flex items-center gap-2 mb-2">
              <Activity className="w-4 h-4 text-emerald-500" />
              <span className="text-xs text-slate-400">Daemon</span>
            </div>
            <div className="text-sm font-medium text-emerald-400">Connected</div>
            <div className="text-xs text-slate-500 mt-0.5">API: /api/v1</div>
          </CardContent>
        </Card>

        <Card className="bg-slate-900 border-slate-800">
          <CardContent className="pt-4 pb-4">
            <div className="flex items-center gap-2 mb-2">
              <Box className="w-4 h-4 text-blue-500" />
              <span className="text-xs text-slate-400">Templates</span>
            </div>
            <div className="text-sm font-medium text-slate-200">{templates.length} active</div>
          </CardContent>
        </Card>

        <Card className="bg-slate-900 border-slate-800">
          <CardContent className="pt-4 pb-4">
            <div className="flex items-center gap-2 mb-2">
              <Layers className="w-4 h-4 text-blue-500" />
              <span className="text-xs text-slate-400">Sessions</span>
            </div>
            <div className="text-sm font-medium text-slate-200">
              {readyCount} ready
              {closedCount > 0 && <span className="text-slate-500"> / {closedCount} closed</span>}
              {crashedCount > 0 && <span className="text-red-400"> / {crashedCount} crashed</span>}
            </div>
          </CardContent>
        </Card>

        <Card className="bg-slate-900 border-slate-800">
          <CardContent className="pt-4 pb-4">
            <div className="flex items-center gap-2 mb-2">
              <Zap className="w-4 h-4 text-blue-500" />
              <span className="text-xs text-slate-400">Total</span>
            </div>
            <div className="text-sm font-medium text-slate-200">{sessions.length} sessions</div>
          </CardContent>
        </Card>
      </div>

      {/* Runtime info */}
      <Card className="bg-slate-900 border-slate-800">
        <CardContent className="pt-5 space-y-3">
          <h2 className="text-sm font-medium text-slate-200 flex items-center gap-2">
            <Terminal className="w-4 h-4 text-blue-500" />
            Runtime
          </h2>
          <div className="grid grid-cols-2 gap-y-2 text-xs">
            <span className="text-slate-500">Engine</span>
            <span className="text-slate-300">goja (ECMAScript 5.1, pure Go)</span>
            <span className="text-slate-500">Storage</span>
            <span className="text-slate-300">Git + SQLite dual-storage</span>
            <span className="text-slate-500">Default limits</span>
            <span className="text-slate-300">cpu 2000ms, wall 5000ms, mem 128MB</span>
            <span className="text-slate-500">Security model</span>
            <span className="text-slate-300">Deny-by-default capabilities</span>
          </div>
        </CardContent>
      </Card>

      {/* Architecture note */}
      <Card className="bg-slate-900 border-slate-800">
        <CardContent className="pt-5 space-y-3">
          <h2 className="text-sm font-medium text-slate-200 flex items-center gap-2">
            <Database className="w-4 h-4 text-blue-500" />
            Architecture
          </h2>
          <div className="text-xs text-slate-400 leading-relaxed space-y-2">
            <p>
              The VM system runs as a long-lived daemon. Templates define configuration blueprints.
              Sessions are runtime instances created from templates. Each session gets its own goja
              runtime and isolated Git worktree.
            </p>
            <p>
              Executions (REPL, run-file, startup) are tracked with timestamped event streams.
              All state is persisted in SQLite alongside Git objects.
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Warning */}
      <div className="rounded-md bg-blue-950/30 border border-blue-800/40 px-3 py-2 text-xs text-blue-300">
        ⓘ Runtime sessions are daemon-owned. Restarting the daemon destroys in-memory state.
        Persisted records (templates, session metadata, executions) remain in the database.
      </div>
    </div>
  );
}
