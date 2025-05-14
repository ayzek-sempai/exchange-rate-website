import React, { useState } from 'react';
import './App.css';

function App() {
  const [base, setBase] = useState('USD');
  const [target, setTarget] = useState('EUR');
  const [rate, setRate] = useState(null);

  const fetchRate = async () => {
    const res = await fetch(`/api/latest?base=${base}&target=${target}`);
    const data = await res.json();
    setRate(data);
  };

  return (
    <div className="app">
      <h1>Exchange Rate Checker</h1>
      <div className="controls">
        <input value={base} onChange={e => setBase(e.target.value.toUpperCase())} />
        <input value={target} onChange={e => setTarget(e.target.value.toUpperCase())} />
        <button onClick={fetchRate}>Check</button>
      </div>
      {rate && (
        <div className="result">
          <p>{rate.base} → {rate.target}</p>
          <p>Rate: {rate.rate}</p>
          <p>Time: {rate.timestamp}</p>
        </div>
      )}
    </div>
  );
}

export default App;
