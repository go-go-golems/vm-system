import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { CodeEditor } from '@/components/CodeEditor';
import { ExecutionConsole } from '@/components/ExecutionConsole';
import { ExecutionLogViewer } from '@/components/ExecutionLogViewer';
import { PresetSelector } from '@/components/PresetSelector';
import {
  useGetSessionQuery,
  useGetTemplateQuery,
  useListExecutionsQuery,
  useCloseSessionMutation,
  useExecuteREPLMutation,
} from '@/lib/api';
import { setCurrentSessionId } from '@/lib/uiSlice';
import type { Execution } from '@/lib/types';
import {
  ArrowLeft,
  History,
  Layers,
  Loader2,
  Play,
  RotateCcw,
  Square,
  Terminal,
} from 'lucide-react';
import { Link, useParams } from 'wouter';
import { useCallback, useEffect, useState } from 'react';
import { useDispatch } from 'react-redux';
import { toast } from 'sonner';

export default function SessionDetail() {
  const params = useParams<{ id: string }>();
  const sessionId = params.id;
  const dispatch = useDispatch();

  const { data: session, isLoading: loadingSession } = useGetSessionQuery(sessionId);
  const { data: template } = useGetTemplateQuery(session?.vmId ?? '', { skip: !session?.vmId });
  const { data: executions = [] } = useListExecutionsQuery(sessionId);
  const [closeSession] = useCloseSessionMutation();
  const [executeREPL] = useExecuteREPLMutation();

  const [code, setCode] = useState('// Write your JavaScript code here\nconsole.log("Hello, VM!");');
  const [isExecuting, setIsExecuting] = useState(false);
  const [activeTab, setActiveTab] = useState('repl');
  const [consoleClearedAt, setConsoleClearedAt] = useState<string | null>(null);

  // Executions visible in the console (respects clear; full list stays in Executions tab)
  const consoleExecutions = consoleClearedAt
    ? executions.filter(e => e.startedAt > consoleClearedAt)
    : executions;

  // Set as current session on mount
  useEffect(() => {
    if (session) {
      dispatch(setCurrentSessionId(session.id));
    }
  }, [session, dispatch]);

  const handleExecute = async () => {
    if (!code.trim()) {
      toast.error('Enter some code to execute');
      return;
    }
    if (!session || session.status !== 'ready') {
      toast.error('Session is not ready');
      return;
    }

    setIsExecuting(true);
    try {
      const execution = await executeREPL({ sessionId, input: code }).unwrap();
      if (execution.status === 'error') {
        toast.error('Execution failed', { description: execution.error });
      }
    } catch (error: any) {
      toast.error('Execution failed', { description: error?.message || error?.data?.message || 'Unknown error' });
    } finally {
      setIsExecuting(false);
    }
  };

  const handleLoadPreset = (presetCode: string) => {
    setCode(presetCode);
  };

  const handleSelectExecution = (execution: Execution) => {
    if (execution.input) {
      setCode(execution.input);
      setActiveTab('repl');
    }
  };

  const handleCloseSession = async () => {
    if (!session) return;
    try {
      await closeSession(session.id).unwrap();
      toast.success('Session closed');
    } catch (error: any) {
      toast.error('Failed to close session', { description: error?.message || error?.data?.message || 'Unknown error' });
    }
  };

  if (loadingSession) {
    return (
      <div className="flex items-center justify-center h-64 text-slate-500">
        <Loader2 className="w-5 h-5 animate-spin mr-2" />
        Loading…
      </div>
    );
  }

  if (!session) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-slate-500">
        <Layers className="w-10 h-10 mb-2 opacity-40" />
        <p className="text-sm">Session not found.</p>
        <Link href="/sessions" className="text-blue-400 text-sm mt-2 hover:underline">
          Back to sessions
        </Link>
      </div>
    );
  }

  const isReady = session.status === 'ready';

  return (
    <div className="flex flex-col h-[calc(100vh-8.5rem)]">
      {/* Header */}
      <div className="flex items-start justify-between gap-4 mb-3 flex-shrink-0">
        <div className="min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <Link href="/sessions" className="text-slate-400 hover:text-slate-200 transition-colors">
              <ArrowLeft className="w-4 h-4" />
            </Link>
            <h1 className="text-lg font-semibold text-slate-100 truncate">{session.name}</h1>
            <Badge
              variant="outline"
              className={`text-[10px] flex-shrink-0 ${
                isReady
                  ? 'bg-emerald-950 border-emerald-800 text-emerald-400'
                  : session.status === 'crashed'
                    ? 'bg-red-950 border-red-800 text-red-400'
                    : 'bg-slate-800 border-slate-700 text-slate-400'
              }`}
            >
              {session.status}
            </Badge>
          </div>
          <div className="flex items-center gap-3 text-xs text-slate-400">
            <Link href={`/templates/${session.vmId}`} className="hover:text-blue-400 transition-colors">
              Template: {template?.name || session.vmProfile}
            </Link>
            <span className="text-slate-700">·</span>
            <span className="font-mono text-slate-500">{session.id.slice(0, 12)}…</span>
            {template && (
              <>
                <span className="text-slate-700">·</span>
                <span>{template.exposedModules.length} modules, {template.libraries.length} libs</span>
              </>
            )}
          </div>
        </div>
        <div className="flex items-center gap-2 flex-shrink-0">
          {isReady && (
            <Button
              size="sm"
              variant="outline"
              className="bg-slate-900 border-slate-700 text-slate-300 h-7 text-xs"
              onClick={handleCloseSession}
            >
              <Square className="w-3 h-3 mr-1" />
              Close
            </Button>
          )}
        </div>
      </div>

      {/* Not-ready banner */}
      {!isReady && (
        <div className="rounded-md bg-amber-950/30 border border-amber-800/40 px-3 py-2 text-xs text-amber-300 mb-3 flex-shrink-0">
          ⚠ This session is <strong>{session.status}</strong>.
          {session.status === 'crashed' && session.lastError && (
            <span> Error: {session.lastError}</span>
          )}
          {session.status === 'closed' && <span> You can view execution history but cannot run new code.</span>}
        </div>
      )}

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col min-h-0">
        <TabsList className="bg-slate-900 border border-slate-800 flex-shrink-0 mb-3">
          <TabsTrigger value="repl" className="data-[state=active]:bg-slate-800 text-xs">
            <Terminal className="w-3.5 h-3.5 mr-1.5" />
            REPL
          </TabsTrigger>
          <TabsTrigger value="executions" className="data-[state=active]:bg-slate-800 text-xs">
            <History className="w-3.5 h-3.5 mr-1.5" />
            Executions ({executions.length})
          </TabsTrigger>
        </TabsList>

        {/* REPL tab */}
        <TabsContent value="repl" className="flex-1 mt-0 flex flex-col min-h-0">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-3 flex-1 min-h-0">
            {/* Left: Editor */}
            <div className="flex flex-col gap-2 min-h-0">
              <div className="flex items-center justify-between gap-2 flex-shrink-0">
                <PresetSelector onSelect={handleLoadPreset} />
                <div className="flex items-center gap-2">
                  <Button
                    size="sm"
                    onClick={handleExecute}
                    disabled={isExecuting || !isReady}
                    className="bg-blue-600 hover:bg-blue-700 text-white h-7 text-xs"
                  >
                    {isExecuting ? (
                      <>
                        <Loader2 className="w-3 h-3 mr-1 animate-spin" />
                        Running…
                      </>
                    ) : (
                      <>
                        <Play className="w-3 h-3 mr-1" />
                        Run
                      </>
                    )}
                  </Button>
                </div>
              </div>
              <div className="flex-1 min-h-0">
                <CodeEditor value={code} onChange={setCode} />
              </div>
            </div>

            {/* Right: Console output */}
            <div className="flex-1 min-h-0">
              <ExecutionConsole
                executions={consoleExecutions}
                onClear={() => setConsoleClearedAt(new Date().toISOString())}
              />
            </div>
          </div>

          {/* Compact execution strip */}
          {executions.length > 0 && (
            <div className="mt-3 border border-slate-800 rounded-lg overflow-hidden flex-shrink-0">
              <div className="px-3 py-1.5 bg-slate-900/50 border-b border-slate-800 text-xs text-slate-400 font-medium">
                Recent executions
              </div>
              <div className="max-h-28 overflow-y-auto">
                {[...executions].reverse().slice(0, 10).map((exec, i) => (
                  <button
                    key={exec.id}
                    className="flex items-center gap-3 w-full px-3 py-1.5 text-xs hover:bg-slate-900/30 transition-colors text-left"
                    onClick={() => handleSelectExecution(exec)}
                  >
                    <span className="text-slate-600 font-mono w-6">#{executions.length - i}</span>
                    <span className={`font-medium ${exec.status === 'ok' ? 'text-emerald-500' : exec.status === 'error' ? 'text-red-500' : 'text-slate-500'}`}>
                      {exec.status}
                    </span>
                    <span className="text-slate-500 truncate flex-1 font-mono">
                      {exec.input?.slice(0, 60) || exec.path || '—'}
                    </span>
                    <span className="text-slate-600 flex-shrink-0">
                      {new Date(exec.startedAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
                    </span>
                    {exec.endedAt && (
                      <span className="text-slate-600 flex-shrink-0">
                        {new Date(exec.endedAt).getTime() - new Date(exec.startedAt).getTime()}ms
                      </span>
                    )}
                  </button>
                ))}
              </div>
            </div>
          )}
        </TabsContent>

        {/* Executions tab */}
        <TabsContent value="executions" className="flex-1 mt-0 min-h-0">
          <div className="h-full border border-slate-800 rounded-lg overflow-hidden">
            <ExecutionLogViewer
              executions={executions}
              onSelectExecution={handleSelectExecution}
            />
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
