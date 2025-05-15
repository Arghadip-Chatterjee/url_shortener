"use client";
import { useState } from "react";

export function SignupForm({ onSignup }: { onSignup: (user: any) => void }) {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSignup(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const res = await fetch("http://localhost:8080/signup", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, email, password }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.message || "Signup failed");
      onSignup(data);
    } catch (err: any) {
      setError(err.message || "Unknown error");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSignup} className="flex flex-col gap-4 max-w-md mx-auto mt-8">
      <input type="text" placeholder="Name" value={name} onChange={e => setName(e.target.value)} required className="border rounded px-3 py-2" />
      <input type="email" placeholder="Email" value={email} onChange={e => setEmail(e.target.value)} required className="border rounded px-3 py-2" />
      <input type="password" placeholder="Password" value={password} onChange={e => setPassword(e.target.value)} required className="border rounded px-3 py-2" />
      <button type="submit" disabled={loading} className="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700 disabled:opacity-50">{loading ? "Signing up..." : "Sign Up"}</button>
      {error && <div className="text-red-600">{error}</div>}
    </form>
  );
}

export function LoginForm({ onLogin }: { onLogin: (user: any) => void }) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleLogin(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const res = await fetch("http://localhost:8080/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.message || "Login failed");
      onLogin(data);
    } catch (err: any) {
      setError(err.message || "Unknown error");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleLogin} className="flex flex-col gap-4 max-w-md mx-auto mt-8">
      <input type="email" placeholder="Email" value={email} onChange={e => setEmail(e.target.value)} required className="border rounded px-3 py-2" />
      <input type="password" placeholder="Password" value={password} onChange={e => setPassword(e.target.value)} required className="border rounded px-3 py-2" />
      <button type="submit" disabled={loading} className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 disabled:opacity-50">{loading ? "Logging in..." : "Login"}</button>
      {error && <div className="text-red-600">{error}</div>}
    </form>
  );
}