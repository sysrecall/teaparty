"use client";

import { useState } from "react";
import TextChat from "./TextChat";
import VideoChat from "./VideoChat";

const MIN_WIDTH = 800;
const MIN_HEIGHT = 600;

const CONSTRAINTS = {
  // audio: { echoCancellation: true },
  video: {
    width: { min: MIN_WIDTH },
    height: { min: MIN_HEIGHT },
  },
};

const ICE_SERVERS = [{ urls: "stun:stun.l.google.com:19302" }];

async function openCamera(constraints: MediaStreamConstraints | undefined) {
  return await navigator.mediaDevices.getUserMedia(constraints);
}

type Status = "idle" | "waiting" | "chatting";

export default function Chat() {
  const [socket, setSocket] = useState<WebSocket>();
  const [localStream, setLocalStream] = useState<MediaStream>();
  const [remoteStream, setRemoteStream] = useState<MediaStream>();
  const [peerConnection, setPeerConnnection] = useState<RTCPeerConnection>();
  const [status, setStatus] = useState<Status>("idle");

  async function connect() {
    const stream = await openCamera(CONSTRAINTS);
    setLocalStream(stream);
    setStatus("waiting");

    const pc = new RTCPeerConnection({ iceServers: ICE_SERVERS });
    setPeerConnnection(pc);

    // add streams to peer connection
    stream.getTracks().forEach((track) => pc.addTrack(track, stream));

    pc.ontrack = (event) => {
      const [rs] = event.streams;
      setRemoteStream(rs);
      setStatus("chatting");
    };

    const ws = new WebSocket("ws://localhost:8080/ws");

    ws.onopen = (event) => {
      console.log("Connected to the server");
    };

    ws.onmessage = async (event) => {
      console.log("Message from the server:", event.data);

      const data = JSON.parse(event.data);

      switch (data.type) {
        case "match":
          pc.onicecandidate = (event) => {
            if (event.candidate) {
              ws.send(
                JSON.stringify({
                  type: "candidate",
                  message: event.candidate,
                }),
              );
            }
          };

          // only send offer if this is the offerer
          if (data.message === "offerer") {
            const offer = await pc.createOffer();
            await pc.setLocalDescription(offer);

            ws.send(
              JSON.stringify({
                type: "offer",
                message: offer,
              }),
            );
          }

          break;

        case "offer":
          await pc.setRemoteDescription(
            new RTCSessionDescription(data.message),
          );
          const answer = await pc.createAnswer();
          await pc.setLocalDescription(answer);

          ws.send(
            JSON.stringify({
              type: "answer",
              message: answer,
            }),
          );

          break;

        case "answer":
          await pc.setRemoteDescription(
            new RTCSessionDescription(data.message),
          );

          break;

        case "candidate":
          await pc.addIceCandidate(new RTCIceCandidate(data.message));
          break;

        default:
          console.error("Invalid message type arrived from the server");
      }
    };

    ws.onerror = (error) => {
      console.error("Websocket Error:", error);
    };

    ws.onclose = (event) => {
      console.log("Disconnected from the server");
      setStatus("idle");
    };

    setSocket(ws);
  }

  return (
    <div className="flex w-full h-full">
      <div className="w-1/3 h-full">
        <VideoChat localStream={localStream} remoteStream={remoteStream} />
      </div>

      <div className="w-2/3 h-full">
        <TextChat connect={connect} />
      </div>
    </div>
  );
}
