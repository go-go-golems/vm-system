import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import type { Execution, ExecutionEvent } from '@/lib/types';
import { CheckCircle2, Trash2, XCircle, Terminal } from 'lucide-react';
import { useEffect, useRef } from 'react';

interface ExecutionConsoleProps {
  executions: Execution[];
  onClear?: () => void;
  className?: string;
}

export function ExecutionConsole({ executions, onClear, className }: ExecutionConsoleProps) {
  const bottomRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom when new executions arrive
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [executions.length]);

  return (
    <div className={cn('flex flex-col h-full bg-slate-950 border border-slate-800 rounded-lg min-h-0', className)}>
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-1.5 border-b border-slate-800 flex-shrink-0">
        <div className="flex items-center gap-2">
          <Terminal className="w-3.5 h-3.5 text-slate-400" />
          <span className="text-xs font-medium text-slate-300">Console Output</span>
          {executions.length > 0 && (
            <span className="text-[10px] text-slate-600">{executions.length} execution{executions.length !== 1 ? 's' : ''}</span>
          )}
        </div>
        {onClear && executions.length > 0 && (
          <Button
            size="sm"
            variant="ghost"
            onClick={onClear}
            className="h-6 px-2 text-slate-500 hover:text-slate-300 hover:bg-slate-800"
          >
            <Trash2 className="w-3 h-3 mr-1" />
            <span className="text-xs">Clear</span>
          </Button>
        )}
      </div>

      {/* Scrollable output */}
      <div className="flex-1 overflow-y-auto min-h-0 p-3">
        {executions.length === 0 ? (
          <div className="flex items-center justify-center h-full text-slate-600">
            <p className="text-xs">No output yet. Run some code.</p>
          </div>
        ) : (
          <div className="space-y-3">
            {executions.map((execution) => (
              <ExecutionBlock key={execution.id} execution={execution} />
            ))}
            <div ref={bottomRef} />
          </div>
        )}
      </div>
    </div>
  );
}

function ExecutionBlock({ execution }: { execution: Execution }) {
  return (
    <div className="space-y-1">
      {/* Execution header */}
      <div className="flex items-center gap-2">
        {execution.status === 'ok' && <CheckCircle2 className="w-3.5 h-3.5 text-emerald-500" />}
        {execution.status === 'error' && <XCircle className="w-3.5 h-3.5 text-rose-500" />}
        {execution.status === 'running' && (
          <div className="w-3.5 h-3.5 border-2 border-blue-500 border-t-transparent rounded-full animate-spin" />
        )}
        <span className="text-[10px] text-slate-500 font-mono">
          {new Date(execution.startedAt).toLocaleTimeString()}
        </span>
      </div>

      {/* Events */}
      <div className="space-y-0.5 pl-5">
        {execution.events.map((event) => (
          <EventLine key={event.seq} event={event} />
        ))}

        {/* Result or error */}
        {execution.status === 'ok' && execution.result !== undefined && (
          <div className="font-mono text-xs text-blue-400">
            <span className="text-slate-600">→ </span>
            {formatValue(execution.result)}
          </div>
        )}

        {execution.status === 'error' && execution.error && (
          <div className="font-mono text-xs text-rose-400">
            <span className="text-slate-600">✗ </span>
            {execution.error}
          </div>
        )}
      </div>
    </div>
  );
}

function EventLine({ event }: { event: ExecutionEvent }) {
  const textPayload =
    typeof event.payload === 'string'
      ? event.payload
      : event.payload?.text || event.payload?.message || JSON.stringify(event.payload);

  switch (event.type) {
    case 'input_echo':
      return (
        <div className="font-mono text-xs text-slate-400">
          <span className="text-slate-600">&gt; </span>
          {textPayload}
        </div>
      );

    case 'console':
      return <div className="font-mono text-xs text-slate-300">{textPayload}</div>;

    case 'stdout':
      return <div className="font-mono text-xs text-slate-300">{textPayload}</div>;

    case 'stderr':
      return <div className="font-mono text-xs text-amber-400">{textPayload}</div>;

    case 'system':
      return <div className="font-mono text-xs text-blue-300">{textPayload}</div>;

    case 'value':
      return null; // Handled separately in ExecutionBlock

    case 'exception':
      return (
        <div className="font-mono text-xs text-rose-400">
          <span className="text-slate-600">✗ </span>
          {textPayload}
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
