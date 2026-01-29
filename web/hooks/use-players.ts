import { useEffect, useMemo, useState } from "react";

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
};

const getDefaultBaseUrl = () => {
  if (process.env.NEXT_PUBLIC_API_URL) {
    return process.env.NEXT_PUBLIC_API_URL;
  }

  if (typeof window === "undefined") {
    return "";
  }

  const protocol = window.location.protocol === "https:" ? "https" : "http";
  return `${protocol}://${window.location.hostname}:8080`;
};

export function usePlayers(options: UsePlayersOptions = {}) {
  const { pollMs = 2000, baseUrl, enabled = true } = options;
  const resolvedBaseUrl = useMemo(
    () => baseUrl ?? getDefaultBaseUrl(),
    [baseUrl]
  );
  const [players, setPlayers] = useState<ServerPlayer[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!enabled || !resolvedBaseUrl) {
      return;
    }

    let isCancelled = false;
    let timerId: ReturnType<typeof setInterval> | null = null;

    const fetchPlayers = async () => {
      try {
        const response = await fetch(`${resolvedBaseUrl}/players`, {
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
  }, [enabled, pollMs, resolvedBaseUrl]);

  return { players, isLoading, error };
}
