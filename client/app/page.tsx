import Chat from "@/components/Chat";

export default function Home() {
  return (
    <div className="w-full h-screen bg-zinc-50">
      <main className="h-full p-2 md:p-8">
        <Chat />
      </main>
    </div>
  );
}
