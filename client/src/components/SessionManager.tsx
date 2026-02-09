import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import type { VMSession } from '@/lib/vmService';
import {
  AlertCircle,
  Check,
  CheckCircle,
  Clock,
  Loader2,
  Plus,
  Terminal,
  XCircle,
} from 'lucide-react';
import { useState } from 'react';

interface SessionManagerProps {
  sessions: VMSession[];
  currentSession: VMSession | null;
  onCreateSession: (name?: string) => Promise<void>;
  onSelectSession: (sessionId: string) => Promise<void>;
  onCloseSession: (sessionId: string) => Promise<void>;
}

export function SessionManager({
  sessions,
  currentSession,
  onCreateSession,
  onSelectSession,
  onCloseSession,
}: SessionManagerProps) {
  const [isCreating, setIsCreating] = useState(false);
  const [newSessionName, setNewSessionName] = useState('');
  const [showNameInput, setShowNameInput] = useState(false);

  const handleCreate = async () => {
    setIsCreating(true);
    try {
      await onCreateSession(newSessionName || undefined);
      setNewSessionName('');
      setShowNameInput(false);
    } finally {
      setIsCreating(false);
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'ready':
        return <CheckCircle className="w-4 h-4 text-emerald-500" />;
      case 'starting':
        return <Loader2 className="w-4 h-4 text-blue-500 animate-spin" />;
      case 'crashed':
        return <XCircle className="w-4 h-4 text-red-500" />;
      case 'closed':
        return <AlertCircle className="w-4 h-4 text-slate-500" />;
      default:
        return <Clock className="w-4 h-4 text-slate-500" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'ready':
        return 'text-emerald-500';
      case 'starting':
        return 'text-blue-500';
      case 'crashed':
        return 'text-red-500';
      case 'closed':
        return 'text-slate-500';
      default:
        return 'text-slate-500';
    }
  };

  const formatTime = (date: Date) => {
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const seconds = Math.floor(diff / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (days > 0) return `${days}d ago`;
    if (hours > 0) return `${hours}h ago`;
    if (minutes > 0) return `${minutes}m ago`;
    return `${seconds}s ago`;
  };

  const formatLastActivity = (date: Date) => {
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const seconds = Math.floor(diff / 1000);
    const minutes = Math.floor(seconds / 60);

    if (minutes >= 5) {
      return <span className="text-amber-500">Idle {minutes}m</span>;
    }
    if (minutes > 0) {
      return <span className="text-slate-500">Active {minutes}m ago</span>;
    }
    return <span className="text-emerald-500">Active now</span>;
  };

  return (
    <div className="h-full flex flex-col bg-slate-950">
      {/* Header */}
      <div className="p-4 border-b border-slate-800">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-lg font-semibold text-slate-100">Sessions</h2>
            <p className="text-sm text-slate-400">{sessions.length} total sessions</p>
          </div>

          {!showNameInput ? (
            <Button
              onClick={() => setShowNameInput(true)}
              size="sm"
              className="bg-blue-600 hover:bg-blue-700 text-white"
            >
              <Plus className="w-4 h-4 mr-2" />
              New Session
            </Button>
          ) : (
            <Button
              onClick={() => setShowNameInput(false)}
              size="sm"
              variant="outline"
              className="bg-slate-900 border-slate-700 text-slate-300"
            >
              Cancel
            </Button>
          )}
        </div>

        {/* Create session input */}
        {showNameInput && (
          <div className="flex gap-2">
            <Input
              placeholder="Session name (optional)"
              value={newSessionName}
              onChange={(e) => setNewSessionName(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') handleCreate();
                if (e.key === 'Escape') setShowNameInput(false);
              }}
              className="bg-slate-900 border-slate-700 text-slate-300"
              autoFocus
            />
            <Button
              onClick={handleCreate}
              disabled={isCreating}
              size="sm"
              className="bg-blue-600 hover:bg-blue-700 text-white"
            >
              {isCreating ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <Check className="w-4 h-4" />
              )}
            </Button>
          </div>
        )}
      </div>

      {/* Session list */}
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {sessions.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-slate-500">
            <Terminal className="w-12 h-12 mb-3 opacity-50" />
            <p className="text-sm">No sessions yet</p>
            <p className="text-xs text-slate-600 mt-1">Create one to get started</p>
          </div>
        ) : (
          sessions.map((session) => {
            const isActive = currentSession?.id === session.id;
            const canSelect = session.status === 'ready' && !isActive;
            const canClose = session.status === 'ready';

            return (
              <Card
                key={session.id}
                className={`bg-slate-900 border-slate-800 transition-colors ${
                  isActive
                    ? 'ring-2 ring-blue-500 border-blue-500'
                    : 'hover:border-slate-700'
                }`}
              >
                <CardContent className="p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div className="flex items-start gap-3 flex-1 min-w-0">
                      <div className="mt-1">{getStatusIcon(session.status)}</div>

                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-1">
                          <h3 className="text-sm font-semibold text-slate-200 truncate">
                            {session.name}
                          </h3>
                          {isActive && (
                            <span className="px-2 py-0.5 bg-blue-950 border border-blue-800 text-blue-400 text-xs rounded-full">
                              Active
                            </span>
                          )}
                        </div>

                        <div className="space-y-1">
                          <div className="flex items-center gap-2 text-xs">
                            <span
                              className={`font-medium uppercase ${getStatusColor(session.status)}`}
                            >
                              {session.status}
                            </span>
                            <span className="text-slate-600">•</span>
                            <span className="text-slate-500">
                              Created {formatTime(session.createdAt)}
                            </span>
                          </div>

                          {session.status === 'ready' && (
                            <div className="text-xs">{formatLastActivity(session.lastActivityAt)}</div>
                          )}

                          {session.closedAt && (
                            <div className="text-xs text-slate-500">
                              Closed {formatTime(session.closedAt)}
                            </div>
                          )}

                          <div className="text-xs text-slate-600 font-mono">
                            {session.id.substring(0, 16)}...
                          </div>
                          
                          {/* Display enabled modules */}
                          {session.vm && session.vm.exposedModules && session.vm.exposedModules.length > 0 && (
                            <div className="mt-2 pt-2 border-t border-slate-800">
                              <div className="text-xs text-slate-500 mb-1">Modules:</div>
                              <div className="flex flex-wrap gap-1">
                                {session.vm.exposedModules.map((module) => (
                                  <span
                                    key={module}
                                    className="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-mono bg-blue-500/10 text-blue-400 border border-blue-500/20"
                                  >
                                    {module}
                                  </span>
                                ))}
                              </div>
                            </div>
                          )}
                          
                          {/* Display enabled libraries */}
                          {session.vm && session.vm.libraries && session.vm.libraries.length > 0 && (
                            <div className="mt-2 pt-2 border-t border-slate-800">
                              <div className="text-xs text-slate-500 mb-1">Libraries:</div>
                              <div className="flex flex-wrap gap-1">
                                {session.vm.libraries.map((library) => (
                                  <span
                                    key={library}
                                    className="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-mono bg-green-500/10 text-green-400 border border-green-500/20"
                                  >
                                    {library}
                                  </span>
                                ))}
                              </div>
                            </div>
                          )}
                        </div>
                      </div>
                    </div>

                    <div className="flex flex-col gap-2">
                      {canSelect && (
                        <Button
                          onClick={() => onSelectSession(session.id)}
                          size="sm"
                          variant="outline"
                          className="bg-slate-800 border-slate-700 text-slate-300 hover:bg-slate-700"
                        >
                          Switch
                        </Button>
                      )}

                      {canClose && (
                        <Button
                          onClick={() => onCloseSession(session.id)}
                          size="sm"
                          variant="outline"
                          className="bg-slate-800 border-slate-700 text-slate-300 hover:bg-slate-700"
                        >
                          Close
                        </Button>
                      )}
                    </div>
                  </div>
                </CardContent>
              </Card>
            );
          })
        )}
      </div>

      {/* Footer info */}
      <div className="p-4 border-t border-slate-800 bg-slate-900/50">
        <div className="text-xs text-slate-500 space-y-1">
          <p>Sessions are daemon-owned runtime instances.</p>
          <p>Close a ready session to release backend runtime resources.</p>
        </div>
      </div>
    </div>
  );
}
