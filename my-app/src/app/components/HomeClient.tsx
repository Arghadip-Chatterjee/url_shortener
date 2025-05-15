// components/HomeClient.tsx
"use client";

import { useState, useEffect } from "react";
import { SignupForm, LoginForm } from "../auth";

export default function HomeClient() {
  const [user, setUser] = useState<any>(null);
  const [showLogin, setShowLogin] = useState(false);
  const [showSignup, setShowSignup] = useState(false);

  useEffect(() => {
    const stored = localStorage.getItem("user");
    if (stored) setUser(JSON.parse(stored));
  }, []);

  function handleLogin(data: any) {
    setUser(data);
    localStorage.setItem("user", JSON.stringify(data));
    setShowLogin(false);
    window.location.href = `/user/${data.user_id || data.id}`;
  }

  function handleSignup(data: any) {
    setUser(data);
    localStorage.setItem("user", JSON.stringify(data));
    setShowSignup(false);
    window.location.href = `/user/${data.user_id || data.id}`;
  }

  return (
    <div className="flex flex-col items-center w-full">
      <div className="mb-4 w-full flex flex-col items-center gap-2 text-white">
        <button onClick={() => setShowLogin(true)} className="underline p-4 m-3 bg-blue-400">
          Login
        </button>
        <button onClick={() => setShowSignup(true)} className="underline m-3 p-4 bg-green-400">
          Sign Up
        </button>
      </div>

      {showLogin && <LoginForm onLogin={handleLogin} />}
      {showSignup && <SignupForm onSignup={handleSignup} />}

      {!user && !showLogin && !showSignup && (
        <div className="text-gray-600 mt-4">Please login or sign up to shorten URLs.</div>
      )}
    </div>
  );
}
