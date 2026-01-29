const trimTrailingSlash = (value?: string) => {
  if (!value) {
    return undefined;
  }
  return value.replace(/\/$/, "");
};

export const getApiBaseUrl = () => {
  const envUrl = trimTrailingSlash(process.env.NEXT_PUBLIC_API_URL);
  if (envUrl) {
    return envUrl;
  }

  if (typeof window === "undefined") {
    return "";
  }

  const protocol = window.location.protocol === "https:" ? "https" : "http";
  return `${protocol}://${window.location.hostname}:8080`;
};

export const getWsBaseUrl = () => {
  const envUrl = trimTrailingSlash(process.env.NEXT_PUBLIC_WS_URL);
  if (envUrl) {
    return envUrl;
  }

  if (typeof window === "undefined") {
    return "";
  }

  const protocol = window.location.protocol === "https:" ? "wss" : "ws";
  return `${protocol}://${window.location.hostname}:8080`;
};
