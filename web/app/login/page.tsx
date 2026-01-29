"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

import { LoginForm, LoginFormValues } from "@/components/ui/8bit/blocks/login-form-2";
import { getApiBaseUrl } from "@/lib/config";
import { useAuthStore } from "@/store/authStore";

export default function LoginPage() {
  const router = useRouter();
  const setSession = useAuthStore((state) => state.setSession);
  const token = useAuthStore((state) => state.token);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (token) {
      router.replace("/play");
    }
  }, [router, token]);

  const handleSubmit = async (values: LoginFormValues) => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await fetch(`${getApiBaseUrl()}/auth/login`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(values),
      });

      if (!response.ok) {
        const message = await response.text();
        throw new Error(message || "Failed to login");
      }

      const data = (await response.json()) as {
        token: string;
        user: { id: string; email: string; username: string };
      };

      setSession({ token: data.token, user: data.user });
      router.replace("/play");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unexpected error");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-gradient-to-b from-black via-slate-900 to-zinc-900 p-4 text-white">
      <LoginForm
        className="w-full max-w-sm"
        onSubmit={handleSubmit}
        isLoading={isLoading}
        error={error}
      />
    </div>
  );
}
