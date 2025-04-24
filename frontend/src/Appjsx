import React from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import LinkGenerator from "./pages/LinkGenerator";
import DeepTrackPage from "./pages/DeepTrackPage";

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<LinkGenerator />} />
        <Route path="/track/:id" element={<DeepTrackPage />} />
      </Routes>
    </Router>
  );
}

export default App;