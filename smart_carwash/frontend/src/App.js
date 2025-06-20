import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import './shared/styles/global.css';

// Импорт приложений
import HomePage from './apps/home/HomePage';
import TelegramApp from './apps/telegram/TelegramApp';
import AdminApp from './apps/admin/AdminApp';
import CashierApp from './apps/cashier/CashierApp';

/**
 * Главный компонент приложения, который маршрутизирует запросы между разными интерфейсами
 */
function App() {
  return (
    <Router>
      <Routes>
        {/* Главная страница */}
        <Route path="/" element={<HomePage />} />
        
        {/* Маршруты для Telegram Mini App */}
        <Route path="/telegram/*" element={<TelegramApp />} />
        
        {/* Маршруты для интерфейса администратора */}
        <Route path="/admin/*" element={<AdminApp />} />
        
        {/* Маршруты для интерфейса кассира */}
        <Route path="/cashier/*" element={<CashierApp />} />
        
        {/* Перенаправление неизвестных маршрутов на главную страницу */}
        <Route path="*" element={<Navigate to="/" />} />
      </Routes>
    </Router>
  );
}

export default App;
