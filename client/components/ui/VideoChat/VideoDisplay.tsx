"use client";

import { useEffect, useRef } from "react";

export default function VideoDisplay({
  videoSource,
}: {
  videoSource: MediaStream | undefined;
}) {
  const videoRef = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    if (videoRef.current && videoSource) {
      videoRef.current.srcObject = videoSource;
    }
  }, [videoSource]);

  return (
    <video
      ref={videoRef}
      autoPlay
      playsInline
      muted
      className="w-full h-full object-cover bg-zinc-200 rounded-lg"
    />
  );
}
