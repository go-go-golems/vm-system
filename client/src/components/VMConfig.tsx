import { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Settings, Package, Layers } from 'lucide-react';
import { BUILTIN_MODULES, BUILTIN_LIBRARIES, type VMProfile } from '@/lib/vmService';

interface VMConfigProps {
  vm: VMProfile;
  onUpdate: (vm: VMProfile) => void;
}

export function VMConfig({ vm, onUpdate }: VMConfigProps) {
  const [selectedModules, setSelectedModules] = useState<Set<string>>(
    new Set(vm.exposedModules)
  );
  const [selectedLibraries, setSelectedLibraries] = useState<Set<string>>(
    new Set(vm.libraries)
  );

  const handleModuleToggle = (moduleId: string) => {
    const newModules = new Set(selectedModules);
    if (newModules.has(moduleId)) {
      newModules.delete(moduleId);
    } else {
      newModules.add(moduleId);
    }
    setSelectedModules(newModules);
    onUpdate({
      ...vm,
      exposedModules: Array.from(newModules),
    });
  };

  const handleLibraryToggle = (libraryId: string) => {
    const newLibraries = new Set(selectedLibraries);
    if (newLibraries.has(libraryId)) {
      newLibraries.delete(libraryId);
    } else {
      newLibraries.add(libraryId);
    }
    setSelectedLibraries(newLibraries);
    onUpdate({
      ...vm,
      libraries: Array.from(newLibraries),
    });
  };

  return (
    <div className="space-y-4">
      {/* VM Info */}
      <Card className="bg-slate-900 border-slate-800">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-slate-100">
            <Settings className="w-5 h-5" />
            VM Configuration
          </CardTitle>
          <CardDescription className="text-slate-400">
            Configure exposed modules and libraries for {vm.name}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <div className="text-slate-500">Engine</div>
              <div className="text-slate-200 font-mono">{vm.engine}</div>
            </div>
            <div>
              <div className="text-slate-500">Status</div>
              <Badge variant={vm.isActive ? 'default' : 'secondary'} className="mt-1">
                {vm.isActive ? 'Active' : 'Inactive'}
              </Badge>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Exposed Modules */}
      <Card className="bg-slate-900 border-slate-800">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-slate-100">
            <Layers className="w-5 h-5" />
            Exposed Modules
          </CardTitle>
          <CardDescription className="text-slate-400">
            Select which built-in modules are available in the VM runtime
          </CardDescription>
        </CardHeader>
        <CardContent>
          <ScrollArea className="h-[300px] pr-4">
            <div className="space-y-4">
              {BUILTIN_MODULES.map((module) => (
                <div
                  key={module.id}
                  className="flex items-start space-x-3 p-3 rounded-lg bg-slate-950 border border-slate-800 hover:border-slate-700 transition-colors"
                >
                  <Checkbox
                    id={`module-${module.id}`}
                    checked={selectedModules.has(module.id)}
                    onCheckedChange={() => handleModuleToggle(module.id)}
                    className="mt-1"
                  />
                  <div className="flex-1">
                    <Label
                      htmlFor={`module-${module.id}`}
                      className="text-slate-200 font-medium cursor-pointer"
                    >
                      {module.name}
                    </Label>
                    <p className="text-sm text-slate-400 mt-1">{module.description}</p>
                    <div className="flex flex-wrap gap-1 mt-2">
                      {module.functions.map((func) => (
                        <Badge
                          key={func}
                          variant="outline"
                          className="text-xs bg-slate-900 text-slate-400 border-slate-700"
                        >
                          {func}
                        </Badge>
                      ))}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </ScrollArea>
          <Separator className="my-4 bg-slate-800" />
          <div className="text-sm text-slate-400">
            {selectedModules.size} of {BUILTIN_MODULES.length} modules selected
          </div>
        </CardContent>
      </Card>

      {/* Libraries */}
      <Card className="bg-slate-900 border-slate-800">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-slate-100">
            <Package className="w-5 h-5" />
            Libraries
          </CardTitle>
          <CardDescription className="text-slate-400">
            Select external libraries to load into the VM
          </CardDescription>
        </CardHeader>
        <CardContent>
          <ScrollArea className="h-[300px] pr-4">
            <div className="space-y-4">
              {BUILTIN_LIBRARIES.map((library) => (
                <div
                  key={library.id}
                  className="flex items-start space-x-3 p-3 rounded-lg bg-slate-950 border border-slate-800 hover:border-slate-700 transition-colors"
                >
                  <Checkbox
                    id={`library-${library.id}`}
                    checked={selectedLibraries.has(library.id)}
                    onCheckedChange={() => handleLibraryToggle(library.id)}
                    className="mt-1"
                  />
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <Label
                        htmlFor={`library-${library.id}`}
                        className="text-slate-200 font-medium cursor-pointer"
                      >
                        {library.name}
                      </Label>
                      <Badge variant="secondary" className="text-xs">
                        v{library.version}
                      </Badge>
                    </div>
                    <p className="text-sm text-slate-400 mt-1">{library.description}</p>
                    <div className="flex items-center gap-2 mt-2">
                      <Badge
                        variant="outline"
                        className="text-xs bg-slate-900 text-slate-400 border-slate-700"
                      >
                        {library.type}
                      </Badge>
                      <span className="text-xs text-slate-500 font-mono">
                        global: {library.config.global}
                      </span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </ScrollArea>
          <Separator className="my-4 bg-slate-800" />
          <div className="text-sm text-slate-400">
            {selectedLibraries.size} of {BUILTIN_LIBRARIES.length} libraries selected
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
