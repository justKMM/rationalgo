import type { AppState } from "./types";

const defaultBase = "http://localhost:8080";

function apiBase(): string {
  return import.meta.env.VITE_API_URL ?? defaultBase;
}

export async function fetchState(): Promise<AppState> {
  const res = await fetch(`${apiBase()}/api/state`);
  if (!res.ok) {
    throw new Error(`API /api/state returned ${res.status}`);
  }
  return res.json() as Promise<AppState>;
}

export function isApiConfigured(): boolean {
  return import.meta.env.VITE_USE_API !== "false";
}
