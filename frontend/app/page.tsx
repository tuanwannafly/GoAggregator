"use client";

import * as React from "react";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Badge } from "@/components/ui/Badge";

type SearchType = "flights" | "hotels";

type FlightResult = {
  id: string;
  from: string;
  to: string;
  date: string;
  price: number;
  currency: string;
  airline: string;
  provider?: string;
};

type HotelResult = {
  id: string;
  city: string;
  checkin: string;
  checkout: string;
  name: string;
  price_per_night: number;
  currency: string;
  provider?: string;
};

type ProviderError = { provider: string; error: string };

type FlightResponse = {
  results: FlightResult[];
  meta: {
    providers_called: number;
    providers_succeeded: number;
    providers_failed: ProviderError[];
    cache_hit: boolean;
    duration_ms: number;
  };
};

type HotelResponse = {
  results: HotelResult[];
  meta: {
    providers_called: number;
    providers_succeeded: number;
    providers_failed: ProviderError[];
    cache_hit: boolean;
    duration_ms: number;
  };
};

type SearchResponse = FlightResponse | HotelResponse;

type ProviderStatus = {
  name: string;
  type: "flight" | "hotel";
  state: "closed" | "open" | "half-open";
};

// Provider chaos knobs — same ports as docker-compose.yml
const PROVIDERS = [
  { name: "provider-fast", port: 18080 },
  { name: "provider-slow", port: 18081 },
  { name: "provider-flaky", port: 18082 },
  { name: "provider-timeout", port: 18083 },
  { name: "provider-down", port: 18084 },
] as const;

const API_BASE = "http://localhost:8081";

function formatPrice(amount: number, currency: string) {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency,
    maximumFractionDigits: 0,
  }).format(amount);
}

function stateVariant(state: string) {
  if (state === "closed") return "success" as const;
  if (state === "half-open") return "warning" as const;
  return "error" as const;
}

export default function HomePage() {
  const [searchType, setSearchType] = React.useState<SearchType>("flights");

  // Flight fields
  const [from, setFrom] = React.useState("SGN");
  const [to, setTo] = React.useState("HAN");
  const [date, setDate] = React.useState("2026-08-15");

  // Hotel fields
  const [city, setCity] = React.useState("HAN");
  const [checkin, setCheckin] = React.useState("2026-08-15");
  const [checkout, setCheckout] = React.useState("2026-08-17");

  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);
  const [data, setData] = React.useState<SearchResponse | null>(null);

  const [providers, setProviders] = React.useState<ProviderStatus[]>([]);
  const [providersLoading, setProvidersLoading] = React.useState(false);
  const [lastStatusError, setLastStatusError] = React.useState(false);

  const loadProviders = React.useCallback(async () => {
    setProvidersLoading(true);
    try {
      const res = await fetch(`${API_BASE}/providers/status`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const json = await res.json();
      setProviders(json.providers ?? []);
      setLastStatusError(false);
    } catch {
      setProviders([]);
      setLastStatusError(true);
    } finally {
      setProvidersLoading(false);
    }
  }, []);

  React.useEffect(() => {
    loadProviders();
    const id = setInterval(loadProviders, 5000);
    return () => clearInterval(id);
  }, [loadProviders]);

  async function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setData(null);

    try {
      const params =
        searchType === "flights"
          ? new URLSearchParams({ from, to, date })
          : new URLSearchParams({ city, checkin, checkout });

      const res = await fetch(`${API_BASE}/search/${searchType}?${params}`);
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.error ?? `HTTP ${res.status}`);
      }
      const json: SearchResponse = await res.json();
      setData(json);
      loadProviders();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Request failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="min-h-screen bg-surface">
      <div className="container mx-auto py-8 flex flex-col gap-6 max-w-6xl">
        {/* HEADER */}
        <header className="flex flex-col gap-2">
          <div className="flex items-center gap-2">
            <Badge variant="primary">GoAggregator</Badge>
            <Badge>v0.1.0</Badge>
          </div>
          <h1 className="font-display text-[32px] sm:text-[40px] font-semibold leading-tight tracking-tight">
            Flight & Hotel Search
          </h1>
          <p className="text-on-muted text-[15px] leading-relaxed">
            Fan-out to 5 mock providers · circuit breaker · rate limiter · cache-aside
          </p>
        </header>

        {/* SEARCH */}
        <section className="bg-surface-elevated border border-border rounded-lg p-6 flex flex-col gap-4">
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => setSearchType("flights")}
              className={`px-4 py-2 rounded-md text-[14px] font-bold tracking-wide transition-colors ${
                searchType === "flights"
                  ? "bg-primary text-on-primary"
                  : "bg-surface-muted text-on-surface hover:border-border-strong"
              }`}
            >
              Flights
            </button>
            <button
              type="button"
              onClick={() => setSearchType("hotels")}
              className={`px-4 py-2 rounded-md text-[14px] font-bold tracking-wide transition-colors ${
                searchType === "hotels"
                  ? "bg-primary text-on-primary"
                  : "bg-surface-muted text-on-surface hover:border-border-strong"
              }`}
            >
              Hotels
            </button>
          </div>

          <form onSubmit={handleSearch} className="flex flex-col gap-4">
            {searchType === "flights" ? (
              <div className="grid gap-4 sm:grid-cols-3">
                <Field label="From">
                  <Input value={from} onChange={(e) => setFrom(e.target.value)} placeholder="SGN" required />
                </Field>
                <Field label="To">
                  <Input value={to} onChange={(e) => setTo(e.target.value)} placeholder="HAN" required />
                </Field>
                <Field label="Date">
                  <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} required />
                </Field>
              </div>
            ) : (
              <div className="grid gap-4 sm:grid-cols-3">
                <Field label="City">
                  <Input value={city} onChange={(e) => setCity(e.target.value)} placeholder="HAN" required />
                </Field>
                <Field label="Check-in">
                  <Input type="date" value={checkin} onChange={(e) => setCheckin(e.target.value)} required />
                </Field>
                <Field label="Check-out">
                  <Input type="date" value={checkout} onChange={(e) => setCheckout(e.target.value)} required />
                </Field>
              </div>
            )}

            <div className="flex flex-wrap items-center gap-4">
              <Button type="submit" disabled={loading}>
                {loading ? "Searching…" : "Search"}
              </Button>
              {data && (
                <MetaStrip
                  searchType={searchType}
                  meta={data.meta}
                  from={from}
                  to={to}
                  city={city}
                  count={data.results.length}
                />
              )}
            </div>

            {error && (
              <div className="bg-error/10 border border-error/30 text-error rounded-md px-4 py-3 text-[14px]">
                Error: {error}. Is the API running on {API_BASE}?
              </div>
            )}
          </form>
        </section>

        {/* PROVIDERS STATUS */}
        <section className="bg-surface-elevated border border-border rounded-lg p-6 flex flex-col gap-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <h2 className="font-display text-[18px] font-semibold">Providers</h2>
              <span className="text-on-muted text-[12px] uppercase tracking-wide font-semibold">
                Circuit breakers · auto-refresh 5s
              </span>
            </div>
            <button
              onClick={loadProviders}
              disabled={providersLoading}
              className="text-on-muted text-[12px] uppercase tracking-wide font-semibold hover:text-on-surface disabled:opacity-50"
            >
              {providersLoading ? "Refreshing…" : "Refresh"}
            </button>
          </div>

          {lastStatusError && providers.length === 0 ? (
            <p className="text-on-muted text-[13px]">
              Cannot reach API at {API_BASE}/providers/status
            </p>
          ) : (
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-5">
              {providers.map((p) => (
                <div
                  key={`${p.type}-${p.name}`}
                  className="border border-border rounded-md p-3 flex flex-col gap-2 bg-surface"
                >
                  <span className="font-mono text-[12px] font-semibold truncate">{p.name}</span>
                  <div className="flex items-center gap-1.5">
                    <Badge variant={stateVariant(p.state)}>{p.state}</Badge>
                    <span className="text-on-muted text-[10px] uppercase tracking-wide">
                      {p.type}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>

        {/* PROVIDER CONTROL (chaos engineering) */}
        <ProviderControl onAfterControl={loadProviders} />

        {/* RESULTS */}
        {data && (
          <section className="bg-surface-elevated border border-border rounded-lg p-6 flex flex-col gap-4">
            <div className="flex items-center gap-2 flex-wrap">
              <h2 className="font-display text-[18px] font-semibold">Results</h2>
              <span className="text-on-muted text-[12px] uppercase tracking-wide font-semibold">
                {searchType === "flights" ? `${from} → ${to}` : city}
              </span>
              {data.meta.cache_hit && <Badge variant="primary">cached</Badge>}
            </div>

            {data.results.length === 0 ? (
              <p className="text-on-muted text-[14px]">No results returned.</p>
            ) : searchType === "flights" ? (
              <FlightTable results={data.results as FlightResult[]} />
            ) : (
              <HotelTable results={data.results as HotelResult[]} />
            )}

            {data.meta.providers_failed.length > 0 && (
              <div className="border-t border-border pt-4 flex flex-col gap-1 text-[13px]">
                <span className="font-semibold text-on-surface">Provider failures</span>
                <ul className="flex flex-col gap-1">
                  {data.meta.providers_failed.map((f, i) => (
                    <li key={i} className="text-error">
                      <span className="font-mono">{f.provider}</span>: {f.error}
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </section>
        )}
      </div>
    </main>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="flex flex-col gap-1.5">
      <span className="text-on-muted text-[12px] font-semibold uppercase tracking-wide">{label}</span>
      {children}
    </label>
  );
}

function MetaStrip({
  searchType,
  meta,
  from,
  to,
  city,
  count,
}: {
  searchType: SearchType;
  meta: SearchResponse["meta"];
  from: string;
  to: string;
  city: string;
  count: number;
}) {
  return (
    <span className="text-on-muted text-[13px] flex items-center gap-2 flex-wrap">
      <span>
        {count} result{count === 1 ? "" : "s"}
      </span>
      <span>·</span>
      <span>
        {meta.providers_succeeded}/{meta.providers_called} providers
      </span>
      <span>·</span>
      <span>{meta.duration_ms}ms</span>
      {meta.cache_hit && (
        <>
          <span>·</span>
          <Badge variant="primary">cached</Badge>
        </>
      )}
    </span>
  );
}

function FlightTable({ results }: { results: FlightResult[] }) {
  const sorted = [...results].sort((a, b) => a.price - b.price);
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-[14px]">
        <thead className="text-on-muted text-[11px] uppercase tracking-wide">
          <tr className="border-b border-border">
            <th className="text-left py-2 pr-4 font-semibold">Provider</th>
            <th className="text-left py-2 pr-4 font-semibold">Airline</th>
            <th className="text-left py-2 pr-4 font-semibold">Route</th>
            <th className="text-left py-2 pr-4 font-semibold">Date</th>
            <th className="text-right py-2 font-semibold">Price</th>
          </tr>
        </thead>
        <tbody>
          {sorted.map((r) => (
            <tr key={`${r.provider}-${r.id}`} className="border-b border-border last:border-0">
              <td className="py-3 pr-4 font-mono text-[12px]">{r.provider || "—"}</td>
              <td className="py-3 pr-4">{r.airline}</td>
              <td className="py-3 pr-4 text-on-muted font-mono text-[12px]">
                {r.from} → {r.to}
              </td>
              <td className="py-3 pr-4 text-on-muted">{r.date}</td>
              <td className="py-3 text-right font-bold">{formatPrice(r.price, r.currency)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function HotelTable({ results }: { results: HotelResult[] }) {
  const sorted = [...results].sort((a, b) => a.price_per_night - b.price_per_night);
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-[14px]">
        <thead className="text-on-muted text-[11px] uppercase tracking-wide">
          <tr className="border-b border-border">
            <th className="text-left py-2 pr-4 font-semibold">Provider</th>
            <th className="text-left py-2 pr-4 font-semibold">Hotel</th>
            <th className="text-left py-2 pr-4 font-semibold">City</th>
            <th className="text-left py-2 pr-4 font-semibold">Stay</th>
            <th className="text-right py-2 font-semibold">Per night</th>
          </tr>
        </thead>
        <tbody>
          {sorted.map((r) => (
            <tr key={`${r.provider}-${r.id}`} className="border-b border-border last:border-0">
              <td className="py-3 pr-4 font-mono text-[12px]">{r.provider || "—"}</td>
              <td className="py-3 pr-4 font-semibold">{r.name}</td>
              <td className="py-3 pr-4 text-on-muted font-mono text-[12px]">{r.city}</td>
              <td className="py-3 pr-4 text-on-muted">
                {r.checkin} → {r.checkout}
              </td>
              <td className="py-3 text-right font-bold">
                {formatPrice(r.price_per_night, r.currency)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

// ---------- Provider Control (chaos engineering) ----------

type ControlForm = {
  latency_ms: number;
  error_rate: number;
};

type ControlState = Record<string, ControlForm>;

const PRESETS: Record<string, ControlForm> = {
  recover: { latency_ms: 0, error_rate: 0 },
  slow: { latency_ms: 2000, error_rate: 0 },
  kill: { latency_ms: 0, error_rate: 1 },
  chaos: { latency_ms: 1500, error_rate: 0.5 },
};

function ProviderControl({ onAfterControl }: { onAfterControl: () => void }) {
  const [controls, setControls] = React.useState<ControlState>({});
  const [busy, setBusy] = React.useState<string | null>(null);
  const [customLatency, setCustomLatency] = React.useState("1000");
  const [customError, setCustomError] = React.useState("0.5");
  const [feedback, setFeedback] = React.useState<string | null>(null);

  const loadAllControls = React.useCallback(async () => {
    const entries = await Promise.all(
      PROVIDERS.map(async (p) => {
        try {
          const res = await fetch(`http://localhost:${p.port}/control`);
          if (!res.ok) return [p.name, { latency_ms: 0, error_rate: 0 }] as const;
          const json = await res.json();
          return [
            p.name,
            {
              latency_ms: Number(json.latency_ms) || 0,
              error_rate: Number(json.error_rate) || 0,
            },
          ] as const;
        } catch {
          return [p.name, { latency_ms: 0, error_rate: 0 }] as const;
        }
      }),
    );
    setControls(Object.fromEntries(entries));
  }, []);

  React.useEffect(() => {
    loadAllControls();
  }, [loadAllControls]);

  async function applyControl(
    providerName: string,
    port: number,
    body: ControlForm,
  ) {
    setBusy(providerName);
    setFeedback(null);
    try {
      const res = await fetch(`http://localhost:${port}/control`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      if (!res.ok) {
        const errBody = await res.json().catch(() => ({}));
        throw new Error(errBody.error ?? `HTTP ${res.status}`);
      }
      setControls((prev) => ({ ...prev, [providerName]: body }));
      setFeedback(`${providerName}: latency=${body.latency_ms}ms error=${body.error_rate}`);
      onAfterControl();
    } catch (err) {
      setFeedback(
        `${providerName}: ${err instanceof Error ? err.message : "failed"}`,
      );
    } finally {
      setBusy(null);
    }
  }

  async function applyCustom(port: number) {
    const lat = parseInt(customLatency, 10);
    const err = parseFloat(customError);
    if (Number.isNaN(lat) || lat < 0) {
      setFeedback("latency must be >= 0");
      return;
    }
    if (Number.isNaN(err) || err < 0 || err > 1) {
      setFeedback("error_rate must be 0..1");
      return;
    }
    // apply to all providers at once
    await Promise.all(
      PROVIDERS.map((p) =>
        applyControl(p.name, p.port, { latency_ms: lat, error_rate: err }),
      ),
    );
  }

  return (
    <section className="bg-surface-elevated border border-border rounded-lg p-6 flex flex-col gap-4">
      <div className="flex items-center justify-between flex-wrap gap-2">
        <div className="flex items-center gap-2">
          <h2 className="font-display text-[18px] font-semibold">Chaos control</h2>
          <span className="text-on-muted text-[12px] uppercase tracking-wide font-semibold">
            POST /control on each mock provider
          </span>
        </div>
        <button
          onClick={loadAllControls}
          className="text-on-muted text-[12px] uppercase tracking-wide font-semibold hover:text-on-surface"
        >
          Reload state
        </button>
      </div>

      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-5">
        {PROVIDERS.map((p) => {
          const c = controls[p.name] ?? { latency_ms: 0, error_rate: 0 };
          const isBusy = busy === p.name;
          return (
            <div
              key={p.name}
              className="border border-border rounded-md p-3 flex flex-col gap-2 bg-surface"
            >
              <span className="font-mono text-[12px] font-semibold truncate">{p.name}</span>
              <span className="text-on-muted text-[11px] font-mono">
                {c.latency_ms}ms · {(c.error_rate * 100).toFixed(0)}% err
              </span>
              <div className="grid grid-cols-2 gap-1.5 mt-1">
                <button
                  onClick={() => applyControl(p.name, p.port, PRESETS.recover)}
                  disabled={isBusy}
                  className="text-[11px] font-semibold uppercase tracking-wide rounded-sm bg-success/20 text-success hover:bg-success/30 disabled:opacity-50 px-2 py-1"
                >
                  Recover
                </button>
                <button
                  onClick={() => applyControl(p.name, p.port, PRESETS.slow)}
                  disabled={isBusy}
                  className="text-[11px] font-semibold uppercase tracking-wide rounded-sm bg-warning/20 text-warning hover:bg-warning/30 disabled:opacity-50 px-2 py-1"
                >
                  Slow
                </button>
                <button
                  onClick={() => applyControl(p.name, p.port, PRESETS.kill)}
                  disabled={isBusy}
                  className="text-[11px] font-semibold uppercase tracking-wide rounded-sm bg-error/20 text-error hover:bg-error/30 disabled:opacity-50 px-2 py-1"
                >
                  Kill
                </button>
                <button
                  onClick={() => applyControl(p.name, p.port, PRESETS.chaos)}
                  disabled={isBusy}
                  className="text-[11px] font-semibold uppercase tracking-wide rounded-sm bg-secondary/20 text-secondary hover:bg-secondary/30 disabled:opacity-50 px-2 py-1"
                >
                  Chaos
                </button>
              </div>
            </div>
          );
        })}
      </div>

      <div className="border-t border-border pt-4 flex flex-col gap-3">
        <div className="text-on-muted text-[12px] uppercase tracking-wide font-semibold">
          Bulk apply to all providers
        </div>
        <div className="flex flex-wrap items-end gap-3">
          <Field label="Latency (ms)">
            <Input
              type="number"
              min="0"
              value={customLatency}
              onChange={(e) => setCustomLatency(e.target.value)}
              className="max-w-[140px]"
            />
          </Field>
          <Field label="Error rate (0–1)">
            <Input
              type="number"
              min="0"
              max="1"
              step="0.05"
              value={customError}
              onChange={(e) => setCustomError(e.target.value)}
              className="max-w-[140px]"
            />
          </Field>
          <Button
            variant="secondary"
            size="sm"
            onClick={() => {
              const customPort = PROVIDERS[0].port;
              void applyCustom(customPort);
            }}
            disabled={busy !== null}
          >
            Apply to all
          </Button>
          {feedback && (
            <span className="text-on-muted text-[12px] font-mono">{feedback}</span>
          )}
        </div>
      </div>
    </section>
  );
}
