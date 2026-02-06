import { Button } from '@/components/ui/button';
import { CodeEditor } from '@/components/CodeEditor';
import { ExecutionConsole } from '@/components/ExecutionConsole';
import { PresetSelector } from '@/components/PresetSelector';
import { VMInfo } from '@/components/VMInfo';
import { vmService, type Execution } from '@/lib/vmService';
import { BookOpen, Play, RotateCcw, Terminal } from 'lucide-react';
import { Link } from 'wouter';
import { useEffect, useState } from 'react';
import { toast } from 'sonner';

export default function Home() {
  const [code, setCode] = useState('// Write your JavaScript code here\nconsole.log("Hello, VM!");\n');
  const [executions, setExecutions] = useState<Execution[]>([]);
  const [isExecuting, setIsExecuting] = useState(false);

  const vm = vmService.getVMs()[0];
  const session = vmService.getCurrentSession();

  useEffect(() => {
    // Load recent executions on mount
    setExecutions(vmService.getRecentExecutions());
  }, []);

  const handleExecute = async () => {
    if (!code.trim()) {
      toast.error('Please enter some code to execute');
      return;
    }

    setIsExecuting(true);

    try {
      const execution = await vmService.executeREPL(code);
      setExecutions((prev) => [execution, ...prev].slice(0, 10));

      if (execution.status === 'ok') {
        toast.success('Execution completed successfully');
      } else if (execution.status === 'error') {
        toast.error('Execution failed with error');
      }
    } catch (error: any) {
      toast.error(error.message || 'Execution failed');
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
    toast.success('Example loaded');
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault();
      handleExecute();
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 flex flex-col" onKeyDown={handleKeyDown}>
      {/* Header */}
      <header className="border-b border-slate-800 bg-slate-900/50 backdrop-blur">
        <div className="container py-4">
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

      {/* Main content */}
      <main className="flex-1 container py-6">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 h-[calc(100vh-180px)]">
          {/* Left column: Editor */}
          <div className="lg:col-span-2 flex flex-col gap-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold text-slate-200">Code Editor</h2>
              <div className="text-xs text-slate-500 font-mono">
                Press <kbd className="px-2 py-1 bg-slate-800 rounded">⌘/Ctrl + Enter</kbd> to run
              </div>
            </div>
            <CodeEditor
              value={code}
              onChange={setCode}
              placeholder="// Write your JavaScript code here..."
              className="flex-1"
            />
          </div>

          {/* Right column: Console and VM info */}
          <div className="flex flex-col gap-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold text-slate-200">Output</h2>
              <Button
                variant="outline"
                size="sm"
                onClick={handleClear}
                className="bg-slate-900 border-slate-700 text-slate-400 hover:bg-slate-800 hover:text-slate-300"
              >
                <RotateCcw className="w-3 h-3 mr-2" />
                Clear
              </Button>
            </div>
            <ExecutionConsole executions={executions} className="flex-1 min-h-[300px]" />

            <VMInfo vm={vm} session={session} />
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t border-slate-800 bg-slate-900/50 backdrop-blur">
        <div className="container py-4">
          <div className="flex items-center justify-between text-sm text-slate-500">
            <div>
              Built with <span className="text-blue-500">goja</span> VM system
            </div>
            <div className="flex items-center gap-4">
              <span>Session: {session?.id.slice(0, 8)}...</span>
              <span>•</span>
              <span>Executions: {executions.length}</span>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}
