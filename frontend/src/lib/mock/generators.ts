const B32 = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567";

// Mulberry32 seeded PRNG — deterministic across SSR/client.
function makeRng(seed: number) {
  let s = seed >>> 0;
  return () => {
    s = (s + 0x6D2B79F5) >>> 0;
    let t = s;
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

let _seedCounter = 1;
export function seededRng(seed?: number) {
  return makeRng(seed ?? _seedCounter++);
}

export function algoTx(rng: () => number = Math.random): string {
  let s = "";
  for (let i = 0; i < 52; i++) s += B32[Math.floor(rng() * 32)];
  return s;
}

export function shortId(rng: () => number = Math.random, prefix = "DEC"): string {
  let part = "";
  const alpha = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789";
  for (let i = 0; i < 6; i++) part += alpha[Math.floor(rng() * alpha.length)];
  return `${prefix}-${part}`;
}

export function nowIso(offsetMs = 0): string {
  return new Date(Date.now() + offsetMs).toISOString();
}
