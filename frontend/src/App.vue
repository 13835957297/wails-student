<script setup>
import { ref, onMounted, onUnmounted } from "vue";
import { Quit } from "../wailsjs/runtime/runtime";
// ===== CONFIGURATION =====
// Set your target URL here
const TARGET_URL = "https://aigc0002.cb-ec.cn/";

// Path to the loading video (served from public/, copied to dist root)
const LOADING_VIDEO = "./loading.webm";

// Health check interval (ms) when URL is unhealthy
const RETRY_INTERVAL = 6000;
// =========================

const state = ref("loading"); // loading | ready | error
const errorMsg = ref("");
const videoFailed = ref(false); // fallback to CSS animation if video fails

let timer = null;

function onVideoError() {
  videoFailed.value = true;
}

async function checkHealth() {
  try {
    // Use Wails runtime binding to call Go backend
    const result = await window.go.main.App.CheckURLHealth(TARGET_URL);
    if (result === "ok") {
      state.value = "ready";
      clearInterval(timer);
    } else {
      state.value = "error";
      errorMsg.value = result;
    }
  } catch (e) {
    state.value = "error";
    errorMsg.value = e?.toString?.() || "Failed to reach backend";
  }
}

function onKeyDown(e) {
  if (e.ctrlKey && (e.key === "q" || e.key === "Q" || e.code === "KeyQ")) {
    e.preventDefault();
    e.stopImmediatePropagation();
    Quit();
  }
}

onMounted(async() => {
  // Start periodic health check
  // checkHealth();
  timer = setInterval(checkHealth, RETRY_INTERVAL);

  // Listen for Ctrl+Q to quit (capture phase to intercept before iframe)
  document.addEventListener("keydown", onKeyDown, true);
});

onUnmounted(() => {
  clearInterval(timer);
  document.removeEventListener("keydown", onKeyDown, true);
});
</script>

<template>
  <div class="kiosk-container">
    <!-- Loading Screen -->
    <!-- <div v-if="state === 'loading' || state === 'error'" class="loading-screen"> -->
    <div v-if="state === 'loading' || state === 'error'" class="loading-screen">
      <video
        class="loading-video"
        :src="LOADING_VIDEO"
        autoplay
        muted
        loop
        playsinline
        preload="auto"
        @error="onVideoError"
      ></video>
      <!-- <div v-if="videoFailed" class="loading-fallback">
        <div class="spinner"></div>
        <p class="loading-text">Loading...</p>
      </div>
      <div v-if="state === 'error'" class="error-overlay">
        <div class="error-box">
          <div class="error-icon">&#x26A0;</div>
          <p>Waiting for service...</p>
          <p class="error-detail">{{ errorMsg }}</p>
          <p class="retry-note">Retrying every {{ RETRY_INTERVAL / 1000 }}s...</p>
        </div>
      </div> -->
    </div>

    <!-- Target WebView (iframe) -->
    <div v-show="state === 'ready'" class="target-frame-wrapper">
      <iframe
        class="target-frame"
        :src="TARGET_URL"
        frameborder="0"
        allow="autoplay; camera; microphone; fullscreen"
        sandbox="allow-same-origin allow-scripts allow-popups allow-forms allow-modals allow-downloads"
      ></iframe>
    </div>
  </div>
</template>

<style scoped>
.kiosk-container {
  width: 100vw;
  height: 100vh;
  position: relative;
  /* background: #000; */
  overflow: hidden;
}

.loading-screen {
  position: absolute;
  inset: 0;
  z-index: 10;
}

.loading-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.loading-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
}

.spinner {
  width: 64px;
  height: 64px;
  border: 4px solid rgba(255, 255, 255, 0.2);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-text {
  color: #fff;
  font-family: "Helvetica Neue", Arial, sans-serif;
  font-size: 16px;
  margin-top: 20px;
  opacity: 0.7;
  letter-spacing: 2px;
}

.error-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 20;
}

.error-box {
  text-align: center;
  color: #fff;
  font-family: "Helvetica Neue", Arial, sans-serif;
  padding: 32px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(8px);
  max-width: 480px;
}

.error-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.error-box p {
  margin: 8px 0;
  font-size: 18px;
}

.error-detail {
  font-size: 13px !important;
  opacity: 0.6;
  word-break: break-all;
}

.retry-note {
  font-size: 13px !important;
  opacity: 0.5;
  margin-top: 16px !important;
}

.target-frame-wrapper {
  position: absolute;
  inset: 0;
  overflow: hidden;
  z-index: 30;
}

.target-frame {
  width: calc(100% + 17px);
  height: 100%;
  border: none;
}
</style>