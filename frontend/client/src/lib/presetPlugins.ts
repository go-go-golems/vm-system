import type { SharedDomainName } from "@runtime/redux-adapter/store";

// Preset plugins for the unified v1 runtime contract

export interface PluginDefinition {
  id: string;
  title: string;
  description: string;
  capabilities?: {
    readShared?: SharedDomainName[];
    writeShared?: SharedDomainName[];
    systemCommands?: string[];
  };
  code: string;
}

// Counter Plugin
export const counterPlugin: PluginDefinition = {
  id: "counter",
  title: "Counter",
  description: "Simple local counter with shared counter summary updates",
  capabilities: {
    readShared: ["counter-summary"],
    writeShared: ["counter-summary"],
  },
  code: `
definePlugin(({ ui }) => {
  return {
    id: "counter",
    title: "Counter",
    description: "Simple counter",
    initialState: { value: 0 },
    widgets: {
      counter: {
        render({ pluginState, globalState }) {
          const value = Number(pluginState?.value ?? 0);
          const sharedCounter = globalState?.shared?.["counter-summary"];
          const totalValue = Number(sharedCounter?.totalValue ?? 0);
          const instanceCount = Number(sharedCounter?.instanceCount ?? 0);

          return ui.panel([
            ui.text("Counter: " + value),
            ui.row([
              ui.badge("Shared total: " + totalValue),
              ui.badge("Instances: " + instanceCount),
            ]),
            ui.row([
              ui.button("Decrement", { onClick: { handler: "decrement" } }),
              ui.button("Reset", { onClick: { handler: "reset" }, variant: "destructive" }),
              ui.button("Increment", { onClick: { handler: "increment" } }),
            ]),
          ]);
        },
        handlers: {
          increment({ dispatchPluginAction, dispatchSharedAction, pluginState }) {
            const next = Number(pluginState?.value ?? 0) + 1;
            dispatchPluginAction("increment");
            dispatchSharedAction("counter-summary", "set-instance", { value: next });
          },
          decrement({ dispatchPluginAction, dispatchSharedAction, pluginState }) {
            const next = Number(pluginState?.value ?? 0) - 1;
            dispatchPluginAction("decrement");
            dispatchSharedAction("counter-summary", "set-instance", { value: next });
          },
          reset({ dispatchPluginAction, dispatchSharedAction }) {
            dispatchPluginAction("reset");
            dispatchSharedAction("counter-summary", "set-instance", { value: 0 });
          },
        },
      },
    },
  };
});
  `,
};

// Calculator Plugin
export const calculatorPlugin: PluginDefinition = {
  id: "calculator",
  title: "Simple Calculator",
  description: "A basic calculator with +, -, *, / operations",
  capabilities: {
    readShared: [],
    writeShared: [],
  },
  code: `
definePlugin(({ ui }) => {
  return {
    id: "calculator",
    title: "Calculator",
    description: "Basic arithmetic calculator",
    initialState: {
      display: "0",
      accumulator: 0,
      operation: null,
    },
    widgets: {
      display: {
        render({ pluginState }) {
          const display = String(pluginState?.display ?? "0");
          return ui.panel([
            ui.text("Display: " + display),
            ui.row([
              ui.button("7", { onClick: { handler: "digit", args: 7 } }),
              ui.button("8", { onClick: { handler: "digit", args: 8 } }),
              ui.button("9", { onClick: { handler: "digit", args: 9 } }),
              ui.button("/", { onClick: { handler: "operation", args: "/" } }),
            ]),
            ui.row([
              ui.button("4", { onClick: { handler: "digit", args: 4 } }),
              ui.button("5", { onClick: { handler: "digit", args: 5 } }),
              ui.button("6", { onClick: { handler: "digit", args: 6 } }),
              ui.button("*", { onClick: { handler: "operation", args: "*" } }),
            ]),
            ui.row([
              ui.button("1", { onClick: { handler: "digit", args: 1 } }),
              ui.button("2", { onClick: { handler: "digit", args: 2 } }),
              ui.button("3", { onClick: { handler: "digit", args: 3 } }),
              ui.button("-", { onClick: { handler: "operation", args: "-" } }),
            ]),
            ui.row([
              ui.button("0", { onClick: { handler: "digit", args: 0 } }),
              ui.button("=", { onClick: { handler: "equals" } }),
              ui.button("C", { onClick: { handler: "clear" }, variant: "destructive" }),
              ui.button("+", { onClick: { handler: "operation", args: "+" } }),
            ]),
          ]);
        },
        handlers: {
          digit({ dispatchPluginAction }, digit) {
            dispatchPluginAction("digit", digit);
          },
          operation({ dispatchPluginAction }, op) {
            dispatchPluginAction("operation", op);
          },
          equals({ dispatchPluginAction }) {
            dispatchPluginAction("equals");
          },
          clear({ dispatchPluginAction }) {
            dispatchPluginAction("clear");
          },
        },
      },
    },
  };
});
  `,
};

// Status Dashboard Plugin
export const statusDashboardPlugin: PluginDefinition = {
  id: "status-dashboard",
  title: "Status Dashboard",
  description: "Shows unified runtime status and shared domain metrics",
  capabilities: {
    readShared: ["counter-summary", "runtime-metrics", "runtime-registry"],
    writeShared: [],
  },
  code: `
definePlugin(({ ui }) => {
  return {
    id: "status-dashboard",
    title: "Status Dashboard",
    description: "Runtime status dashboard",
    widgets: {
      status: {
        render({ globalState }) {
          const counterSummary = globalState?.shared?.["counter-summary"] ?? {};
          const runtimeMetrics = globalState?.shared?.["runtime-metrics"] ?? {};
          const pluginCount = Number(runtimeMetrics?.pluginCount ?? 0);
          const dispatchCount = Number(runtimeMetrics?.dispatchCount ?? 0);
          const counterValue = Number(counterSummary?.totalValue ?? 0);

          return ui.panel([
            ui.text("System Status"),
            ui.row([
              ui.badge("Plugins: " + pluginCount),
              ui.badge("Shared Counter: " + counterValue),
              ui.badge("Dispatches: " + dispatchCount),
            ]),
            ui.table(
              [
                ["Plugin Count", String(pluginCount)],
                ["Shared Counter Total", String(counterValue)],
                ["Dispatch Count", String(dispatchCount)],
              ],
              { headers: ["Metric", "Value"] }
            ),
          ]);
        },
        handlers: {},
      },
    },
  };
});
  `,
};

// Greeter Plugin
export const greeterPlugin: PluginDefinition = {
  id: "greeter",
  title: "Interactive Greeter",
  description: "Simple local state demo with input handling",
  capabilities: {
    readShared: ["greeter-profile"],
    writeShared: ["greeter-profile"],
  },
  code: `
definePlugin(({ ui }) => {
  return {
    id: "greeter",
    title: "Greeter",
    description: "Simple greeter",
    initialState: { name: "" },
    widgets: {
      greeter: {
        render({ pluginState }) {
          const name = String(pluginState?.name ?? "");
          const greeting = name ? "Hello, " + name + "!" : "Enter your name...";

          return ui.panel([
            ui.text(greeting),
            ui.input(name, {
              placeholder: "Your name",
              onChange: { handler: "updateName" },
            }),
          ]);
        },
        handlers: {
          updateName({ dispatchPluginAction, dispatchSharedAction }, args) {
            const name = args?.value ?? "";
            dispatchPluginAction("nameChanged", name);
            dispatchSharedAction("greeter-profile", "set-name", name);
          },
        },
      },
    },
  };
});
  `,
};

// Runtime Monitor Plugin
export const runtimeMonitorPlugin: PluginDefinition = {
  id: "runtime-monitor",
  title: "Runtime Monitor",
  description: "Shows loaded plugin registry from shared runtime state",
  capabilities: {
    readShared: ["runtime-registry"],
    writeShared: [],
  },
  code: `
definePlugin(({ ui }) => {
  return {
    id: "runtime-monitor",
    title: "Runtime Monitor",
    description: "Runtime monitor",
    widgets: {
      monitor: {
        render({ globalState }) {
          const plugins = Array.isArray(globalState?.shared?.["runtime-registry"])
            ? globalState.shared["runtime-registry"]
            : [];

          return ui.panel([
            ui.text("Plugin Registry"),
            ui.text("Total: " + plugins.length + " plugins"),
            ui.table(
              plugins.map((p) => [
                String(p.instanceId ?? p.id ?? ""),
                String(p.packageId ?? ""),
                String(p.status),
                p.enabled ? "YES" : "NO",
                String(p.widgets),
              ]),
              { headers: ["Instance", "Package", "Status", "Enabled", "Widgets"] }
            ),
          ]);
        },
        handlers: {},
      },
    },
  };
});
  `,
};

// Shared Greeter State Viewer Plugin
export const sharedGreeterStatePlugin: PluginDefinition = {
  id: "greeter-shared-state",
  title: "Greeter Shared State",
  description: "Shows greeter state from shared domain",
  capabilities: {
    readShared: ["greeter-profile"],
    writeShared: [],
  },
  code: `
definePlugin(({ ui }) => {
  return {
    id: "greeter-shared-state",
    title: "Greeter Shared State",
    description: "Shared greeter state viewer",
    widgets: {
      sharedGreeter: {
        render({ globalState }) {
          const greeterShared = globalState?.shared?.["greeter-profile"] ?? {};
          const name = String(greeterShared?.name ?? "");
          const greeting = name ? "Shared greeting: Hello, " + name + "!" : "Shared greeting: (empty)";

          return ui.panel([
            ui.text("Reads from globalState.shared['greeter-profile']"),
            ui.badge(name ? "SYNCED" : "NO NAME"),
            ui.text(greeting),
          ]);
        },
        handlers: {},
      },
    },
  };
});
  `,
};

// Column Starter Plugin
export const columnStarterPlugin: PluginDefinition = {
  id: "column-starter",
  title: "Column Starter",
  description: "Minimal plugin demonstrating ui.column vertical layout",
  capabilities: {
    readShared: [],
    writeShared: [],
  },
  code: `
definePlugin(({ ui }) => {
  return {
    id: "column-starter",
    title: "Column Starter",
    description: "ui.column starter demo",
    initialState: { clicks: 0 },
    widgets: {
      main: {
        render({ pluginState }) {
          const clicks = Number(pluginState?.clicks ?? 0);
          return ui.column([
            ui.text("Column layout demo"),
            ui.badge("Clicks: " + clicks),
            ui.button("Click", { onClick: { handler: "inc" } }),
          ]);
        },
        handlers: {
          inc({ dispatchPluginAction, pluginState }) {
            const next = Number(pluginState?.clicks ?? 0) + 1;
            dispatchPluginAction("state/merge", { clicks: next });
          },
        },
      },
    },
  };
});
  `,
};

// Task Checklist Plugin
export const checklistPlugin: PluginDefinition = {
  id: "checklist",
  title: "Checklist",
  description: "Small checklist UI built with ui.column + ui.row",
  capabilities: {
    readShared: [],
    writeShared: [],
  },
  code: `
definePlugin(({ ui }) => {
  return {
    id: "checklist",
    title: "Checklist",
    description: "Simple local checklist demo",
    initialState: {
      draft: "",
      items: [
        { id: "1", text: "Write plugin", done: true },
        { id: "2", text: "Use ui.column", done: false },
      ],
    },
    widgets: {
      main: {
        render({ pluginState }) {
          const items = Array.isArray(pluginState?.items) ? pluginState.items : [];
          const draft = String(pluginState?.draft ?? "");
          const doneCount = items.filter((i) => i?.done).length;

          return ui.column([
            ui.text("Checklist"),
            ui.badge("Done: " + doneCount + "/" + items.length),
            ui.row([
              ui.input(draft, {
                placeholder: "New task...",
                onChange: { handler: "setDraft" },
              }),
              ui.button("Add", { onClick: { handler: "add" } }),
            ]),
            items.length === 0
              ? ui.text("No items yet")
              : ui.table(
                  items.map((item) => [
                    item.done ? "DONE" : "TODO",
                    String(item.text ?? ""),
                    String(item.id ?? ""),
                  ]),
                  { headers: ["Status", "Task", "ID"] }
                ),
            ui.row([
              ui.button("Toggle First", { onClick: { handler: "toggleFirst" } }),
              ui.button("Clear", { onClick: { handler: "clear" }, variant: "destructive" }),
            ]),
          ]);
        },
        handlers: {
          setDraft({ dispatchPluginAction }, args) {
            dispatchPluginAction("state/merge", { draft: String(args?.value ?? "") });
          },
          add({ dispatchPluginAction, pluginState }) {
            const draft = String(pluginState?.draft ?? "").trim();
            if (!draft) return;
            const current = Array.isArray(pluginState?.items) ? pluginState.items : [];
            const next = current.concat([{ id: String(Date.now()), text: draft, done: false }]);
            dispatchPluginAction("state/replace", { draft: "", items: next });
          },
          toggleFirst({ dispatchPluginAction, pluginState }) {
            const current = Array.isArray(pluginState?.items) ? pluginState.items : [];
            if (current.length === 0) return;
            const next = current.slice();
            next[0] = {
              ...next[0],
              done: !Boolean(next[0]?.done),
            };
            dispatchPluginAction("state/merge", { items: next });
          },
          clear({ dispatchPluginAction }) {
            dispatchPluginAction("state/replace", { draft: "", items: [] });
          },
        },
      },
    },
  };
});
  `,
};

// Runtime Snapshot Plugin
export const runtimeSnapshotPlugin: PluginDefinition = {
  id: "runtime-snapshot",
  title: "Runtime Snapshot",
  description: "Read-only runtime telemetry laid out with ui.column",
  capabilities: {
    readShared: ["runtime-metrics", "runtime-registry"],
    writeShared: [],
  },
  code: `
definePlugin(({ ui }) => {
  return {
    id: "runtime-snapshot",
    title: "Runtime Snapshot",
    description: "Read-only runtime telemetry",
    widgets: {
      snapshot: {
        render({ globalState }) {
          const metrics = globalState?.shared?.["runtime-metrics"] ?? {};
          const registry = Array.isArray(globalState?.shared?.["runtime-registry"])
            ? globalState.shared["runtime-registry"]
            : [];

          return ui.column([
            ui.text("Runtime Snapshot"),
            ui.row([
              ui.badge("Plugins: " + Number(metrics?.pluginCount ?? 0)),
              ui.badge("Dispatches: " + Number(metrics?.dispatchCount ?? 0)),
            ]),
            ui.text("Recent Plugin Entries"),
            registry.length === 0
              ? ui.text("No plugins loaded")
              : ui.table(
                  registry.slice(0, 5).map((p) => [
                    String(p.instanceId ?? p.id ?? ""),
                    String(p.packageId ?? ""),
                    String(p.status ?? ""),
                  ]),
                  { headers: ["Instance", "Package", "Status"] }
                ),
          ]);
        },
        handlers: {},
      },
    },
  };
});
  `,
};

export const presetPlugins = [
  counterPlugin,
  calculatorPlugin,
  columnStarterPlugin,
  checklistPlugin,
  statusDashboardPlugin,
  greeterPlugin,
  sharedGreeterStatePlugin,
  runtimeMonitorPlugin,
  runtimeSnapshotPlugin,
];
