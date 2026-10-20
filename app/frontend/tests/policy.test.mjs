import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import vm from "node:vm";
import ts from "typescript";

// Run the real policy and boot modules with bridge and host boundaries stubbed.
async function setup() {
  const context = vm.createContext({ console });
  const manifest = { id: "cached-extension", permissions: ["assets:read"] };
  const state = {
    cached: { mode: "disabled", allowed: [] },
    live: { mode: "disabled", allowed: [] },
    executions: 0,
    plugins: [],
  };
  const result = async (value) => {
    if (value instanceof Error) throw value;
    return await value;
  };
  const bridge = {
    CachedPluginPolicy: () => result(state.cached),
    GetPluginPolicy: () => result(state.live),
    PluginConsents: async () => ({ [manifest.id]: manifest.permissions }),
    SetPluginConsent: async () => {},
    PluginDecisions: async () => ({ [manifest.id]: true }),
    CachedVaultPlugins: async () => [{ manifest, source: "cached code" }],
    ListVaultPlugins: async () => { throw new Error("offline listing"); },
    AppVersion: async () => "1.0.0",
  };
  const host = {
    listPlugins: () => state.plugins,
    parseVaultManifest: (m) => m,
    registerVaultPlugin: (m, source) => {
      if (!state.plugins.some((p) => p.manifest.id === m.id)) {
        state.plugins.push({ manifest: m, source, builtIn: false, enabled: false });
      }
    },
    unregisterVaultPlugin: (id) => { state.plugins = state.plugins.filter((p) => p.manifest.id !== id); },
    enablePlugin: async (id) => {
      state.plugins.find((p) => p.manifest.id === id).enabled = true;
      state.executions++;
    },
    disablePlugin: (id) => { state.plugins.find((p) => p.manifest.id === id).enabled = false; },
    registerBuiltIn: () => {}, loaderPreflight: async () => ({}), setAppVersion: () => {},
  };
  function mock(exports) {
    return new vm.SyntheticModule(Object.keys(exports), function () {
      for (const [name, value] of Object.entries(exports)) this.setExport(name, value);
    }, { context });
  }
  async function source(name) {
    const text = await readFile(new URL(`../src/plugins/${name}.ts`, import.meta.url), "utf8");
    return new vm.SourceTextModule(ts.transpileModule(text, {
      compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext },
    }).outputText, { context });
  }
  const policy = await source("policy");
  const bridgeModule = mock(bridge);
  await policy.link(() => bridgeModule);
  await policy.evaluate();
  const boot = await source("boot");
  await boot.link((specifier) => {
    if (specifier === "./policy") return policy;
    if (specifier === "./host") return mock(host);
    if (specifier.includes("wailsjs")) return bridgeModule;
    // Built-ins are registered only by bootExtensions, not this sync test.
    const names = {
      "publish-doctor": "publishDoctor", "adoption-widget": "adoptionWidget",
      "usage-trends-widget": "usageTrendsWidget", "leaderboard-widget": "leaderboardWidget",
      templates: "templates", importer: "importer", "skill-doctor": "skillDoctor",
    };
    const name = names[specifier.split("/").at(-1)];
    return mock({ default: class {}, [`${name}Manifest`]: {} });
  });
  await boot.evaluate();
  return { state, policy: policy.namespace, boot: boot.namespace };
}

test("initial state and reset block third-party code and built-ins", async () => {
  const { state, policy } = await setup();
  assert.ok(policy.policyBlocks("cached-extension", false));
  assert.ok(policy.policyBlocks("builtin", true));
  state.live = { mode: "open", allowed: [] };
  await policy.loadPolicyAndConsents();
  policy.resetPolicy();
  assert.ok(policy.policyBlocks("cached-extension", false));
});

for (const invalid of [new Error("bridge unavailable"), null, {}, { mode: "future" }, { mode: "open", allowed: "bad" }]) {
  test(`cached code and consent cannot bypass invalid policy: ${String(invalid)}`, async () => {
    const { state, boot } = await setup();
    state.cached = invalid;
    state.live = invalid;
    await boot.syncVaultExtensions();
    assert.equal(state.executions, 0);
  });
}

test("cached code waits for live permission and a new profile cannot reuse it", async () => {
  const { state, boot } = await setup();
  let resolveLive;
  state.live = new Promise((resolve) => { resolveLive = resolve; });
  const sync = boot.syncVaultExtensions();
  // Wait until the boot's cached enablement pass has completed.
  await new Promise((resolve) => setImmediate(resolve));
  assert.equal(state.plugins.length, 1);
  assert.equal(state.executions, 0);
  resolveLive({ mode: "open", allowed: [] });
  await sync;
  assert.equal(state.executions, 1);
  state.cached = new Error("new profile has no policy cache");
  state.live = new Error("new profile is offline");
  await boot.syncVaultExtensions();
  assert.equal(state.executions, 1);
  assert.equal(state.plugins[0].enabled, false);
});

test("valid cached allowlist permits only listed code and exempts built-ins", async () => {
  const { state, policy, boot } = await setup();
  state.cached = { mode: "allowlist", allowed: ["cached-extension"] };
  // Go returns the valid cached policy on retrieval failure.
  state.live = state.cached;
  await boot.syncVaultExtensions();
  assert.equal(state.executions, 1);
  assert.ok(policy.policyBlocks("unlisted", false));
  assert.equal(policy.policyBlocks("builtin", true), null);
});
