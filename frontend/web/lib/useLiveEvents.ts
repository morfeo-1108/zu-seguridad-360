"use client";

export interface AgentEvent {
  id: string;
  agent_key: string;
  severity: string;
  action: string;
  created_at: string;
}

// Deliberately inert during the synthetic pilot: no EventSource, polling, or token in URL.
export function useLiveEvents(_token: string | null, _max = 50) {
  return { events: [] as AgentEvent[], connected: false };
}
