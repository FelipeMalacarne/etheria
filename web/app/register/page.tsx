"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

import { LoginForm, LoginFormValues } from "@/components/ui/8bit/blocks/login-form-2";
import { Button } from "@/components/ui/button";
import { getApiBaseUrl } from "@/lib/config";
import { useAuthStore } from "@/store/authStore";

export default function RegisterPage() {
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
      if (!values.username) {
        throw new Error("Username is required");
      }

      const response = await fetch(`${getApiBaseUrl()}/auth/register`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          email: values.email,
          password: values.password,
          username: values.username,
        }),
      });

      if (!response.ok) {
        const message = await response.text();
        throw new Error(message || "Failed to register");
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
        showUsername
        title="Join the adventure"
        description="Create your RuneWeb account"
        submitLabel="Create account"
        footerLink={{
          href: "/login",
          label: "Sign in",
          prefix: "Already have an account?",
        }}
      />
    </div>
  );
}
