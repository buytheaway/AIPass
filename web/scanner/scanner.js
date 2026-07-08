const video = document.getElementById("preview");
const startButton = document.getElementById("startButton");
const stopButton = document.getElementById("stopButton");
const cameraStatus = document.getElementById("cameraStatus");
const result = document.getElementById("result");
const manualForm = document.getElementById("manualForm");
const manualToken = document.getElementById("manualToken");

let stream = null;
let detector = null;
let scanning = false;
let lastToken = "";
let lastScanAt = 0;

const scannerId = localStorage.getItem("aipass_scanner_id") || "front-desk-1";
localStorage.setItem("aipass_scanner_id", scannerId);

function setResult(kind, title, detail) {
  result.className = `result ${kind}`;
  result.innerHTML = `<strong>${title}</strong><span>${detail}</span>`;
}

async function startCamera() {
  if (!("BarcodeDetector" in window)) {
    cameraStatus.textContent = "Manual fallback";
    setResult("neutral", "Camera QR detection unavailable", "Paste the token manually below.");
    return;
  }

  detector = new BarcodeDetector({ formats: ["qr_code"] });
  stream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: "environment" } });
  video.srcObject = stream;
  scanning = true;
  cameraStatus.textContent = "Camera active";
  scanLoop();
}

function stopCamera() {
  scanning = false;
  if (stream) {
    stream.getTracks().forEach((track) => track.stop());
    stream = null;
  }
  video.srcObject = null;
  cameraStatus.textContent = "Camera idle";
}

async function scanLoop() {
  if (!scanning || !detector) return;
  try {
    const codes = await detector.detect(video);
    if (codes.length > 0) {
      const token = codes[0].rawValue;
      const now = Date.now();
      if (token && (token !== lastToken || now - lastScanAt > 4000)) {
        lastToken = token;
        lastScanAt = now;
        await validateToken(token);
      }
    }
  } catch (error) {
    setResult("denied", "Scanner error", error.message);
  }
  requestAnimationFrame(scanLoop);
}

async function validateToken(token) {
  cameraStatus.textContent = "Validating";
  const response = await fetch("/api/v1/scans/validate", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ qr_token: token, scanner_id: scannerId }),
  });
  const data = await response.json();
  if (!response.ok) {
    setResult("denied", "Request failed", data.error || "Unknown error");
    cameraStatus.textContent = "Camera active";
    return;
  }
  if (data.decision === "allowed") {
    setResult("allowed", `${data.event_type.replace("_", " ")} allowed`, data.user ? data.user.full_name : "Member accepted");
  } else {
    setResult("denied", "Access denied", data.reason || "Validation failed");
  }
  cameraStatus.textContent = scanning ? "Camera active" : "Camera idle";
}

startButton.addEventListener("click", () => {
  startCamera().catch((error) => {
    cameraStatus.textContent = "Camera blocked";
    setResult("denied", "Cannot start camera", error.message);
  });
});

stopButton.addEventListener("click", stopCamera);

manualForm.addEventListener("submit", (event) => {
  event.preventDefault();
  const token = manualToken.value.trim();
  if (token) validateToken(token);
});

