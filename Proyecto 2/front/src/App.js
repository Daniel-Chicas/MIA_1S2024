import React from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";

import Vista1 from "./views/Vista1";
import Vista2 from "./views/Vista2";
import Vista3 from "./views/Vista3";


function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<Vista1 />} />
        <Route path="/vista2" element={<Vista2 />} />
        <Route path="/vista3" element={<Vista3 />} />
      </Routes>
    </Router>
  );
}

export default App;
