import { BrowserRouter, Routes, Route, Link } from "react-router-dom";
import { IndexPage } from "./pages/IndexPage";
import { ReplayPage } from "./pages/ReplayPage";

export default function App() {
  return (
    <BrowserRouter>
      <header className="bg-gray-900 text-white px-6 py-3">
        <Link to="/" className="font-semibold">replayreport</Link>
      </header>
      <main className="max-w-6xl mx-auto p-6">
        <Routes>
          <Route path="/" element={<IndexPage />} />
          <Route path="/replay/:matchID" element={<ReplayPage />} />
        </Routes>
      </main>
    </BrowserRouter>
  );
}
