import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import './shared/styles/global.css';

// Импорт приложений
import HomePage from './apps/home/HomePage';
import TelegramApp from './apps/telegram/TelegramApp';
import AdminApp from './apps/admin/AdminApp';
import CashierApp from './apps/cashier/CashierApp';
import AdminLoginPage from './apps/admin/AdminLoginPage';
import CashierLoginPage from './apps/cashier/CashierLoginPage';
import CleanerLoginPage from './apps/cleaner/CleanerLoginPage';
import CleanerApp from './apps/cleaner/CleanerApp';

// Импорт компонентов
import ProtectedRoute from './shared/components/ProtectedRoute';
import AuthDebug from './shared/components/AuthDebug';

/**
 * Главный компонент приложения, который маршрутизирует запросы между разными интерфейсами
 */
function App() {
  return (
    <Router>
      <AuthDebug />
      <Routes>
        {/* Главная страница */}
        <Route path="/" element={<HomePage />} />
        
        {/* Маршруты для Telegram Mini App */}
        <Route path="/telegram/*" element={<TelegramApp />} />
        
        {/* Маршруты для интерфейса администратора */}
        <Route path="/admin/login" element={<AdminLoginPage />} />
        <Route 
          path="/admin/*" 
          element={
            <ProtectedRoute requireAdmin={true}>
              <AdminApp />
            </ProtectedRoute>
          } 
        />
        
        {/* Маршруты для интерфейса кассира */}
        <Route path="/cashier/login" element={<CashierLoginPage />} />
        <Route 
          path="/cashier/*" 
          element={
            <ProtectedRoute requireAdmin={false}>
              <CashierApp />
            </ProtectedRoute>
          } 
        />
        
        {/* Маршруты для интерфейса уборщика */}
        <Route path="/cleaner/login" element={<CleanerLoginPage />} />
        <Route 
          path="/cleaner/*" 
          element={
            <ProtectedRoute requireAdmin={false}>
              <CleanerApp />
            </ProtectedRoute>
          } 
        />
        
        {/* Перенаправление неизвестных маршрутов на главную страницу */}
        <Route path="*" element={<Navigate to="/" />} />
      </Routes>
    </Router>
  );
}

export default App;
