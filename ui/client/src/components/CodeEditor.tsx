import { Textarea } from '@/components/ui/textarea';
import { cn } from '@/lib/utils';

interface CodeEditorProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
}

export function CodeEditor({ value, onChange, placeholder, className }: CodeEditorProps) {
  return (
    <div className={cn('relative h-full', className)}>
      <Textarea
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="font-mono text-sm h-full resize-none bg-slate-950 border-slate-800 text-slate-100 placeholder:text-slate-600 focus-visible:ring-blue-500"
        spellCheck={false}
      />
      <div className="absolute top-2 right-2 text-xs text-slate-600 font-mono pointer-events-none">
        JavaScript
      </div>
    </div>
  );
}
