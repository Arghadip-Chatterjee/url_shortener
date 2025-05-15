"use client";
import { useEffect, useState } from "react";
import { useRouter, useParams } from "next/navigation";
import ShortenForm from "../../shorten-form";
import Link from "next/link";

export default function UserDashboard() {
  const [urls, setUrls] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const router = useRouter();
  const params = useParams();
  const userId = params.user_id;

  console.log(userId); // Add this line t

  useEffect(() => {
    async function fetchUserUrls() {
      setLoading(true);
      setError("");
      try {
        const res = await fetch(`http://localhost:8080/api/user-urls/${userId}`);
        if (!res.ok) throw new Error("Failed to fetch URLs");
        const data = await res.json();
        setUrls(data.urls || []);
      } catch (err) {
        setError("Could not load URLs");
      } finally {
        setLoading(false);
      }
    }
    if (userId) fetchUserUrls();
  }, [userId]);

  if (!userId) return <div>User ID missing from URL.</div>;

  function handleLogout() {
    localStorage.removeItem("user");
    router.push("/");
  }

  return (
    <div className="max-w-2xl mx-auto p-4">
      <div className="flex justify-between items-center mb-4">
        <h1 className="text-2xl font-bold">Your Shortened URLs</h1>
        <button onClick={handleLogout} className="text-red-600 underline">Logout</button>
      </div>
      <ShortenForm
        user={userId}
        onShorten={(shortUrl, shortId, originalUrl, user) => {
          console.log("Shortened:", { shortUrl, shortId, originalUrl, user });
        }}
      />

      {loading && <div>Loading...</div>}
      {error && <div className="text-red-500">{error}</div>}
      {!loading && !error && (
        <ul className="space-y-2">
          {urls.length === 0 && <li>No URLs found.</li>}
          {urls.map((url: any) => (
            <li key={url._id} className="border p-2 rounded">
              <div>Short: <Link href={url.original_url} className="text-blue-600 underline" onClick={async (e) => {
                e.preventDefault();
                // Send analytics event to backend
                await fetch("http://localhost:8080/frontend-analytics", {
                  method: "POST",
                  headers: { "Content-Type": "application/json" },
                  body: JSON.stringify({
                    short_id: url._id,
                    events: [{
                      user_agent: typeof navigator !== "undefined" ? navigator.userAgent : "",
                      browser: typeof navigator !== "undefined" ? navigator.userAgent : "",
                      timestamp: new Date().toISOString(),
                    }],
                  }),
                });
                window.open(url.original_url, "_blank");
              }}>{typeof window !== "undefined" ? window.location.origin : ""}/{url._id}</Link></div>
              <div>Original: <Link href={url.original_url} className="text-gray-700 underline">{url.original_url}</Link></div>
              <AnalyticsSection shortId={url._id} />
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

function AnalyticsSection({ shortId }: { shortId: string }) {
  const [clicks, setClicks] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    async function fetchAnalytics() {
      setLoading(true);
      setError("");
      try {
        const res = await fetch(`http://localhost:8080/analytics/${shortId}`);
        if (!res.ok) throw new Error("Failed to fetch analytics");
        const data = await res.json();
        setClicks(data || []);
      } catch (err) {
        setError("Could not load analytics");
      } finally {
        setLoading(false);
      }
    }
    if (shortId) fetchAnalytics();
  }, [shortId]);

  if (loading) return <div className="text-sm text-gray-500">Loading analytics...</div>;
  if (error) return <div className="text-sm text-red-500">{error}</div>;
  return (
    <div className="mt-2">
      <div className="font-semibold">Clicks: {clicks.length}</div>
      {clicks.length > 0 && (
        <details className="mt-1">
          <summary className="cursor-pointer text-blue-700">View Click Details</summary>
          <ul className="text-xs mt-1 space-y-1">
            {clicks.map((click, idx) => (
              <li key={idx} className="border-b pb-1">
                <div>Date: {new Date(click.timestamp).toLocaleString()}</div>
                <div>IP: {click.ip}</div>
                <div>Country: {click.country}</div>
                <div>Browser: {click.browser}</div>
                <div>User Agent: {click.user_agent}</div>
              </li>
            ))}
          </ul>
        </details>
      )}
    </div>
  );
}