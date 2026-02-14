import { Button } from "@/components/ui/button";
import { AlertCircle, Home } from "lucide-react";
import { useLocation } from "wouter";

export default function NotFound() {
  const [, setLocation] = useLocation();

  return (
    <div className="flex flex-col items-center justify-center h-64 text-slate-500">
      <AlertCircle className="w-10 h-10 mb-3 opacity-40" />
      <h1 className="text-2xl font-bold text-slate-300 mb-1">404</h1>
      <p className="text-sm mb-4">Page not found.</p>
      <Button
        onClick={() => setLocation("/templates")}
        variant="outline"
        size="sm"
        className="bg-slate-900 border-slate-700 text-slate-300"
      >
        <Home className="w-4 h-4 mr-2" />
        Go to Templates
      </Button>
    </div>
  );
}
