import type { AppState, Decision } from "./types";

const HEX = "0123456789abcdef";

export function mockHash(): string {
  let s = "0x";
  for (let i = 0; i < 64; i++) s += HEX[Math.floor(Math.random() * 16)];
  return s;
}

export function truncateHash(h: string): string {
  if (h.length < 14) return h;
  return `${h.slice(0, 8)}…${h.slice(-4)}`;
}

export function mockRound(base = 41_238_900): number {
  return base + Math.floor(Math.random() * 9999);
}

export function uid(): string {
  return Math.random().toString(36).slice(2, 10);
}

const NOW = Date.now();

const seedDecisions: Decision[] = [
  {
    id: uid(),
    vendor: "FuelPriceAPI",
    status: "APPROVED",
    amountEURQ: 0.15,
    intent: "Fuel index lookup BE/NL/DE for nightly fleet rebalancing",
    alternatives: [
      { name: "EIA-public", reason: "EU coverage missing; US-only feed" },
      { name: "cached @ 06:00", reason: "8h stale — diesel moved 2.1% intraday" },
    ],
    expectedValue: "+11% rebalancing margin",
    confidence: 0.74,
    policy: { budgetOk: true, reputation: 4.0, anomaly: "none", vendorAllowed: true },
    reasoningHash: mockHash(),
    round: mockRound(),
    timestamp: NOW - 1000 * 60 * 14,
    outcome: { predicted: "+11%", actual: "+9%", verdict: "Within band", trustDelta: -0.02 },
  },
  {
    id: uid(),
    vendor: "TollGuru",
    status: "APPROVED",
    amountEURQ: 0.04,
    intent: "Toll cost lookup BE→NL corridor, truck class N3",
    alternatives: [
      { name: "ViaMichelin", reason: "no N3 truck class endpoint" },
      { name: "internal-table", reason: "last updated 2024-11; rates revised Q1" },
    ],
    expectedValue: "+4% quote accuracy",
    confidence: 0.88,
    policy: { budgetOk: true, reputation: 4.5, anomaly: "none", vendorAllowed: true },
    reasoningHash: mockHash(),
    round: mockRound(),
    timestamp: NOW - 1000 * 60 * 9,
    outcome: { predicted: "+4%", actual: "+5%", verdict: "Good purchase", trustDelta: 0.05 },
  },
  {
    id: uid(),
    vendor: "OSRM-Pro",
    status: "APPROVED",
    amountEURQ: 0.08,
    intent: "Recompute route after traffic spike on A4 Antwerp→Breda",
    alternatives: [
      { name: "OSRM-public", reason: "no live-traffic layer" },
      { name: "Google Directions", reason: "1.2 EURQ per 1k req — 15× quote" },
    ],
    expectedValue: "+18% ETA accuracy",
    confidence: 0.83,
    policy: { budgetOk: true, reputation: 4.7, anomaly: "none", vendorAllowed: true },
    reasoningHash: mockHash(),
    round: mockRound(),
    timestamp: NOW - 1000 * 60 * 4,
    outcome: { predicted: "+18%", actual: "+21%", verdict: "Good purchase", trustDelta: 0.08 },
  },
  {
    id: uid(),
    vendor: "ScrapeShack",
    status: "BLOCKED",
    amountEURQ: 0.9,
    intent: "Competitor pricing scrape, rotating residential proxy",
    alternatives: [
      { name: "RateIntel", reason: "allowlisted but quote 3× higher" },
    ],
    expectedValue: "—",
    confidence: 0.41,
    policy: { budgetOk: true, reputation: 1.8, anomaly: "none", vendorAllowed: false },
    reasoningHash: mockHash(),
    round: mockRound(),
    timestamp: NOW - 1000 * 60 * 2,
    blockedReason: "Vendor not on allowlist; reputation below 2.5 threshold",
  },
];

export const initialState: AppState = {
  agent: "fleet-router-01",
  balance: 9.41,
  spent: 0.59,
  dailyLimit: 10.0,
  decisions: seedDecisions,
  vendors: [
    { name: "OSRM-Pro", score: 4.7 },
    { name: "TollGuru", score: 4.5 },
    { name: "WeatherAPI", score: 4.2 },
    { name: "FuelPriceAPI", score: 4.0 },
    { name: "MetricsHub.xyz", score: 1.1 },
  ],
  allowedVendors: ["OSRM-Pro", "TollGuru", "WeatherAPI", "FuelPriceAPI", "RateIntel"],
  blockedVendors: ["ScrapeShack", "MetricsHub.xyz"],
  alerts: [
    {
      id: uid(),
      level: "amber",
      message: "WeatherAPI price +12% vs 7d avg — within tolerance, monitoring",
      at: NOW - 1000 * 60 * 22,
    },
  ],
  selectedId: null,
};
