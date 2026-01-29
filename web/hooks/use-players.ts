import { useEffect, useMemo, useState } from "react";

import { getApiBaseUrl } from "@/lib/config";

type ServerPlayer = {
  id: string;
  x: number;
  y: number;
};

type PlayersResponse = {
  players: ServerPlayer[];
};

type UsePlayersOptions = {
  pollMs?: number;
  baseUrl?: string;
  enabled?: boolean;
  token?: string | null;
};

export function usePlayers(options: UsePlayersOptions = {}) {
  const { pollMs = 2000, baseUrl, enabled = true, token } = options;
  const resolvedBaseUrl = useMemo(
    () => baseUrl ?? getApiBaseUrl(),
    [baseUrl]
  );
  const [players, setPlayers] = useState<ServerPlayer[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!enabled || !resolvedBaseUrl || !token) {
      setPlayers([]);
      setError(null);
      setIsLoading(false);
      return;
    }

    let isCancelled = false;
    let timerId: ReturnType<typeof setInterval> | null = null;

    const fetchPlayers = async () => {
      try {
        const url = new URL(`${resolvedBaseUrl}/players`);
        url.searchParams.set("token", token);
        const response = await fetch(url.toString(), {
          cache: "no-store",
        });
        if (!response.ok) {
          throw new Error(`Failed to load players (${response.status})`);
        }

        const data = (await response.json()) as PlayersResponse;
        if (!isCancelled) {
          setPlayers(data.players ?? []);
          setError(null);
        }
      } catch (err) {
        if (!isCancelled) {
          setError(err instanceof Error ? err.message : "Unknown error");
        }
      } finally {
        if (!isCancelled) {
          setIsLoading(false);
        }
      }
    };

    fetchPlayers();
    timerId = setInterval(fetchPlayers, pollMs);

    return () => {
      isCancelled = true;
      if (timerId) {
        clearInterval(timerId);
      }
    };
  }, [enabled, pollMs, resolvedBaseUrl, token]);

  return { players, isLoading, error };
}
