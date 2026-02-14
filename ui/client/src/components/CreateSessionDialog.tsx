import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible';
import { useListTemplatesQuery, useCreateSessionMutation } from '@/lib/api';
import { setCurrentSessionId, setSessionName } from '@/lib/uiSlice';
import { DEFAULT_WORKSPACE_ID, DEFAULT_BASE_COMMIT_OID, DEFAULT_WORKTREE_PATH } from '@/lib/types';
import { ChevronRight, Loader2 } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useDispatch } from 'react-redux';
import { toast } from 'sonner';

interface CreateSessionDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** Pre-select a template */
  defaultTemplateId?: string;
  /** Called after successful creation with the new session ID */
  onCreated: (sessionId: string) => void;
}

export function CreateSessionDialog({
  open,
  onOpenChange,
  defaultTemplateId,
  onCreated,
}: CreateSessionDialogProps) {
  const { data: templates = [] } = useListTemplatesQuery();
  const [createSession] = useCreateSessionMutation();
  const dispatch = useDispatch();

  const [selectedTemplateId, setSelectedTemplateId] = useState(defaultTemplateId || '');
  const [sessionNameInput, setSessionNameInput] = useState('');
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const [creating, setCreating] = useState(false);

  // Reset form when dialog opens
  useEffect(() => {
    if (open) {
      setSelectedTemplateId(defaultTemplateId || templates[0]?.id || '');
      setSessionNameInput('');
      setAdvancedOpen(false);
    }
  }, [open, defaultTemplateId, templates]);

  const selectedTemplate = templates.find(t => t.id === selectedTemplateId);

  const handleCreate = async () => {
    if (!selectedTemplateId) {
      toast.error('Select a template');
      return;
    }
    setCreating(true);
    try {
      const session = await createSession({
        template_id: selectedTemplateId,
        workspace_id: DEFAULT_WORKSPACE_ID,
        base_commit_oid: DEFAULT_BASE_COMMIT_OID,
        worktree_path: DEFAULT_WORKTREE_PATH,
        name: sessionNameInput || undefined,
      }).unwrap();

      // Persist session name and set as current
      if (sessionNameInput.trim()) {
        dispatch(setSessionName({ sessionId: session.id, name: sessionNameInput }));
      }
      dispatch(setCurrentSessionId(session.id));

      toast.success('Session created', { description: session.name });
      onOpenChange(false);
      onCreated(session.id);
    } catch (error: any) {
      toast.error('Failed to create session', { description: error?.message || error?.data?.message || 'Unknown error' });
    } finally {
      setCreating(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="bg-slate-900 border-slate-700 sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="text-slate-100">Create Session</DialogTitle>
          <DialogDescription className="text-slate-400">
            Create a new runtime instance from a template.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* Template picker */}
          <div className="space-y-2">
            <Label className="text-slate-300">Template *</Label>
            <Select value={selectedTemplateId} onValueChange={setSelectedTemplateId}>
              <SelectTrigger className="bg-slate-950 border-slate-700 text-slate-200">
                <SelectValue placeholder="Select a template" />
              </SelectTrigger>
              <SelectContent className="bg-slate-900 border-slate-700">
                {templates.map(t => (
                  <SelectItem key={t.id} value={t.id} className="text-slate-200 focus:bg-slate-800">
                    {t.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Inherited config summary */}
          {selectedTemplate && (
            <div className="rounded-md bg-slate-950 border border-slate-800 p-3 text-xs text-slate-400 space-y-1">
              <div className="text-slate-500 font-medium mb-1">Inherited configuration</div>
              <div>Engine: <span className="text-slate-300">{selectedTemplate.engine}</span></div>
              <div>Modules: <span className="text-slate-300">{selectedTemplate.exposedModules.length > 0 ? selectedTemplate.exposedModules.join(', ') : 'none'}</span></div>
              <div>Libraries: <span className="text-slate-300">{selectedTemplate.libraries.length > 0 ? selectedTemplate.libraries.join(', ') : 'none'}</span></div>
              <div>
                Limits: <span className="text-slate-300">
                  cpu {selectedTemplate.settings.limits.cpu_ms}ms,
                  wall {selectedTemplate.settings.limits.wall_ms}ms,
                  mem {selectedTemplate.settings.limits.mem_mb}MB
                </span>
              </div>
            </div>
          )}

          {/* Session name */}
          <div className="space-y-2">
            <Label className="text-slate-300">Name (optional)</Label>
            <Input
              value={sessionNameInput}
              onChange={e => setSessionNameInput(e.target.value)}
              placeholder="e.g. Workshop Session 1"
              className="bg-slate-950 border-slate-700 text-slate-200"
              onKeyDown={e => {
                if (e.key === 'Enter') handleCreate();
              }}
            />
          </div>

          {/* Advanced */}
          <Collapsible open={advancedOpen} onOpenChange={setAdvancedOpen}>
            <CollapsibleTrigger className="flex items-center gap-1 text-xs text-slate-500 hover:text-slate-300 transition-colors">
              <ChevronRight className={`w-3.5 h-3.5 transition-transform ${advancedOpen ? 'rotate-90' : ''}`} />
              Advanced
            </CollapsibleTrigger>
            <CollapsibleContent className="mt-2 space-y-2 text-xs text-slate-500">
              <div className="rounded-md bg-slate-950 border border-slate-800 p-3 space-y-1.5">
                <div>workspace_id: <span className="text-slate-400 font-mono">{DEFAULT_WORKSPACE_ID}</span></div>
                <div>base_commit_oid: <span className="text-slate-400 font-mono">{DEFAULT_BASE_COMMIT_OID}</span></div>
                <div>worktree_path: <span className="text-slate-400 font-mono">{DEFAULT_WORKTREE_PATH}</span></div>
              </div>
              <p className="text-slate-600">These defaults come from environment variables.</p>
            </CollapsibleContent>
          </Collapsible>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            className="bg-slate-800 border-slate-700 text-slate-300"
          >
            Cancel
          </Button>
          <Button
            onClick={handleCreate}
            disabled={creating || !selectedTemplateId}
            className="bg-blue-600 hover:bg-blue-700 text-white"
          >
            {creating ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Creating…
              </>
            ) : (
              'Create Session'
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
