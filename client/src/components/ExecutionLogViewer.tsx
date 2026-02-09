import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import type { Execution, ExecutionEvent } from '@/lib/vmService';
import {
  AlertCircle,
  CheckCircle,
  ChevronDown,
  ChevronRight,
  Clock,
  Code,
  FileCode,
  Filter,
  Search,
  Terminal,
  XCircle,
} from 'lucide-react';
import { useMemo, useState } from 'react';

interface ExecutionLogViewerProps {
  executions: Execution[];
  onSelectExecution?: (execution: Execution) => void;
}

export function ExecutionLogViewer({ executions, onSelectExecution }: ExecutionLogViewerProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [expandedExecutions, setExpandedExecutions] = useState<Set<string>>(new Set());



  // Filter executions
  const filteredExecutions = useMemo(() => {
    return executions.filter((exec) => {
      // Status filter
      if (statusFilter !== 'all' && exec.status !== statusFilter) {
        return false;
      }



      // Search query
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        const matchesInput = exec.input?.toLowerCase().includes(query);
        const matchesError = exec.error?.toLowerCase().includes(query);
        const matchesEvents = exec.events.some((event) =>
          JSON.stringify(event.payload).toLowerCase().includes(query)
        );
        return matchesInput || matchesError || matchesEvents;
      }

      return true;
    });
  }, [executions, searchQuery, statusFilter]);

  const toggleExpanded = (executionId: string) => {
    const newExpanded = new Set(expandedExecutions);
    if (newExpanded.has(executionId)) {
      newExpanded.delete(executionId);
    } else {
      newExpanded.add(executionId);
    }
    setExpandedExecutions(newExpanded);
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'ok':
        return <CheckCircle className="w-4 h-4 text-emerald-500" />;
      case 'error':
        return <XCircle className="w-4 h-4 text-red-500" />;
      case 'running':
        return <Clock className="w-4 h-4 text-blue-500 animate-pulse" />;
      default:
        return <AlertCircle className="w-4 h-4 text-slate-500" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'ok':
        return 'text-emerald-500';
      case 'error':
        return 'text-red-500';
      case 'running':
        return 'text-blue-500';
      default:
        return 'text-slate-500';
    }
  };

  const getEventIcon = (type: string) => {
    switch (type) {
      case 'console':
        return <Terminal className="w-3 h-3 text-blue-400" />;
      case 'value':
        return <Code className="w-3 h-3 text-emerald-400" />;
      case 'exception':
        return <XCircle className="w-3 h-3 text-red-400" />;
      case 'input_echo':
        return <FileCode className="w-3 h-3 text-slate-400" />;
      case 'stdout':
        return <Terminal className="w-3 h-3 text-slate-300" />;
      case 'stderr':
        return <AlertCircle className="w-3 h-3 text-amber-400" />;
      case 'system':
        return <AlertCircle className="w-3 h-3 text-blue-400" />;
      default:
        return <Terminal className="w-3 h-3 text-slate-400" />;
    }
  };

  const formatEventPayload = (event: ExecutionEvent) => {
    if (typeof event.payload === 'string') {
      return event.payload;
    }
    if (event.payload?.text) {
      return event.payload.text;
    }
    if (event.payload?.message) {
      return event.payload.message;
    }
    if (event.type === 'value' && event.payload?.preview) {
      return event.payload.preview;
    }
    return JSON.stringify(event.payload, null, 2);
  };

  const formatKind = (kind: Execution['kind']) => {
    switch (kind) {
      case 'repl':
        return 'REPL';
      case 'run-file':
        return 'Run File';
      case 'startup':
        return 'Startup';
      default:
        return kind;
    }
  };

  const formatTimestamp = (date: Date) => {
    return new Intl.DateTimeFormat('en-US', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      fractionalSecondDigits: 3,
    }).format(date);
  };

  const formatDuration = (start: Date, end?: Date) => {
    if (!end) return '—';
    const ms = end.getTime() - start.getTime();
    return `${ms}ms`;
  };

  return (
    <div className="h-full flex flex-col bg-slate-950">
      {/* Header */}
      <div className="p-4 border-b border-slate-800">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-lg font-semibold text-slate-100">Execution Log</h2>
            <p className="text-sm text-slate-400">
              {filteredExecutions.length} of {executions.length} executions for current session
            </p>
          </div>
        </div>

        {/* Filters */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
            <Input
              placeholder="Search executions..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9 bg-slate-900 border-slate-700 text-slate-300"
            />
          </div>

          <Select value={statusFilter} onValueChange={setStatusFilter}>
            <SelectTrigger className="bg-slate-900 border-slate-700 text-slate-300">
              <Filter className="w-4 h-4 mr-2" />
              <SelectValue placeholder="Status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Status</SelectItem>
              <SelectItem value="ok">Success</SelectItem>
              <SelectItem value="error">Error</SelectItem>
              <SelectItem value="running">Running</SelectItem>
            </SelectContent>
          </Select>


        </div>
      </div>

      {/* Execution list */}
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {filteredExecutions.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-slate-500">
            <Terminal className="w-12 h-12 mb-3 opacity-50" />
            <p className="text-sm">No executions found</p>
          </div>
        ) : (
          filteredExecutions.map((execution) => {
            const isExpanded = expandedExecutions.has(execution.id);

            return (
              <Card
                key={execution.id}
                className="bg-slate-900 border-slate-800 hover:border-slate-700 transition-colors"
              >
                <CardHeader className="p-4 pb-3">
                  <div className="flex items-start justify-between gap-3">
                    <div className="flex items-start gap-3 flex-1 min-w-0">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => toggleExpanded(execution.id)}
                        className="p-0 h-6 w-6 flex-shrink-0"
                      >
                        {isExpanded ? (
                          <ChevronDown className="w-4 h-4" />
                        ) : (
                          <ChevronRight className="w-4 h-4" />
                        )}
                      </Button>

                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-1">
                          {getStatusIcon(execution.status)}
                          <span
                            className={`text-xs font-medium uppercase ${getStatusColor(execution.status)}`}
                          >
                            {execution.status}
                          </span>
                          <span className="text-xs text-slate-500">
                            {formatKind(execution.kind)}
                          </span>
                          <span className="text-xs text-slate-600">•</span>
                          <span className="text-xs text-slate-500">
                            {formatTimestamp(execution.startedAt)}
                          </span>
                        </div>

                        {execution.input && (
                          <pre className="text-xs text-slate-400 font-mono overflow-hidden text-ellipsis whitespace-nowrap">
                            {execution.input}
                          </pre>
                        )}

                        {execution.error && (
                          <p className="text-xs text-red-400 mt-1">{execution.error}</p>
                        )}
                      </div>
                    </div>

                    <div className="text-right flex-shrink-0">
                      <div className="text-xs text-slate-500">
                        {formatDuration(execution.startedAt, execution.endedAt)}
                      </div>
                      <div className="text-xs text-slate-600 mt-1">
                        {execution.events.length} events
                      </div>
                    </div>
                  </div>
                </CardHeader>

                {isExpanded && (
                  <CardContent className="p-4 pt-0">
                    <Separator className="mb-3 bg-slate-800" />

                    {/* Execution details */}
                    <div className="space-y-2 mb-3">
                      <div className="grid grid-cols-2 gap-2 text-xs">
                        <div>
                          <span className="text-slate-500">Execution ID:</span>
                          <span className="text-slate-300 ml-2 font-mono">
                            {execution.id.substring(0, 12)}...
                          </span>
                        </div>
                        <div>
                          <span className="text-slate-500">Session ID:</span>
                          <span className="text-slate-300 ml-2 font-mono">
                            {execution.sessionId.substring(0, 12)}...
                          </span>
                        </div>
                      </div>

                      {execution.result !== undefined && (
                        <div>
                          <span className="text-slate-500 text-xs">Result:</span>
                          <pre className="text-xs text-emerald-400 font-mono mt-1 p-2 bg-slate-950 rounded border border-slate-800 overflow-x-auto">
                            {JSON.stringify(execution.result, null, 2)}
                          </pre>
                        </div>
                      )}
                    </div>

                    {/* Events */}
                    <div>
                      <h4 className="text-xs font-semibold text-slate-400 mb-2">Events</h4>
                      <div className="space-y-2">
                        {execution.events.map((event) => (
                          <div
                            key={`${execution.id}-${event.seq}`}
                            className="flex items-start gap-2 p-2 bg-slate-950 rounded border border-slate-800"
                          >
                            <div className="flex items-center gap-2 flex-shrink-0">
                              {getEventIcon(event.type)}
                              <span className="text-xs text-slate-500 font-mono w-8">
                                #{event.seq}
                              </span>
                            </div>

                            <div className="flex-1 min-w-0">
                              <div className="flex items-center gap-2 mb-1">
                                <span className="text-xs font-medium text-slate-300">
                                  {event.type}
                                </span>
                                <span className="text-xs text-slate-600">
                                  {formatTimestamp(event.ts)}
                                </span>
                              </div>

                              <pre className="text-xs text-slate-400 font-mono overflow-x-auto">
                                {formatEventPayload(event)}
                              </pre>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>

                    {onSelectExecution && (
                      <Button
                        onClick={() => onSelectExecution(execution)}
                        variant="outline"
                        size="sm"
                        className="mt-3 w-full bg-slate-800 border-slate-700 text-slate-300 hover:bg-slate-700"
                      >
                        Load in Editor
                      </Button>
                    )}
                  </CardContent>
                )}
              </Card>
            );
          })
        )}
      </div>
    </div>
  );
}
