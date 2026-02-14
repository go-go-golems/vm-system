import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { CreateSessionDialog } from '@/components/CreateSessionDialog';
import { useListTemplatesQuery, useListSessionsQuery } from '@/lib/api';
import type { VMSession } from '@/lib/types';
import {
  AlertCircle,
  CheckCircle,
  Layers,
  Loader2,
  Plus,
  XCircle,
} from 'lucide-react';
import { Link, useLocation } from 'wouter';
import { useState } from 'react';

export default function Sessions() {
  const { data: templates = [] } = useListTemplatesQuery();
  const { data: sessions = [], isLoading } = useListSessionsQuery();
  const [, setLocation] = useLocation();
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [createDialogOpen, setCreateDialogOpen] = useState(false);

  const filtered = sessions.filter((s: VMSession) =>
    statusFilter === 'all' ? true : s.status === statusFilter,
  );

  const handleSessionCreated = async (sessionId: string) => {
    setLocation(`/sessions/${sessionId}`);
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'ready': return <CheckCircle className="w-3.5 h-3.5 text-emerald-500" />;
      case 'starting': return <Loader2 className="w-3.5 h-3.5 text-blue-500 animate-spin" />;
      case 'crashed': return <XCircle className="w-3.5 h-3.5 text-red-500" />;
      case 'closed': return <AlertCircle className="w-3.5 h-3.5 text-slate-500" />;
      default: return <AlertCircle className="w-3.5 h-3.5 text-slate-500" />;
    }
  };

  const statusColor = (status: string) => {
    switch (status) {
      case 'ready': return 'bg-emerald-950 border-emerald-800 text-emerald-400';
      case 'starting': return 'bg-blue-950 border-blue-800 text-blue-400';
      case 'crashed': return 'bg-red-950 border-red-800 text-red-400';
      default: return 'bg-slate-800 border-slate-700 text-slate-400';
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64 text-slate-500">
        <Loader2 className="w-5 h-5 animate-spin mr-2" /> Loading…
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold text-slate-100">Sessions</h1>
          <p className="text-sm text-slate-400">Running VM instances created from templates.</p>
        </div>
        <Button size="sm" className="bg-blue-600 hover:bg-blue-700 text-white" onClick={() => setCreateDialogOpen(true)}>
          <Plus className="w-4 h-4 mr-1.5" /> New Session
        </Button>
      </div>

      <div className="flex items-center gap-3">
        <Select value={statusFilter} onValueChange={setStatusFilter}>
          <SelectTrigger className="w-40 bg-slate-900 border-slate-700 text-slate-300 h-8 text-xs">
            <SelectValue placeholder="Filter by status" />
          </SelectTrigger>
          <SelectContent className="bg-slate-900 border-slate-700">
            <SelectItem value="all" className="text-slate-300 text-xs">All ({sessions.length})</SelectItem>
            <SelectItem value="ready" className="text-slate-300 text-xs">Ready ({sessions.filter((s: VMSession) => s.status === 'ready').length})</SelectItem>
            <SelectItem value="starting" className="text-slate-300 text-xs">Starting ({sessions.filter((s: VMSession) => s.status === 'starting').length})</SelectItem>
            <SelectItem value="crashed" className="text-slate-300 text-xs">Crashed ({sessions.filter((s: VMSession) => s.status === 'crashed').length})</SelectItem>
            <SelectItem value="closed" className="text-slate-300 text-xs">Closed ({sessions.filter((s: VMSession) => s.status === 'closed').length})</SelectItem>
          </SelectContent>
        </Select>
        <span className="text-xs text-slate-500">{filtered.length} session{filtered.length !== 1 ? 's' : ''}</span>
      </div>

      {filtered.length === 0 ? (
        <div className="flex flex-col items-center justify-center h-48 text-slate-500">
          <Layers className="w-10 h-10 mb-2 opacity-40" />
          <p className="text-sm">{sessions.length === 0 ? 'No sessions yet.' : 'No sessions match the filter.'}</p>
        </div>
      ) : (
        <div className="border border-slate-800 rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-slate-800 bg-slate-900/50">
                <th className="text-left px-4 py-2.5 text-slate-400 font-medium w-10">Status</th>
                <th className="text-left px-4 py-2.5 text-slate-400 font-medium">Name</th>
                <th className="text-left px-4 py-2.5 text-slate-400 font-medium hidden md:table-cell">Template</th>
                <th className="text-left px-4 py-2.5 text-slate-400 font-medium hidden lg:table-cell">Created</th>
                <th className="text-right px-4 py-2.5 text-slate-400 font-medium">Action</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((session: VMSession) => {
                const tpl = templates.find(t => t.id === session.vmId);
                return (
                  <tr key={session.id} className="border-b border-slate-800/50 hover:bg-slate-900/30 transition-colors">
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        {getStatusIcon(session.status)}
                        <Badge variant="outline" className={`text-[10px] ${statusColor(session.status)}`}>{session.status}</Badge>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <Link href={`/sessions/${session.id}`} className="text-slate-200 hover:text-white font-medium transition-colors">{session.name}</Link>
                      <div className="text-xs text-slate-600 font-mono">{session.id.slice(0, 12)}…</div>
                    </td>
                    <td className="px-4 py-3 hidden md:table-cell">
                      <Link href={`/templates/${session.vmId}`} className="text-slate-400 hover:text-blue-400 transition-colors text-xs">
                        {tpl?.name || session.vmProfile}
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-slate-500 text-xs hidden lg:table-cell">{formatRelativeTime(session.createdAt)}</td>
                    <td className="px-4 py-3 text-right">
                      <Link href={`/sessions/${session.id}`}>
                        <Button size="sm" variant="outline" className="bg-slate-900 border-slate-700 text-slate-300 hover:bg-slate-800 h-7 px-2.5 text-xs">Open</Button>
                      </Link>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      <CreateSessionDialog open={createDialogOpen} onOpenChange={setCreateDialogOpen} onCreated={handleSessionCreated} />
    </div>
  );
}

function formatRelativeTime(dateStr: string): string {
  const now = Date.now();
  const diff = now - new Date(dateStr).getTime();
  const seconds = Math.floor(diff / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);
  if (days > 0) return `${days}d ago`;
  if (hours > 0) return `${hours}h ago`;
  if (minutes > 0) return `${minutes}m ago`;
  return `${seconds}s ago`;
}
