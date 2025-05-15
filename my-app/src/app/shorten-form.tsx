"use client";
import { useState } from "react";

export default function ShortenForm({
  onShorten,
  user,
}: {
  onShorten: (shortUrl: string, shortId: string, originalUrl: string, user: any) => void;
  user: any;
}) {
  const [url, setUrl] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const res = await fetch(`http://localhost:8080/shorten`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        // body: JSON.stringify({ url, user }),
        body: JSON.stringify({ url, user_id: user }), // assuming `user.id` is the user's ID
      });

      if (!res.ok) throw new Error("Failed to shorten URL");

      const data = await res.json();
      const shortId = data.short_url.split("/").pop();
      onShorten(data.short_url, shortId, url, user); // Pass user too
      setUrl("");
      alert("URL shortened successfully!");
    } catch (err: any) {
      setError(err.message || "Unknown error");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col gap-4 max-w-md mx-auto mt-8"
    >
      <input
        type="url"
        placeholder="Enter your URL"
        value={url}
        onChange={(e) => setUrl(e.target.value)}
        required
        className="border rounded px-3 py-2"
      />
      <button
        type="submit"
        disabled={loading}
        className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 disabled:opacity-50"
      >
        {loading ? "Shortening..." : "Shorten URL"}
      </button>
      {error && <div className="text-red-600">{error}</div>}
    </form>
  );
}
