import { Badge } from '@/components/ui/badge';
import { vmService, type VMProfile, type VMSession } from '@/lib/vmService';
import { Box, Layers, Monitor, BookOpen, Terminal, ChevronRight } from 'lucide-react';
import { Link, useLocation } from 'wouter';
import { useEffect, useState, createContext, useContext, useCallback, type ReactNode } from 'react';

// ---------------------------------------------------------------------------
// Global app state context — shared across all pages
// ---------------------------------------------------------------------------

export interface AppState {
  templates: VMProfile[];
  sessions: VMSession[];
  currentSession: VMSession | null;
  initialized: boolean;
  refreshTemplates: () => Promise<void>;
  refreshSessions: () => Promise<void>;
  setCurrentSession: (session: VMSession | null) => void;
}

const AppStateContext = createContext<AppState>({
  templates: [],
  sessions: [],
  currentSession: null,
  initialized: false,
  refreshTemplates: async () => {},
  refreshSessions: async () => {},
  setCurrentSession: () => {},
});

export function useAppState() {
  return useContext(AppStateContext);
}

// ---------------------------------------------------------------------------
// Breadcrumb
// ---------------------------------------------------------------------------

interface BreadcrumbSegment {
  label: string;
  href?: string;
}

function Breadcrumbs({ segments }: { segments: BreadcrumbSegment[] }) {
  if (segments.length === 0) return null;
  return (
    <div className="flex items-center gap-1.5 text-sm text-slate-400 min-w-0">
      {segments.map((seg, i) => (
        <span key={i} className="flex items-center gap-1.5 min-w-0">
          {i > 0 && <ChevronRight className="w-3.5 h-3.5 flex-shrink-0 text-slate-600" />}
          {seg.href ? (
            <Link href={seg.href} className="hover:text-slate-200 transition-colors truncate">
              {seg.label}
            </Link>
          ) : (
            <span className="text-slate-200 truncate">{seg.label}</span>
          )}
        </span>
      ))}
    </div>
  );
}

// Compute breadcrumbs from path
function computeBreadcrumbs(
  path: string,
  templates: VMProfile[],
  sessions: VMSession[],
): BreadcrumbSegment[] {
  const segments: BreadcrumbSegment[] = [];

  if (path === '/' || path.startsWith('/templates')) {
    segments.push({ label: 'Templates', href: path === '/templates' ? undefined : '/templates' });

    const match = path.match(/^\/templates\/([^/]+)/);
    if (match) {
      const tpl = templates.find(t => t.id === match[1]);
      segments.push({ label: tpl?.name || match[1].slice(0, 8) + '…' });
    }
  } else if (path.startsWith('/sessions')) {
    segments.push({ label: 'Sessions', href: path === '/sessions' ? undefined : '/sessions' });

    const match = path.match(/^\/sessions\/([^/]+)/);
    if (match) {
      const sess = sessions.find(s => s.id === match[1]);
      segments.push({ label: sess?.name || match[1].slice(0, 8) + '…' });
    }
  } else if (path.startsWith('/system')) {
    segments.push({ label: 'System' });
  } else if (path.startsWith('/reference')) {
    segments.push({ label: 'Reference' });
  }

  return segments;
}

// ---------------------------------------------------------------------------
// Navigation
// ---------------------------------------------------------------------------

const NAV_ITEMS = [
  { href: '/templates', label: 'Templates', icon: Box },
  { href: '/sessions', label: 'Sessions', icon: Layers },
  { href: '/system', label: 'System', icon: Monitor },
  { href: '/reference', label: 'Reference', icon: BookOpen },
] as const;

function NavBar({ path }: { path: string }) {
  return (
    <nav className="flex items-center gap-1">
      {NAV_ITEMS.map(item => {
        const Icon = item.icon;
        const active = path === item.href || path.startsWith(item.href + '/');
        return (
          <Link
            key={item.href}
            href={item.href}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm transition-colors ${
              active
                ? 'bg-slate-800 text-slate-100'
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/50'
            }`}
          >
            <Icon className="w-4 h-4" />
            <span className="hidden sm:inline">{item.label}</span>
          </Link>
        );
      })}
    </nav>
  );
}

// ---------------------------------------------------------------------------
// Footer status bar
// ---------------------------------------------------------------------------

function FooterStatus({ currentSession, sessions }: { currentSession: VMSession | null; sessions: VMSession[] }) {
  const readyCount = sessions.filter(s => s.status === 'ready').length;

  return (
    <footer className="border-t border-slate-800 bg-slate-900/50 backdrop-blur">
      <div className="container py-2">
        <div className="flex items-center justify-between text-xs text-slate-500">
          <div className="flex items-center gap-3">
            {currentSession ? (
              <Link href={`/sessions/${currentSession.id}`} className="flex items-center gap-1.5 hover:text-slate-300 transition-colors">
                <span className={`w-1.5 h-1.5 rounded-full ${currentSession.status === 'ready' ? 'bg-emerald-500' : 'bg-slate-500'}`} />
                <span>Session: {currentSession.name}</span>
                <span className="text-slate-600">·</span>
                <span>{currentSession.status}</span>
              </Link>
            ) : (
              <span>No active session</span>
            )}
            {currentSession?.vm && (
              <>
                <span className="text-slate-700">·</span>
                <Link href={`/templates/${currentSession.vmId}`} className="hover:text-slate-300 transition-colors">
                  Template: {currentSession.vm.name}
                </Link>
              </>
            )}
          </div>
          <div className="flex items-center gap-3">
            <span>{readyCount} ready session{readyCount !== 1 ? 's' : ''}</span>
            <span className="text-slate-700">·</span>
            <span>goja VM</span>
          </div>
        </div>
      </div>
    </footer>
  );
}

// ---------------------------------------------------------------------------
// AppShell
// ---------------------------------------------------------------------------

export function AppShell({ children }: { children: ReactNode }) {
  const [location] = useLocation();
  const [templates, setTemplates] = useState<VMProfile[]>([]);
  const [sessions, setSessions] = useState<VMSession[]>([]);
  const [currentSession, setCurrentSessionState] = useState<VMSession | null>(null);
  const [initialized, setInitialized] = useState(false);

  const refreshTemplates = useCallback(async () => {
    try {
      await vmService.initialize();
      setTemplates(vmService.getVMs());
    } catch {
      // ignore — template list will stay empty
    }
  }, []);

  const refreshSessions = useCallback(async () => {
    try {
      const list = await vmService.listSessions();
      setSessions(list);
      const cur = vmService.getCurrentSession();
      setCurrentSessionState(cur);
    } catch {
      // ignore
    }
  }, []);

  const setCurrentSession = useCallback((session: VMSession | null) => {
    setCurrentSessionState(session);
  }, []);

  useEffect(() => {
    (async () => {
      await refreshTemplates();
      await refreshSessions();
      setInitialized(true);
    })();
  }, [refreshTemplates, refreshSessions]);

  const breadcrumbs = computeBreadcrumbs(location, templates, sessions);

  const appState: AppState = {
    templates,
    sessions,
    currentSession,
    initialized,
    refreshTemplates,
    refreshSessions,
    setCurrentSession,
  };

  return (
    <AppStateContext.Provider value={appState}>
      <div className="min-h-screen flex flex-col bg-slate-950">
        {/* Header */}
        <header className="border-b border-slate-800 bg-slate-900/50 backdrop-blur sticky top-0 z-10">
          <div className="container py-2.5">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <Link href="/templates" className="flex items-center gap-2 flex-shrink-0">
                  <div className="w-7 h-7 rounded-md bg-blue-600 flex items-center justify-center">
                    <Terminal className="w-4 h-4 text-white" />
                  </div>
                  <span className="text-sm font-semibold text-slate-100 hidden sm:inline">VM System</span>
                </Link>
                <NavBar path={location} />
              </div>
            </div>
          </div>
        </header>

        {/* Breadcrumb bar */}
        {breadcrumbs.length > 0 && (
          <div className="border-b border-slate-800/50 bg-slate-950">
            <div className="container py-2">
              <Breadcrumbs segments={breadcrumbs} />
            </div>
          </div>
        )}

        {/* Page content */}
        <main className="flex-1 container py-4">
          {children}
        </main>

        {/* Footer */}
        <FooterStatus currentSession={currentSession} sessions={sessions} />
      </div>
    </AppStateContext.Provider>
  );
}
