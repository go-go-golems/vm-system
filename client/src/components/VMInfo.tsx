import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import type { VMProfile, VMSession } from '@/lib/vmService';
import { Activity, Cpu, Database, HardDrive } from 'lucide-react';

interface VMInfoProps {
  vm: VMProfile;
  session: VMSession | null;
}

export function VMInfo({ vm, session }: VMInfoProps) {
  return (
    <Card className="bg-slate-900 border-slate-800">
      <CardHeader className="pb-3">
        <CardTitle className="text-lg flex items-center gap-2">
          <Activity className="w-5 h-5 text-blue-500" />
          VM Status
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-400">Profile</span>
            <span className="text-sm font-medium text-slate-200">{vm.name}</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-400">Engine</span>
            <Badge variant="outline" className="bg-slate-800 border-slate-700 text-slate-300">
              {vm.engine}
            </Badge>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-400">Session</span>
            {session ? (
              <Badge
                variant="outline"
                className={
                  session.status === 'ready'
                    ? 'bg-emerald-950 border-emerald-800 text-emerald-400'
                    : 'bg-slate-800 border-slate-700 text-slate-300'
                }
              >
                {session.status}
              </Badge>
            ) : (
              <span className="text-sm text-slate-500">None</span>
            )}
          </div>
        </div>

        <div className="pt-3 border-t border-slate-800 space-y-2">
          <div className="text-xs font-medium text-slate-400 mb-2">Resource Limits</div>
          <div className="grid grid-cols-2 gap-2">
            <div className="flex items-center gap-2">
              <Cpu className="w-3 h-3 text-slate-500" />
              <span className="text-xs text-slate-400">CPU: {vm.settings.limits.cpu_ms}ms</span>
            </div>
            <div className="flex items-center gap-2">
              <HardDrive className="w-3 h-3 text-slate-500" />
              <span className="text-xs text-slate-400">Memory: {vm.settings.limits.mem_mb}MB</span>
            </div>
            <div className="flex items-center gap-2">
              <Database className="w-3 h-3 text-slate-500" />
              <span className="text-xs text-slate-400">
                Events: {vm.settings.limits.max_events.toLocaleString()}
              </span>
            </div>
            <div className="flex items-center gap-2">
              <Database className="w-3 h-3 text-slate-500" />
              <span className="text-xs text-slate-400">
                Output: {vm.settings.limits.max_output_kb}KB
              </span>
            </div>
          </div>
        </div>

        <div className="pt-3 border-t border-slate-800 space-y-2">
          <div className="text-xs font-medium text-slate-400 mb-2">Capabilities</div>
          <div className="flex flex-wrap gap-1">
            {vm.capabilities.map((cap) => (
              <Badge
                key={cap.id}
                variant="outline"
                className={
                  cap.enabled
                    ? 'bg-blue-950 border-blue-800 text-blue-400 text-xs'
                    : 'bg-slate-800 border-slate-700 text-slate-500 text-xs'
                }
              >
                {cap.name}
              </Badge>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
