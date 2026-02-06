import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { PRESET_EXAMPLES } from '@/lib/vmService';
import { FileCode } from 'lucide-react';

interface PresetSelectorProps {
  onSelect: (code: string) => void;
}

export function PresetSelector({ onSelect }: PresetSelectorProps) {
  return (
    <div className="flex items-center gap-2">
      <FileCode className="w-4 h-4 text-slate-400" />
      <span className="text-sm text-slate-400">Examples:</span>
      <Select onValueChange={(value) => {
        const preset = PRESET_EXAMPLES.find((p) => p.id === value);
        if (preset) {
          onSelect(preset.code);
        }
      }}>
        <SelectTrigger className="w-[200px] bg-slate-900 border-slate-700 text-slate-300">
          <SelectValue placeholder="Select an example" />
        </SelectTrigger>
        <SelectContent className="bg-slate-900 border-slate-700">
          {PRESET_EXAMPLES.map((preset) => (
            <SelectItem
              key={preset.id}
              value={preset.id}
              className="text-slate-300 focus:bg-slate-800 focus:text-slate-100"
            >
              <div>
                <div className="font-medium">{preset.name}</div>
                <div className="text-xs text-slate-500">{preset.description}</div>
              </div>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}
