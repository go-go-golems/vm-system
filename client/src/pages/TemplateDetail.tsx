import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useAppState } from '@/components/AppShell';
import { CreateSessionDialog } from '@/components/CreateSessionDialog';
import {
  vmService,
  BUILTIN_MODULES,
  BUILTIN_LIBRARIES,
  type VMProfile,
} from '@/lib/vmService';
import {
  ArrowLeft,
  Box,
  CheckCircle,
  Cpu,
  Database,
  FileCode,
  HardDrive,
  Layers,
  Loader2,
  Package,
  Play,
  Plus,
  Settings,
  XCircle,
} from 'lucide-react';
import { Link, useLocation, useParams } from 'wouter';
import { useCallback, useEffect, useState } from 'react';
import { toast } from 'sonner';

export default function TemplateDetail() {
  const params = useParams<{ id: string }>();
  const templateId = params.id;
  const { templates, sessions, refreshTemplates, refreshSessions } = useAppState();
  const [, setLocation] = useLocation();
  const [template, setTemplate] = useState<VMProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);

  // Module/library toggle state
  const [selectedModules, setSelectedModules] = useState<Set<string>>(new Set());
  const [selectedLibraries, setSelectedLibraries] = useState<Set<string>>(new Set());
  const [updatingModules, setUpdatingModules] = useState(false);
  const [updatingLibraries, setUpdatingLibraries] = useState(false);

  const loadTemplate = useCallback(async () => {
    try {
      const tpl = vmService.getVM(templateId);
      if (tpl) {
        setTemplate(tpl);
        setSelectedModules(new Set(tpl.exposedModules));
        setSelectedLibraries(new Set(tpl.libraries));
      } else {
        // try fetching
        await refreshTemplates();
        const fetched = vmService.getVM(templateId);
        if (fetched) {
          setTemplate(fetched);
          setSelectedModules(new Set(fetched.exposedModules));
          setSelectedLibraries(new Set(fetched.libraries));
        }
      }
    } finally {
      setLoading(false);
    }
  }, [templateId, refreshTemplates]);

  useEffect(() => {
    loadTemplate();
  }, [loadTemplate]);

  const derivedSessions = sessions.filter(s => s.vmId === templateId);

  const handleModuleToggle = async (moduleId: string) => {
    if (!template) return;
    const prev = new Set(selectedModules);
    const next = new Set(prev);
    if (next.has(moduleId)) next.delete(moduleId);
    else next.add(moduleId);
    setSelectedModules(next);
    setUpdatingModules(true);

    try {
      const updated = await vmService.updateTemplateModules(template.id, Array.from(next));
      setTemplate(updated);
      setSelectedModules(new Set(updated.exposedModules));
      await refreshTemplates();
      toast.success('Modules updated');
    } catch (error: any) {
      setSelectedModules(prev);
      toast.error('Failed to update modules', { description: error.message });
    } finally {
      setUpdatingModules(false);
    }
  };

  const handleLibraryToggle = async (libraryId: string) => {
    if (!template) return;
    const prev = new Set(selectedLibraries);
    const next = new Set(prev);
    if (next.has(libraryId)) next.delete(libraryId);
    else next.add(libraryId);
    setSelectedLibraries(next);
    setUpdatingLibraries(true);

    try {
      const updated = await vmService.updateTemplateLibraries(template.id, Array.from(next));
      setTemplate(updated);
      setSelectedLibraries(new Set(updated.libraries));
      await refreshTemplates();
      toast.success('Libraries updated');
      toast.info('Create a new session to pick up library changes.', { duration: 4000 });
    } catch (error: any) {
      setSelectedLibraries(prev);
      toast.error('Failed to update libraries', { description: error.message });
    } finally {
      setUpdatingLibraries(false);
    }
  };

  const handleSessionCreated = async (sessionId: string) => {
    await refreshSessions();
    setLocation(`/sessions/${sessionId}`);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64 text-slate-500">
        <Loader2 className="w-5 h-5 animate-spin mr-2" />
        Loading…
      </div>
    );
  }

  if (!template) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-slate-500">
        <Box className="w-10 h-10 mb-2 opacity-40" />
        <p className="text-sm">Template not found.</p>
        <Link href="/templates" className="text-blue-400 text-sm mt-2 hover:underline">
          Back to templates
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <Link href="/templates" className="text-slate-400 hover:text-slate-200 transition-colors">
              <ArrowLeft className="w-4 h-4" />
            </Link>
            <h1 className="text-lg font-semibold text-slate-100 truncate">{template.name}</h1>
          </div>
          <div className="flex items-center gap-3 text-xs text-slate-400">
            <span>Engine: <span className="text-slate-300">{template.engine}</span></span>
            <span className="text-slate-700">·</span>
            <span>{template.exposedModules.length} modules</span>
            <span className="text-slate-700">·</span>
            <span>{template.libraries.length} libraries</span>
            <span className="text-slate-700">·</span>
            <span>{derivedSessions.filter(s => s.status === 'ready').length} active sessions</span>
          </div>
        </div>
        <Button
          size="sm"
          className="bg-blue-600 hover:bg-blue-700 text-white flex-shrink-0"
          onClick={() => setCreateDialogOpen(true)}
        >
          <Play className="w-3.5 h-3.5 mr-1.5" />
          New Session
        </Button>
      </div>

      {/* Scope banner */}
      <div className="rounded-md bg-blue-950/30 border border-blue-800/40 px-3 py-2 text-xs text-blue-300">
        ⓘ Changes to this template apply to <strong>new sessions only</strong>. Running sessions keep their original configuration.
      </div>

      {/* Tabs */}
      <Tabs defaultValue="overview" className="space-y-4">
        <TabsList className="bg-slate-900 border border-slate-800">
          <TabsTrigger value="overview" className="data-[state=active]:bg-slate-800 text-xs">Overview</TabsTrigger>
          <TabsTrigger value="modules" className="data-[state=active]:bg-slate-800 text-xs">Modules</TabsTrigger>
          <TabsTrigger value="libraries" className="data-[state=active]:bg-slate-800 text-xs">Libraries</TabsTrigger>
          <TabsTrigger value="startup" className="data-[state=active]:bg-slate-800 text-xs">Startup Files</TabsTrigger>
          <TabsTrigger value="settings" className="data-[state=active]:bg-slate-800 text-xs">Settings</TabsTrigger>
        </TabsList>

        {/* Overview */}
        <TabsContent value="overview">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Identity */}
            <Card className="bg-slate-900 border-slate-800">
              <CardContent className="pt-5 space-y-3">
                <h3 className="text-sm font-medium text-slate-200 flex items-center gap-2">
                  <Box className="w-4 h-4 text-blue-500" />
                  Identity
                </h3>
                <div className="grid grid-cols-2 gap-y-2 text-xs">
                  <span className="text-slate-500">Name</span>
                  <span className="text-slate-200">{template.name}</span>
                  <span className="text-slate-500">Engine</span>
                  <span className="text-slate-200">{template.engine}</span>
                  <span className="text-slate-500">Status</span>
                  <Badge variant={template.isActive ? 'default' : 'secondary'} className="w-fit text-xs">
                    {template.isActive ? 'Active' : 'Inactive'}
                  </Badge>
                  <span className="text-slate-500">ID</span>
                  <span className="text-slate-400 font-mono text-xs">{template.id.slice(0, 16)}…</span>
                </div>
              </CardContent>
            </Card>

            {/* Resource limits */}
            <Card className="bg-slate-900 border-slate-800">
              <CardContent className="pt-5 space-y-3">
                <h3 className="text-sm font-medium text-slate-200 flex items-center gap-2">
                  <Settings className="w-4 h-4 text-blue-500" />
                  Resource Limits
                </h3>
                <div className="grid grid-cols-2 gap-2">
                  <div className="flex items-center gap-2 text-xs">
                    <Cpu className="w-3 h-3 text-slate-500" />
                    <span className="text-slate-400">CPU: <span className="text-slate-300">{template.settings.limits.cpu_ms}ms</span></span>
                  </div>
                  <div className="flex items-center gap-2 text-xs">
                    <HardDrive className="w-3 h-3 text-slate-500" />
                    <span className="text-slate-400">Memory: <span className="text-slate-300">{template.settings.limits.mem_mb}MB</span></span>
                  </div>
                  <div className="flex items-center gap-2 text-xs">
                    <Database className="w-3 h-3 text-slate-500" />
                    <span className="text-slate-400">Events: <span className="text-slate-300">{template.settings.limits.max_events.toLocaleString()}</span></span>
                  </div>
                  <div className="flex items-center gap-2 text-xs">
                    <Database className="w-3 h-3 text-slate-500" />
                    <span className="text-slate-400">Output: <span className="text-slate-300">{template.settings.limits.max_output_kb}KB</span></span>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Modules summary */}
            <Card className="bg-slate-900 border-slate-800">
              <CardContent className="pt-5 space-y-3">
                <h3 className="text-sm font-medium text-slate-200 flex items-center gap-2">
                  <Layers className="w-4 h-4 text-blue-500" />
                  Modules ({template.exposedModules.length})
                </h3>
                <div className="flex flex-wrap gap-1">
                  {template.exposedModules.length > 0 ? (
                    template.exposedModules.map(m => (
                      <Badge key={m} variant="outline" className="bg-blue-950/30 border-blue-800/40 text-blue-400 text-xs">
                        {m}
                      </Badge>
                    ))
                  ) : (
                    <span className="text-xs text-slate-500">No modules enabled</span>
                  )}
                </div>
              </CardContent>
            </Card>

            {/* Libraries summary */}
            <Card className="bg-slate-900 border-slate-800">
              <CardContent className="pt-5 space-y-3">
                <h3 className="text-sm font-medium text-slate-200 flex items-center gap-2">
                  <Package className="w-4 h-4 text-blue-500" />
                  Libraries ({template.libraries.length})
                </h3>
                <div className="flex flex-wrap gap-1">
                  {template.libraries.length > 0 ? (
                    template.libraries.map(l => (
                      <Badge key={l} variant="outline" className="bg-emerald-950/30 border-emerald-800/40 text-emerald-400 text-xs">
                        {l}
                      </Badge>
                    ))
                  ) : (
                    <span className="text-xs text-slate-500">No libraries loaded</span>
                  )}
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* Modules tab */}
        <TabsContent value="modules">
          <Card className="bg-slate-900 border-slate-800">
            <CardContent className="pt-5">
              <div className="flex items-center justify-between mb-3">
                <p className="text-xs text-slate-400">{selectedModules.size} of {BUILTIN_MODULES.length} modules enabled</p>
              </div>
              <div className="space-y-2">
                {BUILTIN_MODULES.map(mod => (
                  <div
                    key={mod.id}
                    className="flex items-start gap-3 p-3 rounded-md bg-slate-950 border border-slate-800 hover:border-slate-700 transition-colors"
                  >
                    <Checkbox
                      id={`mod-${mod.id}`}
                      checked={selectedModules.has(mod.id)}
                      onCheckedChange={() => handleModuleToggle(mod.id)}
                      disabled={updatingModules}
                      className="mt-0.5"
                    />
                    <div className="flex-1 min-w-0">
                      <Label htmlFor={`mod-${mod.id}`} className="text-slate-200 text-sm font-medium cursor-pointer">
                        {mod.name}
                      </Label>
                      <p className="text-xs text-slate-500 mt-0.5">{mod.description}</p>
                      <div className="flex flex-wrap gap-1 mt-1.5">
                        {mod.functions.map(f => (
                          <Badge key={f} variant="outline" className="text-[10px] bg-slate-900 text-slate-500 border-slate-700">
                            {f}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Libraries tab */}
        <TabsContent value="libraries">
          <Card className="bg-slate-900 border-slate-800">
            <CardContent className="pt-5">
              <div className="flex items-center justify-between mb-3">
                <p className="text-xs text-slate-400">{selectedLibraries.size} of {BUILTIN_LIBRARIES.length} libraries loaded</p>
              </div>
              <div className="space-y-2">
                {BUILTIN_LIBRARIES.map(lib => (
                  <div
                    key={lib.id}
                    className="flex items-start gap-3 p-3 rounded-md bg-slate-950 border border-slate-800 hover:border-slate-700 transition-colors"
                  >
                    <Checkbox
                      id={`lib-${lib.id}`}
                      checked={selectedLibraries.has(lib.id)}
                      onCheckedChange={() => handleLibraryToggle(lib.id)}
                      disabled={updatingLibraries}
                      className="mt-0.5"
                    />
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <Label htmlFor={`lib-${lib.id}`} className="text-slate-200 text-sm font-medium cursor-pointer">
                          {lib.name}
                        </Label>
                        <Badge variant="secondary" className="text-[10px]">v{lib.version}</Badge>
                      </div>
                      <p className="text-xs text-slate-500 mt-0.5">{lib.description}</p>
                      <div className="flex items-center gap-2 mt-1.5">
                        <Badge variant="outline" className="text-[10px] bg-slate-900 text-slate-500 border-slate-700">
                          {lib.type}
                        </Badge>
                        <span className="text-[10px] text-slate-600 font-mono">global: {lib.config.global}</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Startup Files tab */}
        <TabsContent value="startup">
          <Card className="bg-slate-900 border-slate-800">
            <CardContent className="pt-5">
              {template.startupFiles.length > 0 ? (
                <div className="space-y-2">
                  {template.startupFiles
                    .sort((a, b) => a.orderIndex - b.orderIndex)
                    .map((file, i) => (
                      <div key={file.id} className="flex items-center gap-3 p-3 rounded-md bg-slate-950 border border-slate-800">
                        <span className="text-xs text-slate-600 font-mono w-6">{i + 1}.</span>
                        <FileCode className="w-4 h-4 text-slate-500" />
                        <span className="text-sm text-slate-200 font-mono">{file.path}</span>
                        <Badge variant="outline" className="text-[10px] bg-slate-900 text-slate-500 border-slate-700 ml-auto">
                          {file.mode}
                        </Badge>
                      </div>
                    ))}
                </div>
              ) : (
                <p className="text-sm text-slate-500">No startup files configured.</p>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Settings tab */}
        <TabsContent value="settings">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Card className="bg-slate-900 border-slate-800">
              <CardContent className="pt-5 space-y-3">
                <h3 className="text-sm font-medium text-slate-200">Limits</h3>
                <div className="grid grid-cols-2 gap-y-2 text-xs">
                  <span className="text-slate-500">CPU time</span>
                  <span className="text-slate-300 font-mono">{template.settings.limits.cpu_ms}ms</span>
                  <span className="text-slate-500">Wall time</span>
                  <span className="text-slate-300 font-mono">{template.settings.limits.wall_ms}ms</span>
                  <span className="text-slate-500">Memory</span>
                  <span className="text-slate-300 font-mono">{template.settings.limits.mem_mb}MB</span>
                  <span className="text-slate-500">Max events</span>
                  <span className="text-slate-300 font-mono">{template.settings.limits.max_events.toLocaleString()}</span>
                  <span className="text-slate-500">Max output</span>
                  <span className="text-slate-300 font-mono">{template.settings.limits.max_output_kb}KB</span>
                </div>
              </CardContent>
            </Card>

            <Card className="bg-slate-900 border-slate-800">
              <CardContent className="pt-5 space-y-3">
                <h3 className="text-sm font-medium text-slate-200">Resolver</h3>
                <div className="space-y-2 text-xs">
                  <div>
                    <span className="text-slate-500">Roots: </span>
                    <span className="text-slate-300 font-mono">{template.settings.resolver.roots.join(', ')}</span>
                  </div>
                  <div>
                    <span className="text-slate-500">Extensions: </span>
                    <span className="text-slate-300 font-mono">{template.settings.resolver.extensions.join(', ')}</span>
                  </div>
                  <div>
                    <span className="text-slate-500">Absolute repo imports: </span>
                    <span className="text-slate-300">{template.settings.resolver.allow_absolute_repo_imports ? 'yes' : 'no'}</span>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className="bg-slate-900 border-slate-800">
              <CardContent className="pt-5 space-y-3">
                <h3 className="text-sm font-medium text-slate-200">Runtime</h3>
                <div className="grid grid-cols-2 gap-y-2 text-xs">
                  <span className="text-slate-500">ESM</span>
                  <span className="text-slate-300">{template.settings.runtime.esm ? 'enabled' : 'disabled'}</span>
                  <span className="text-slate-500">Strict mode</span>
                  <span className="text-slate-300">{template.settings.runtime.strict ? 'enabled' : 'disabled'}</span>
                  <span className="text-slate-500">Console API</span>
                  <span className="text-slate-300">{template.settings.runtime.console ? 'enabled' : 'disabled'}</span>
                </div>
              </CardContent>
            </Card>

            {/* Capabilities */}
            <Card className="bg-slate-900 border-slate-800">
              <CardContent className="pt-5 space-y-3">
                <h3 className="text-sm font-medium text-slate-200">Capabilities</h3>
                {template.capabilities.length > 0 ? (
                  <div className="space-y-1.5">
                    {template.capabilities.map(cap => (
                      <div key={cap.id} className="flex items-center gap-2 text-xs">
                        {cap.enabled ? (
                          <CheckCircle className="w-3 h-3 text-emerald-500" />
                        ) : (
                          <XCircle className="w-3 h-3 text-slate-600" />
                        )}
                        <span className={cap.enabled ? 'text-slate-300' : 'text-slate-500'}>{cap.name}</span>
                        <span className="text-slate-600">({cap.kind})</span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-xs text-slate-500">No capabilities configured.</p>
                )}
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>

      {/* Derived sessions */}
      <Separator className="bg-slate-800" />
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-medium text-slate-300">Sessions using this template</h2>
          <Button
            size="sm"
            variant="outline"
            className="bg-slate-900 border-slate-700 text-slate-300 h-7 text-xs"
            onClick={() => setCreateDialogOpen(true)}
          >
            <Plus className="w-3 h-3 mr-1" />
            New Session
          </Button>
        </div>
        {derivedSessions.length > 0 ? (
          <div className="space-y-1.5">
            {derivedSessions.map(s => (
              <Link
                key={s.id}
                href={`/sessions/${s.id}`}
                className="flex items-center justify-between p-2.5 rounded-md bg-slate-900 border border-slate-800 hover:border-slate-700 transition-colors"
              >
                <div className="flex items-center gap-2.5">
                  <span className={`w-1.5 h-1.5 rounded-full ${
                    s.status === 'ready' ? 'bg-emerald-500' :
                    s.status === 'starting' ? 'bg-blue-500 animate-pulse' :
                    s.status === 'crashed' ? 'bg-red-500' :
                    'bg-slate-500'
                  }`} />
                  <span className="text-sm text-slate-200">{s.name}</span>
                  <Badge variant="outline" className="text-[10px] bg-transparent border-slate-700 text-slate-500">
                    {s.status}
                  </Badge>
                </div>
                <span className="text-xs text-slate-500">{formatRelativeTime(s.createdAt)}</span>
              </Link>
            ))}
          </div>
        ) : (
          <p className="text-xs text-slate-500">No sessions created from this template yet.</p>
        )}
      </div>

      <CreateSessionDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        templates={templates}
        defaultTemplateId={templateId}
        onCreated={handleSessionCreated}
      />
    </div>
  );
}

function formatRelativeTime(date: Date): string {
  const now = Date.now();
  const diff = now - date.getTime();
  const seconds = Math.floor(diff / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);
  if (days > 0) return `${days}d ago`;
  if (hours > 0) return `${hours}h ago`;
  if (minutes > 0) return `${minutes}m ago`;
  return `${seconds}s ago`;
}
