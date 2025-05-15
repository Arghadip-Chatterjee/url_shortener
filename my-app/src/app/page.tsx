// app/page.tsx (or pages/index.tsx depending on your setup)

import HomeClient from "@/app/components/HomeClient"; // Adjust path as needed

export default function Page() {
  return (
    <div className="m-auto justify-center items-center flex flex-col min-h-screen bg-amber-200">
        <h1 className="text-black ">URL Shortener</h1>
      <HomeClient />
    </div>
  );
}
