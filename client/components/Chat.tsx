"use client";

import { useState } from "react";
import TextChat from "./TextChat";
import VideoChat from "./VideoChat";

const MIN_WIDTH = 800;
const MIN_HEIGHT = 600;

const constraints = {
  // audio: { echoCancellation: true },
  video: {
    width: { min: MIN_WIDTH },
    height: { min: MIN_HEIGHT },
  },
};

async function openCamera(constraints: MediaStreamConstraints | undefined) {
  return await navigator.mediaDevices.getUserMedia(constraints);
}

export default function Chat() {
  const [socket, setSocket] = useState<WebSocket>();
  const [localStream, setLocalStream] = useState<MediaStream>();

  async function connect() {
    const stream = await openCamera(constraints);
    setLocalStream(stream);

    const ws = new WebSocket("ws://localhost:8080/ws");

    ws.onopen = (event) => {
      console.log("Connected to the server");
    };

    ws.onmessage = (event) => {
      console.log("Message from the server:", event.data);
    };

    ws.onerror = (error) => {
      console.error("Websocket Error:", error);
    };

    ws.onclose = (event) => {
      console.log("Disconnected from the server");
    };

    setSocket(ws);
  }

  return (
    <div className="flex w-full h-full">
      <div className="w-1/3 h-full">
        <VideoChat localStream={localStream} strangerStream={undefined} />
      </div>

      <div className="w-2/3 h-full">
        <TextChat connect={connect} />
      </div>
    </div>
  );
}
