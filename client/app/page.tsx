import Chat from "@/components/Chat";

export default function Home() {
  return (
    <div className="flex flex-col w-full items-center justify-center bg-zinc-50 font-sans ">
      <main className="flex items-center justify-between py-32 px-16 bg-white sm:items-start">
        <div className="w-full h-full">
          <Chat />
        </div>
      </main>
    </div>
  );
}
