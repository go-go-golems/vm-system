import { Toaster } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import NotFound from "@/pages/NotFound";
import { Redirect, Route, Switch } from "wouter";
import ErrorBoundary from "./components/ErrorBoundary";
import { AppShell } from "./components/AppShell";
import { ThemeProvider } from "./contexts/ThemeContext";
import Templates from "./pages/Templates";
import TemplateDetail from "./pages/TemplateDetail";
import Sessions from "./pages/Sessions";
import SessionDetail from "./pages/SessionDetail";
import System from "./pages/System";
import Reference from "./pages/Reference";

function Router() {
  return (
    <AppShell>
      <Switch>
        <Route path="/">
          <Redirect to="/templates" />
        </Route>
        <Route path="/templates" component={Templates} />
        <Route path="/templates/:id" component={TemplateDetail} />
        <Route path="/sessions" component={Sessions} />
        <Route path="/sessions/:id" component={SessionDetail} />
        <Route path="/system" component={System} />
        <Route path="/reference" component={Reference} />
        <Route path="/404" component={NotFound} />
        <Route component={NotFound} />
      </Switch>
    </AppShell>
  );
}

function App() {
  return (
    <ErrorBoundary>
      <ThemeProvider defaultTheme="dark">
        <TooltipProvider>
          <Toaster />
          <Router />
        </TooltipProvider>
      </ThemeProvider>
    </ErrorBoundary>
  );
}

export default App;
