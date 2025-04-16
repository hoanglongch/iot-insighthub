// app.js - Enhanced with WebSocket signaling integration, monitoring & telemetry, and modular WASM support.

// --- Monitoring & Telemetry Setup (from previous step) ---

if (typeof Sentry !== 'undefined') {
  Sentry.init({
      dsn: 'https://examplePublicKey@o0.ingest.sentry.io/0',
      release: 'iot-insighthub@1.0.0',
      environment: 'production'
  });
} else {
  console.warn("Sentry not loaded; front-end errors won't be reported to Sentry.");
}

async function sendTelemetry(eventType, eventData) {
  try {
      await fetch('/telemetry', {  // Your backend telemetry endpoint.
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
              type: eventType,
              data: eventData,
              timestamp: new Date().toISOString()
          })
      });
  } catch (err) {
      console.error('Failed to send telemetry:', err);
  }
}

window.addEventListener('error', (e) => {
  console.error("Global error captured:", e.error);
  if (typeof Sentry !== 'undefined') {
      Sentry.captureException(e.error);
  }
  sendTelemetry('error', { message: e.message, stack: e.error ? e.error.stack : 'no stack' });
});

// --- WASM Module Loading & Modular Integration (unchanged) ---

async function loadWasmModule() {
  if (!WebAssembly.instantiateStreaming) {
      console.warn("WebAssembly.instantiateStreaming not available; using arrayBuffer fallback.");
  }
  try {
      const response = await fetch('anomaly_detection.wasm');
      let wasmModule;
      if (WebAssembly.instantiateStreaming) {
          wasmModule = await WebAssembly.instantiateStreaming(response, {});
      } else {
          const buffer = await response.arrayBuffer();
          wasmModule = await WebAssembly.instantiate(buffer, {});
      }
      console.log("WASM module loaded successfully.");
      return wasmModule.instance;
  } catch (e) {
      console.error("Failed to load WASM module, activating fallback.", e);
      sendTelemetry('wasm_load_error', { error: e.toString() });
      return null;
  }
}

function fallbackDetectAnomaly(value) {
  const threshold = 75.0;
  return value > threshold;
}

async function initAnomalyDetection() {
  const wasmInstance = await loadWasmModule();
  if (wasmInstance && typeof window.DetectAnomaly === 'function' &&
      typeof window.BenchmarkAnomaly === 'function') {
      window.detectAnomaly = function(value) {
          try {
              return window.DetectAnomaly(value);
          } catch (e) {
              console.error("Error during WASM detection:", e);
              sendTelemetry('wasm_detection_error', { error: e.toString() });
              return fallbackDetectAnomaly(value);
          }
      };
      window.benchmarkAnomaly = function(iterations, value) {
          try {
              return window.BenchmarkAnomaly(iterations, value);
          } catch (e) {
              console.error("Error during WASM benchmark:", e);
              sendTelemetry('wasm_benchmark_error', { error: e.toString() });
              return 0;
          }
      };
      console.log("Using WASM anomaly detection.");
  } else {
      window.detectAnomaly = fallbackDetectAnomaly;
      window.benchmarkAnomaly = function() { return 0; };
      console.log("Fallback anomaly detection activated.");
  }
}

initAnomalyDetection();

// --- WebRTC Signaling Integration ---

// Connect to the signaling server.
const clientId = prompt("Enter your client ID:");  // In production, use a proper authentication mechanism.
const signalingSocket = new WebSocket(`ws://localhost:8081/ws?id=${clientId}`);

signalingSocket.onopen = () => {
  console.log("Connected to signaling server.");
};

signalingSocket.onmessage = (messageEvent) => {
  const data = JSON.parse(messageEvent.data);
  console.log("Received signaling message:", data);

  // Process the signaling message based on its type.
  switch (data.type) {
      case "offer":
          // Set remote description and create answer.
          peerConnection.setRemoteDescription(new RTCSessionDescription(data.payload))
              .then(() => peerConnection.createAnswer())
              .then(answer => peerConnection.setLocalDescription(answer))
              .then(() => {
                  // Send the answer back to the caller.
                  signalingSocket.send(JSON.stringify({
                      type: "answer",
                      from: clientId,
                      to: data.from,
                      payload: peerConnection.localDescription
                  }));
              })
              .catch(err => console.error("Error handling offer:", err));
          break;
      case "answer":
          // Set remote description upon receiving the answer.
          peerConnection.setRemoteDescription(new RTCSessionDescription(data.payload))
              .catch(err => console.error("Error setting remote description:", err));
          break;
      case "candidate":
          // Add the received ICE candidate.
          peerConnection.addIceCandidate(new RTCIceCandidate(data.payload))
              .catch(err => console.error("Error adding ICE candidate:", err));
          break;
      default:
          console.warn("Unknown signaling message type:", data.type);
  }
};

signalingSocket.onerror = (error) => {
  console.error("Signaling socket error:", error);
  sendTelemetry('signaling_error', { error: error.toString() });
};

// When ICE candidates are gathered, send them to the signaling server.
const configuration = { iceServers: [{ urls: 'stun:stun.l.google.com:19302' }] };
const peerConnection = new RTCPeerConnection(configuration);

peerConnection.onicecandidate = (event) => {
  if (event.candidate) {
      signalingSocket.send(JSON.stringify({
          type: "candidate",
          from: clientId,
          // In a full app, the target client ID would be determined by your app's logic.
          to: prompt("Enter target client ID to send ICE candidate:"), 
          payload: event.candidate
      }));
  }
};

// Example: When starting a call, create an offer and send it via the signaling server.
document.getElementById("startCallBtn")?.addEventListener("click", async () => {
  // Set your target client ID.
  const targetClient = prompt("Enter target client ID:");
  const offer = await peerConnection.createOffer();
  await peerConnection.setLocalDescription(offer);
  signalingSocket.send(JSON.stringify({
      type: "offer",
      from: clientId,
      to: targetClient,
      payload: offer
  }));
});

// --- Existing WebRTC Media Code ---

navigator.mediaDevices.getUserMedia({ video: true, audio: true })
  .then(stream => {
      document.getElementById('localVideo').srcObject = stream;
      stream.getTracks().forEach(track => peerConnection.addTrack(track, stream));
  })
  .catch(error => {
      console.error('Error accessing media:', error);
      sendTelemetry('media_error', { error: error.toString() });
  });

peerConnection.ontrack = (event) => {
  document.getElementById('remoteVideo').srcObject = event.streams[0];
};
