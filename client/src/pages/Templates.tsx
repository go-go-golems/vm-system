import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { useAppState } from '@/components/AppShell';
import { CreateSessionDialog } from '@/components/CreateSessionDialog';
import { vmService } from '@/lib/vmService';
import { Box, Play, Plus, Loader2 } from 'lucide-react';
import { Link, useLocation } from 'wouter';
import { useState } from 'react';
import { toast } from 'sonner';

export default function Templates() {
  const { templates, sessions, initialized, refreshTemplates, refreshSessions } = useAppState();
  const [, setLocation] = useLocation();
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [createDialogTemplateId, setCreateDialogTemplateId] = useState<string | undefined>();
  const [creatingTemplate, setCreatingTemplate] = useState(false);

  const sessionCountByTemplate = (templateId: string) =>
    sessions.filter(s => s.vmId === templateId && s.status === 'ready').length;

  const handleCreateTemplate = async () => {
    setCreatingTemplate(true);
    try {
      const name = `New Template ${templates.length + 1}`;
      const baseUrl = ((import.meta as any).env.VITE_VM_SYSTEM_API_BASE_URL || '').replace(/\/$/, '');
      const res = await fetch(`${baseUrl}/api/v1/templates`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, engine: 'goja' }),
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      toast.success('Template created', { description: name });
      await refreshTemplates();
    } catch (error: any) {
      toast.error('Failed to create template', { description: error.message });
    } finally {
      setCreatingTemplate(false);
    }
  };

  const openCreateSession = (templateId: string) => {
    setCreateDialogTemplateId(templateId);
    setCreateDialogOpen(true);
  };

  const handleSessionCreated = async (sessionId: string) => {
    await refreshSessions();
    setLocation(`/sessions/${sessionId}`);
  };

  if (!initialized) {
    return (
      <div className="flex items-center justify-center h-64 text-slate-500">
        <Loader2 className="w-5 h-5 animate-spin mr-2" />
        Loading…
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold text-slate-100">Templates</h1>
          <p className="text-sm text-slate-400">Configuration blueprints for VM sessions.</p>
        </div>
        <Button
          size="sm"
          onClick={handleCreateTemplate}
          disabled={creatingTemplate}
          className="bg-blue-600 hover:bg-blue-700 text-white"
        >
          {creatingTemplate ? <Loader2 className="w-4 h-4 mr-1.5 animate-spin" /> : <Plus className="w-4 h-4 mr-1.5" />}
          New Template
        </Button>
      </div>

      {/* Table */}
      {templates.length === 0 ? (
        <div className="flex flex-col items-center justify-center h-48 text-slate-500">
          <Box className="w-10 h-10 mb-2 opacity-40" />
          <p className="text-sm">No templates yet.</p>
        </div>
      ) : (
        <div className="border border-slate-800 rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-slate-800 bg-slate-900/50">
                <th className="text-left px-4 py-2.5 text-slate-400 font-medium">Name</th>
                <th className="text-left px-4 py-2.5 text-slate-400 font-medium hidden sm:table-cell">Engine</th>
                <th className="text-center px-4 py-2.5 text-slate-400 font-medium hidden md:table-cell">Modules</th>
                <th className="text-center px-4 py-2.5 text-slate-400 font-medium hidden md:table-cell">Libraries</th>
                <th className="text-center px-4 py-2.5 text-slate-400 font-medium hidden lg:table-cell">Sessions</th>
                <th className="text-right px-4 py-2.5 text-slate-400 font-medium">Actions</th>
              </tr>
            </thead>
            <tbody>
              {templates.map(tpl => {
                const readySessions = sessionCountByTemplate(tpl.id);
                return (
                  <tr
                    key={tpl.id}
                    className="border-b border-slate-800/50 hover:bg-slate-900/30 transition-colors"
                  >
                    <td className="px-4 py-3">
                      <Link href={`/templates/${tpl.id}`} className="text-slate-200 hover:text-white font-medium transition-colors">
                        {tpl.name}
                      </Link>
                    </td>
                    <td className="px-4 py-3 hidden sm:table-cell">
                      <Badge variant="outline" className="bg-slate-900 border-slate-700 text-slate-400 text-xs">
                        {tpl.engine}
                      </Badge>
                    </td>
                    <td className="px-4 py-3 text-center text-slate-400 hidden md:table-cell">
                      {tpl.exposedModules.length}
                    </td>
                    <td className="px-4 py-3 text-center text-slate-400 hidden md:table-cell">
                      {tpl.libraries.length}
                    </td>
                    <td className="px-4 py-3 text-center hidden lg:table-cell">
                      {readySessions > 0 ? (
                        <Badge variant="outline" className="bg-emerald-950 border-emerald-800 text-emerald-400 text-xs">
                          {readySessions} ready
                        </Badge>
                      ) : (
                        <span className="text-slate-600">—</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex items-center justify-end gap-2">
                        <Link href={`/templates/${tpl.id}`}>
                          <Button size="sm" variant="outline" className="bg-slate-900 border-slate-700 text-slate-300 hover:bg-slate-800 h-7 px-2.5 text-xs">
                            Open
                          </Button>
                        </Link>
                        <Button
                          size="sm"
                          variant="outline"
                          className="bg-slate-900 border-slate-700 text-slate-300 hover:bg-slate-800 h-7 px-2.5 text-xs"
                          onClick={() => openCreateSession(tpl.id)}
                        >
                          <Play className="w-3 h-3 mr-1" />
                          Session
                        </Button>
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      <CreateSessionDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        templates={templates}
        defaultTemplateId={createDialogTemplateId}
        onCreated={handleSessionCreated}
      />
    </div>
  );
}
