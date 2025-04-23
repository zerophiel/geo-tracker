import { useState } from "react";

export default function LinkGenerator() {
  const [decoyUrl, setDecoyUrl] = useState("");
  const [generatedLink, setGeneratedLink] = useState("");

  const handleGenerate = async () => {
    const res = await fetch("http://13.214.77.124:8080/api/generate", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ decoyUrl }),
    });
    const data = await res.json();
    setGeneratedLink(data.link);
  };

  return (
    <div style={{ padding: "2rem" }}>
      <h1>Tracking Link Generator</h1>
      <input
        type="text"
        placeholder="Enter decoy URL"
        value={decoyUrl}
        onChange={(e) => setDecoyUrl(e.target.value)}
      />
      <button onClick={handleGenerate}>Generate Link</button>
      {generatedLink && (
        <div>
          <p>Tracking Link:</p>
          <a href={generatedLink} target="_blank" rel="noopener noreferrer">
            {generatedLink}
          </a>
        </div>
      )}
    </div>
  );
}
