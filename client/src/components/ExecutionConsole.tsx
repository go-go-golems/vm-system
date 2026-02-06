import { ScrollArea } from '@/components/ui/scroll-area';
import { cn } from '@/lib/utils';
import type { Execution, ExecutionEvent } from '@/lib/vmService';
import { CheckCircle2, XCircle, Terminal } from 'lucide-react';

interface ExecutionConsoleProps {
  executions: Execution[];
  className?: string;
}

export function ExecutionConsole({ executions, className }: ExecutionConsoleProps) {
  return (
    <div className={cn('flex flex-col h-full bg-slate-950 border border-slate-800 rounded-lg', className)}>
      <div className="flex items-center gap-2 px-4 py-2 border-b border-slate-800">
        <Terminal className="w-4 h-4 text-slate-400" />
        <span className="text-sm font-medium text-slate-300">Console Output</span>
      </div>

      <ScrollArea className="flex-1 p-4">
        {executions.length === 0 ? (
          <div className="flex items-center justify-center h-full text-slate-600">
            <p className="text-sm">No executions yet. Run some code to see output.</p>
          </div>
        ) : (
          <div className="space-y-4">
            {executions.map((execution) => (
              <ExecutionBlock key={execution.id} execution={execution} />
            ))}
          </div>
        )}
      </ScrollArea>
    </div>
  );
}

function ExecutionBlock({ execution }: { execution: Execution }) {
  return (
    <div className="space-y-2">
      {/* Execution header */}
      <div className="flex items-center gap-2">
        {execution.status === 'ok' && <CheckCircle2 className="w-4 h-4 text-emerald-500" />}
        {execution.status === 'error' && <XCircle className="w-4 h-4 text-rose-500" />}
        {execution.status === 'running' && (
          <div className="w-4 h-4 border-2 border-blue-500 border-t-transparent rounded-full animate-spin" />
        )}
        <span className="text-xs text-slate-500 font-mono">
          {execution.startedAt.toLocaleTimeString()}
        </span>
      </div>

      {/* Events */}
      <div className="space-y-1 pl-6">
        {execution.events.map((event) => (
          <EventLine key={event.seq} event={event} />
        ))}

        {/* Result or error */}
        {execution.status === 'ok' && execution.result !== undefined && (
          <div className="font-mono text-sm text-blue-400">
            <span className="text-slate-600">→ </span>
            {formatValue(execution.result)}
          </div>
        )}

        {execution.status === 'error' && execution.error && (
          <div className="font-mono text-sm text-rose-400">
            <span className="text-slate-600">✗ </span>
            {execution.error}
          </div>
        )}
      </div>
    </div>
  );
}

function EventLine({ event }: { event: ExecutionEvent }) {
  switch (event.type) {
    case 'input_echo':
      return (
        <div className="font-mono text-sm text-slate-400">
          <span className="text-slate-600">&gt; </span>
          {event.payload.text}
        </div>
      );

    case 'console':
      return (
        <div className="font-mono text-sm text-slate-300">
          {event.payload.text}
        </div>
      );

    case 'value':
      return null; // Handled separately in ExecutionBlock

    case 'exception':
      return (
        <div className="font-mono text-sm text-rose-400">
          <span className="text-slate-600">✗ </span>
          {event.payload.message}
        </div>
      );

    default:
      return null;
  }
}

function formatValue(value: any): string {
  if (value === null) return 'null';
  if (value === undefined) return 'undefined';
  if (typeof value === 'string') return `"${value}"`;
  if (typeof value === 'object') {
    try {
      return JSON.stringify(value, null, 2);
    } catch {
      return String(value);
    }
  }
  return String(value);
}
