import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { CodeEditor } from '@/components/CodeEditor';
import { ExecutionConsole } from '@/components/ExecutionConsole';
import { ExecutionLogViewer } from '@/components/ExecutionLogViewer';
import { PresetSelector } from '@/components/PresetSelector';
import { SessionManager } from '@/components/SessionManager';
import { VMInfo } from '@/components/VMInfo';
import { vmService, type Execution, type VMSession } from '@/lib/vmService';
import { BookOpen, History, Layers, Play, RotateCcw, Terminal } from 'lucide-react';
import { Link } from 'wouter';
import { useEffect, useState } from 'react';
import { toast } from 'sonner';

export default function Home() {
  const [code, setCode] = useState('// Write your JavaScript code here\nconsole.log("Hello, VM!");');
  const [executions, setExecutions] = useState<Execution[]>([]);
  const [isExecuting, setIsExecuting] = useState(false);
  const [sessions, setSessions] = useState<VMSession[]>([]);
  const [currentSession, setCurrentSession] = useState<VMSession | null>(null);
  const [activeTab, setActiveTab] = useState('editor');

  // Load sessions and executions on mount
  useEffect(() => {
    loadSessions();
    loadExecutions();
  }, []);

  const loadSessions = async () => {
    const sessionList = await vmService.listSessions();
    setSessions(sessionList);
    const current = vmService.getCurrentSession();
    setCurrentSession(current);
  };

  const loadExecutions = async () => {
    const current = vmService.getCurrentSession();
    if (current) {
      const sessionExecs = await vmService.getExecutionsBySession(current.id);
      setExecutions(sessionExecs);
    } else {
      setExecutions([]);
    }
  };

  const handleExecute = async () => {
    if (!code.trim()) {
      toast.error('Please enter some code to execute');
      return;
    }

    const current = vmService.getCurrentSession();
    if (!current) {
      toast.error('No active session');
      return;
    }

    setIsExecuting(true);
    try {
      const execution = await vmService.executeREPL(code);
      setExecutions((prev) => [...prev, execution]);

      if (execution.status === 'error') {
        toast.error('Execution failed', {
          description: execution.error,
        });
      } else {
        toast.success('Execution completed');
      }
    } catch (error: any) {
      toast.error('Execution failed', {
        description: error.message,
      });
    } finally {
      setIsExecuting(false);
    }
  };

  const handleClear = () => {
    setExecutions([]);
    toast.success('Console cleared');
  };

  const handleLoadPreset = (presetCode: string) => {
    setCode(presetCode);
    toast.success('Preset loaded');
  };

  const handleCreateSession = async (name?: string) => {
    try {
      const session = await vmService.createSession(name);
      await loadSessions();
      await vmService.setCurrentSession(session.id);
      setCurrentSession(session);
      setExecutions([]);
      toast.success('Session created', {
        description: `Now using ${session.name}`,
      });
    } catch (error: any) {
      toast.error('Failed to create session', {
        description: error.message,
      });
    }
  };

  const handleSelectSession = async (sessionId: string) => {
    try {
      await vmService.setCurrentSession(sessionId);
      const session = await vmService.getSession(sessionId);
      setCurrentSession(session);
      const sessionExecs = await vmService.getExecutionsBySession(sessionId);
      setExecutions(sessionExecs);
      toast.success('Session switched', {
        description: `Now using ${session?.name}`,
      });
      setActiveTab('editor');
    } catch (error: any) {
      toast.error('Failed to switch session', {
        description: error.message,
      });
    }
  };

  const handleCloseSession = async (sessionId: string) => {
    try {
      await vmService.closeSession(sessionId);
      await loadSessions();
      toast.success('Session closed');
    } catch (error: any) {
      toast.error('Failed to close session', {
        description: error.message,
      });
    }
  };

  const handleDeleteSession = async (sessionId: string) => {
    try {
      await vmService.deleteSession(sessionId);
      await loadSessions();
      toast.success('Session deleted');
    } catch (error: any) {
      toast.error('Failed to delete session', {
        description: error.message,
      });
    }
  };

  const handleSelectExecution = (execution: Execution) => {
    if (execution.input) {
      setCode(execution.input);
      setActiveTab('editor');
      toast.success('Code loaded from execution');
    }
  };

  return (
    <div className="min-h-screen flex flex-col bg-slate-950">
      {/* Header */}
      <header className="border-b border-slate-800 bg-slate-900/50 backdrop-blur">
        <div className="container py-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center">
                <Terminal className="w-6 h-6 text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold text-slate-100">VM System</h1>
                <p className="text-sm text-slate-400">JavaScript Execution Environment</p>
              </div>
            </div>

            <div className="flex items-center gap-4">
              <PresetSelector onSelect={handleLoadPreset} />
              <Link href="/docs">
                <Button
                  variant="outline"
                  className="bg-slate-900 border-slate-700 text-slate-300 hover:bg-slate-800 hover:text-slate-100"
                >
                  <BookOpen className="w-4 h-4 mr-2" />
                  Docs
                </Button>
              </Link>
              <Button
                onClick={handleExecute}
                disabled={isExecuting}
                className="bg-blue-600 hover:bg-blue-700 text-white"
              >
                {isExecuting ? (
                  <>
                    <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2" />
                    Running...
                  </>
                ) : (
                  <>
                    <Play className="w-4 h-4 mr-2" />
                    Run Code
                  </>
                )}
              </Button>
            </div>
          </div>
        </div>
      </header>

      {/* Main content with tabs */}
      <main className="flex-1 container py-4">
        <Tabs value={activeTab} onValueChange={setActiveTab} className="h-full flex flex-col">
          <TabsList className="bg-slate-900 border border-slate-800 mb-4">
            <TabsTrigger value="editor" className="data-[state=active]:bg-slate-800">
              <Terminal className="w-4 h-4 mr-2" />
              Editor
            </TabsTrigger>
            <TabsTrigger value="sessions" className="data-[state=active]:bg-slate-800">
              <Layers className="w-4 h-4 mr-2" />
              Sessions ({sessions.length})
            </TabsTrigger>
            <TabsTrigger value="logs" className="data-[state=active]:bg-slate-800">
              <History className="w-4 h-4 mr-2" />
              Execution Log ({executions.length})
            </TabsTrigger>
          </TabsList>

          <TabsContent value="editor" className="flex-1 mt-0">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 h-full">
              {/* Left: Code editor */}
              <div className="flex flex-col gap-4">
                <CodeEditor value={code} onChange={setCode} />
                {currentSession && (
                  <VMInfo
                    vm={vmService.getVMs()[0]}
                    session={currentSession}
                  />
                )}
              </div>

              {/* Right: Output console */}
              <div className="flex flex-col gap-4">
                <ExecutionConsole executions={executions} />
                <Button
                  onClick={handleClear}
                  variant="outline"
                  size="sm"
                  className="bg-slate-900 border-slate-700 text-slate-300"
                >
                  <RotateCcw className="w-4 h-4 mr-2" />
                  Clear Console
                </Button>
              </div>
            </div>
          </TabsContent>

          <TabsContent value="sessions" className="flex-1 mt-0">
            <div className="h-[calc(100vh-12rem)] border border-slate-800 rounded-lg overflow-hidden">
              <SessionManager
                sessions={sessions}
                currentSession={currentSession}
                onCreateSession={handleCreateSession}
                onSelectSession={handleSelectSession}
                onCloseSession={handleCloseSession}
                onDeleteSession={handleDeleteSession}
              />
            </div>
          </TabsContent>

          <TabsContent value="logs" className="flex-1 mt-0">
            <div className="h-[calc(100vh-12rem)] border border-slate-800 rounded-lg overflow-hidden">
              {currentSession ? (
                <ExecutionLogViewer
                  executions={executions}
                  onSelectExecution={handleSelectExecution}
                />
              ) : (
                <div className="flex flex-col items-center justify-center h-full text-slate-500">
                  <History className="w-12 h-12 mb-3 opacity-50" />
                  <p className="text-sm">No active session</p>
                  <p className="text-xs text-slate-600 mt-1">Create or select a session to view execution logs</p>
                </div>
              )}
            </div>
          </TabsContent>
        </Tabs>
      </main>

      {/* Footer */}
      <footer className="border-t border-slate-800 bg-slate-900/50 backdrop-blur">
        <div className="container py-3">
          <div className="flex items-center justify-between text-sm text-slate-500">
            <div>Built with goja VM system</div>
            <div className="flex items-center gap-4">
              <span>
                Session: {currentSession?.name || 'None'} •{' '}
                {currentSession?.status || 'No session'}
              </span>
              <span>Executions: {executions.length}</span>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}
