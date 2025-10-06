import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import CleanerApp from './CleanerApp';
import CleanerLoginPage from './CleanerLoginPage';

/**
 * Приложение уборщика с маршрутизацией
 * @returns {React.ReactNode} - Приложение уборщика
 */
const CleanerAppRouter = () => {
  return (
    <Router>
      <Routes>
        <Route path="/login" element={<CleanerLoginPage />} />
        <Route path="/*" element={<CleanerApp />} />
        <Route path="/" element={<Navigate to="/cleaner/login" replace />} />
      </Routes>
    </Router>
  );
};

export default CleanerAppRouter;

